//nolint:nilnil // A handler without result returns nil, which is the null result of JSON-RPC.
package server

import (
	"bytes"
	"context"
	"slices"

	golrfmt "github.com/backbone81/golr/internal/fmt"
	"github.com/backbone81/golr/internal/lsp/jsonrpc"
	"github.com/backbone81/golr/internal/lsp/model"
	"github.com/backbone81/golr/internal/lsp/protocol"
	golrparser "github.com/backbone81/golr/internal/parsergen/frontend/golr/parser"
)

// errNoNameToRename answers a rename at a position without a name.
var errNoNameToRename = NewError(int(protocol.LSPErrorCodesRequestFailed), "no name to rename at the position")

// The semantic token types and modifiers of the legend, by their index in it.
const (
	semanticTokenNonterminal = iota
	semanticTokenTerminal
)

const semanticTokenDeclaration = 1 << 0

// semanticTokensLegend is the legend of the semantic tokens: "class" for nonterminals and "enum" for terminals, as
// common themes color them without configuration, and "declaration" on definitions.
var semanticTokensLegend = protocol.SemanticTokensLegend{
	TokenTypes:     []string{string(protocol.SemanticTokenTypesClass), string(protocol.SemanticTokenTypesEnum)},
	TokenModifiers: []string{string(protocol.SemanticTokenModifiersDeclaration)},
}

// This file holds the application logic of the LSP operations. The handlers which change server state run on the
// reader. The others run in their own goroutine and must only read the documents they get.

// handleInitialize handles the initialize request, see
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#initialize
func (s *Server) handleInitialize(
	_ context.Context,
	_ *Documents,
	params *protocol.InitializeParams,
) (protocol.InitializeResult, error) {
	if s.lifecycle != LifecycleUninitialized {
		return protocol.InitializeResult{}, NewError(jsonrpc.ErrorCodeInvalidRequest, "server already initialized")
	}
	s.lifecycle = LifecycleRunning
	if params.ProcessId != nil {
		s.watcher.Watch(*params.ProcessId)
	}
	general, _ := params.Capabilities.General.Get()
	if slices.Contains(general.PositionEncodings, protocol.PositionEncodingKindUTF8) {
		s.encoding = protocol.PositionEncodingKindUTF8
	}

	serverInfo := protocol.ServerInfo{Name: "GoLR"}
	if s.version != "" {
		serverInfo.Version = protocol.NewOptional(s.version)
	}
	return protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			PositionEncoding: protocol.NewOptional(s.encoding),
			TextDocumentSync: protocol.NewServerCapabilitiesTextDocumentSync(protocol.TextDocumentSyncKindFull),
			CompletionProvider: protocol.NewOptional(protocol.CompletionOptions{
				TriggerCharacters: []string{"@"},
			}),
			DefinitionProvider: protocol.NewServerCapabilitiesDefinitionProvider(true),
			ReferencesProvider: protocol.NewServerCapabilitiesReferencesProvider(true),
			RenameProvider:     s.renameProvider(params.Capabilities),
			SemanticTokensProvider: protocol.NewServerCapabilitiesSemanticTokensProvider(protocol.SemanticTokensOptions{
				Legend: semanticTokensLegend,
				Full:   protocol.NewSemanticTokensOptionsFull(true),
			}),
			DocumentFormattingProvider: protocol.NewServerCapabilitiesDocumentFormattingProvider(true),
		},
		ServerInfo: protocol.NewOptional(serverInfo),
	}, nil
}

// renameProvider returns the rename capability. It announces prepareRename only to a client which supports it, as
// the specification allows RenameOptions only then.
func (s *Server) renameProvider(capabilities protocol.ClientCapabilities) protocol.ServerCapabilitiesRenameProvider {
	textDocument, _ := capabilities.TextDocument.Get()
	rename, _ := textDocument.Rename.Get()
	if prepareSupport, _ := rename.PrepareSupport.Get(); !prepareSupport {
		return protocol.NewServerCapabilitiesRenameProvider(true)
	}
	return protocol.NewServerCapabilitiesRenameProvider(protocol.RenameOptions{
		PrepareProvider: protocol.NewOptional(true),
	})
}

// handleInitialized handles the initialized notification, see
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#initialized
func (s *Server) handleInitialized(context.Context, *Documents, *protocol.InitializedParams) (any, error) {
	return nil, nil
}

// handleShutdown handles the shutdown request, see
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#shutdown
func (s *Server) handleShutdown(context.Context, *Documents, *struct{}) (any, error) {
	s.lifecycle = LifecycleShutDown
	return nil, nil
}

// handleExit ends the session on the exit notification, see
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#exit
func (s *Server) handleExit(context.Context, *Documents, *struct{}) (any, error) {
	return nil, errEndOfSession
}

// handleCancel cancels a pending request on the $/cancelRequest notification, see
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#cancelRequest
// A missing id is logged and ignored, as the request then only runs to its end. An unknown id is ignored, as the
// response may be on its way.
func (s *Server) handleCancel(_ context.Context, _ *Documents, params *protocol.CancelParams) (any, error) {
	var id jsonrpc.ID
	switch value := params.Id.Value().(type) {
	case int:
		id = jsonrpc.NewNumberID(value)
	case string:
		id = jsonrpc.NewTextID(value)
	default:
		s.logger.Error("$/cancelRequest without id.")
		return nil, nil
	}

	s.pendingMutex.Lock()
	defer s.pendingMutex.Unlock()

	if cancel, ok := s.pending[id.String()]; ok {
		cancel()
	}
	return nil, nil
}

// handleDidOpenTextDocument handles the textDocument/didOpen notification, see
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#textDocument_didOpen
func (s *Server) handleDidOpenTextDocument(
	_ context.Context,
	_ *Documents,
	params *protocol.DidOpenTextDocumentParams,
) (any, error) {
	document := params.TextDocument
	if document.Uri == "" {
		s.logger.Error("Notification without uri.", "method", protocol.MethodTextDocumentDidOpen)
		return nil, nil
	}
	s.documents = s.documents.with(newSnapshot(document.Uri, document.Version, document.Text, s.encoding))
	return nil, nil
}

// handleDidChangeTextDocument handles the textDocument/didChange notification, see
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#textDocument_didChange
func (s *Server) handleDidChangeTextDocument(
	_ context.Context,
	_ *Documents,
	params *protocol.DidChangeTextDocumentParams,
) (any, error) {
	document := params.TextDocument
	if document.Uri == "" {
		s.logger.Error("Notification without uri.", "method", protocol.MethodTextDocumentDidChange)
		return nil, nil
	}
	if len(params.ContentChanges) == 0 {
		return nil, nil
	}
	// With full synchronization, the last change holds the whole text.
	last := params.ContentChanges[len(params.ContentChanges)-1]
	change, ok := last.Value().(protocol.TextDocumentContentChangeWholeDocument)
	if !ok {
		return nil, NewError(
			jsonrpc.ErrorCodeInvalidParams,
			"change is not the whole document, which full synchronization requires",
		)
	}
	text := change.Text
	s.documents = s.documents.with(newSnapshot(document.Uri, document.Version, text, s.encoding))
	return nil, nil
}

// handleDidCloseTextDocument handles the textDocument/didClose notification, see
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#textDocument_didClose
func (s *Server) handleDidCloseTextDocument(
	_ context.Context,
	_ *Documents,
	params *protocol.DidCloseTextDocumentParams,
) (any, error) {
	if params.TextDocument.Uri == "" {
		s.logger.Error("Notification without uri.", "method", protocol.MethodTextDocumentDidClose)
		return nil, nil
	}
	s.documents = s.documents.without(params.TextDocument.Uri)
	return nil, nil
}

// handleDocumentFormatting handles the textDocument/formatting request. It replaces the whole document with its
// canonical layout, or returns no edit when it is formatted. A document with a syntax error is not formatted, as the
// formatter works on the tokens and would mangle the layout after the error, e.g. join the lines after an
// unterminated string. See
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#textDocument_formatting
func (s *Server) handleDocumentFormatting(
	_ context.Context,
	documents *Documents,
	params *protocol.DocumentFormattingParams,
) ([]protocol.TextEdit, error) {
	snapshot, err := s.openSnapshot(documents, params.TextDocument.Uri)
	if err != nil {
		return nil, err
	}

	if _, err := golrparser.NewParser().Parse(
		golrparser.NewTokenSkipper(golrparser.NewScanner(snapshot.Text, string(snapshot.URI))),
	); err != nil {
		return nil, NewError(int(protocol.LSPErrorCodesRequestFailed), "not formatted: "+err.Error())
	}

	formatted := golrfmt.NewFormatter().Format(snapshot.Text, string(snapshot.URI))
	if bytes.Equal(formatted, snapshot.Text) {
		return []protocol.TextEdit{}, nil
	}
	return []protocol.TextEdit{
		{
			Range: protocol.Range{
				End: snapshot.Position(len(snapshot.Text)),
			},
			NewText: string(formatted),
		},
	}, nil
}

// handleCompletion handles the textDocument/completion request. It offers the keywords valid at the caret and, where
// a symbol may be referenced, the names of the terminals and nonterminals. Fragments are left out, as they are only
// valid inside regular expressions. Every item replaces the word before the caret including a leading "@", which the
// word a client derives leaves out. See
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#textDocument_completion
func (s *Server) handleCompletion(
	_ context.Context,
	documents *Documents,
	params *protocol.CompletionParams,
) ([]protocol.CompletionItem, error) {
	snapshot, err := s.openSnapshot(documents, params.TextDocument.Uri)
	if err != nil {
		return nil, err
	}

	symbols := snapshot.Model()
	offset := snapshot.Offset(params.Position)
	completion := symbols.CompletionAt(offset)
	replaced := snapshot.Range(model.Span{Offset: completion.Offset, Length: offset - completion.Offset})

	items := make([]protocol.CompletionItem, 0, len(completion.Keywords))
	for _, keyword := range completion.Keywords {
		items = append(items, protocol.CompletionItem{
			Label:    keyword,
			Kind:     protocol.NewOptional(protocol.CompletionItemKindKeyword),
			TextEdit: protocol.NewCompletionItemTextEdit(protocol.TextEdit{Range: replaced, NewText: keyword}),
		})
	}
	if !completion.Symbols {
		return items, nil
	}

	offered := map[string]bool{}
	for _, definition := range symbols.Definitions() {
		if definition.Kind == model.KindFragment || offered[definition.Name] {
			continue
		}
		offered[definition.Name] = true

		detail := "terminal"
		if definition.Kind == model.KindNonterminal {
			detail = "nonterminal"
		}
		items = append(items, protocol.CompletionItem{
			Label:    definition.Name,
			Kind:     protocol.NewOptional(protocol.CompletionItemKindVariable),
			Detail:   protocol.NewOptional(detail),
			TextEdit: protocol.NewCompletionItemTextEdit(protocol.TextEdit{Range: replaced, NewText: definition.Name}),
		})
	}
	return items, nil
}

// handleDefinition handles the textDocument/definition request. From a definition or a reference, it returns every
// declaration of the name: the definitions of a symbol, the aliases of a string. A position on no name gives no
// location. See
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#textDocument_definition
func (s *Server) handleDefinition(
	_ context.Context,
	documents *Documents,
	params *protocol.DefinitionParams,
) ([]protocol.Location, error) {
	snapshot, err := s.openSnapshot(documents, params.TextDocument.Uri)
	if err != nil {
		return nil, err
	}

	symbols := snapshot.Model()
	occurrence, ok := symbols.OccurrenceAt(snapshot.Offset(params.Position))
	if !ok {
		return []protocol.Location{}, nil
	}
	declarations := symbols.DeclarationsNamed(occurrence.Name)
	locations := make([]protocol.Location, 0, len(declarations))
	for _, declaration := range declarations {
		locations = append(locations, snapshot.Location(declaration.Span))
	}
	return locations, nil
}

// handleReferences handles the textDocument/references request. It returns the references of the name at the caret
// in the order of the text: for a terminal also those by its aliases, for an alias only those by its string. With
// includeDeclaration, the declarations of the name are added. A position on no name gives no location. See
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#textDocument_references
func (s *Server) handleReferences(
	_ context.Context,
	documents *Documents,
	params *protocol.ReferenceParams,
) ([]protocol.Location, error) {
	snapshot, err := s.openSnapshot(documents, params.TextDocument.Uri)
	if err != nil {
		return nil, err
	}

	symbols := snapshot.Model()
	occurrence, ok := symbols.OccurrenceAt(snapshot.Offset(params.Position))
	if !ok {
		return []protocol.Location{}, nil
	}
	usages := symbols.UsagesNamed(occurrence.Name)
	spans := make([]model.Span, 0, len(usages))
	for _, usage := range usages {
		spans = append(spans, usage.Span)
	}
	if params.Context.IncludeDeclaration {
		for _, declaration := range symbols.DeclarationsNamed(occurrence.Name) {
			spans = append(spans, declaration.Span)
		}
		slices.SortFunc(spans, model.Span.Compare)
	}

	locations := make([]protocol.Location, 0, len(spans))
	for _, span := range spans {
		locations = append(locations, snapshot.Location(span))
	}
	return locations, nil
}

// handlePrepareRename handles the textDocument/prepareRename request. It returns the range of the name at the caret,
// or an error when there is no name to rename. See
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#textDocument_prepareRename
func (s *Server) handlePrepareRename(
	_ context.Context,
	documents *Documents,
	params *protocol.PrepareRenameParams,
) (protocol.PrepareRenameResult, error) {
	snapshot, err := s.openSnapshot(documents, params.TextDocument.Uri)
	if err != nil {
		return protocol.PrepareRenameResult{}, err
	}

	occurrence, ok := snapshot.Model().OccurrenceAt(snapshot.Offset(params.Position))
	if !ok {
		return protocol.PrepareRenameResult{}, errNoNameToRename
	}
	return protocol.NewPrepareRenameResult(snapshot.Range(occurrence.Span)), nil
}

// handleRename handles the textDocument/rename request. It rewrites the declarations and references of the name at the
// caret in one edit. Renaming a terminal keeps the references by its aliases, renaming an alias keeps the terminal. A
// new name which is not valid is answered with the reason as the error. See
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#textDocument_rename
func (s *Server) handleRename(
	_ context.Context,
	documents *Documents,
	params *protocol.RenameParams,
) (protocol.WorkspaceEdit, error) {
	snapshot, err := s.openSnapshot(documents, params.TextDocument.Uri)
	if err != nil {
		return protocol.WorkspaceEdit{}, err
	}

	symbols := snapshot.Model()
	occurrence, ok := symbols.OccurrenceAt(snapshot.Offset(params.Position))
	if !ok {
		return protocol.WorkspaceEdit{}, errNoNameToRename
	}
	if err := symbols.ValidateRename(occurrence, params.NewName); err != nil {
		return protocol.WorkspaceEdit{}, NewError(int(protocol.LSPErrorCodesRequestFailed), err.Error())
	}

	declarations := symbols.DeclarationsNamed(occurrence.Name)
	references := symbols.ReferencesNamed(occurrence.Name)
	spans := make([]model.Span, 0, len(declarations)+len(references))
	for _, declaration := range declarations {
		spans = append(spans, declaration.Span)
	}
	for _, reference := range references {
		spans = append(spans, reference.Span)
	}
	slices.SortFunc(spans, model.Span.Compare)

	edits := make([]protocol.TextEdit, 0, len(spans))
	for _, span := range spans {
		edits = append(edits, protocol.TextEdit{
			Range:   snapshot.Range(span),
			NewText: params.NewName,
		})
	}
	return protocol.WorkspaceEdit{
		Changes: map[protocol.DocumentUri][]protocol.TextEdit{snapshot.URI: edits},
	}, nil
}

// handleSemanticTokensFull handles the textDocument/semanticTokens/full request. It marks the definitions and
// references of terminals and nonterminals, the definitions with the declaration modifier. A reference has the kind
// of the first definition with its name, an unresolved one is a nonterminal. Aliases and references by alias get no
// token, they keep the color of a string. See
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#textDocument_semanticTokens
func (s *Server) handleSemanticTokensFull(
	_ context.Context,
	documents *Documents,
	params *protocol.SemanticTokensParams,
) (protocol.SemanticTokens, error) {
	snapshot, err := s.openSnapshot(documents, params.TextDocument.Uri)
	if err != nil {
		return protocol.SemanticTokens{}, err
	}

	symbols := snapshot.Model()
	data := []int{}
	var previous protocol.Position
	for _, occurrence := range symbols.Occurrences() {
		if occurrence.Role == model.RoleAlias || symbols.IsAlias(occurrence.Name) {
			continue
		}
		modifiers := 0
		if occurrence.Role == model.RoleDefinition {
			modifiers = semanticTokenDeclaration
		}

		// A name is on one line, so its length is the difference of the characters.
		start := snapshot.Position(occurrence.Span.Offset)
		end := snapshot.Position(occurrence.Span.End())
		deltaStart := start.Character
		if start.Line == previous.Line {
			deltaStart -= previous.Character
		}
		data = append(data,
			start.Line-previous.Line,
			deltaStart,
			end.Character-start.Character,
			s.semanticTokenType(symbols, occurrence),
			modifiers,
		)
		previous = start
	}
	return protocol.SemanticTokens{Data: data}, nil
}

// semanticTokenType returns the token type of a definition or reference: the kind of the definition, or for a
// reference the kind of the first definition with its name. An unresolved reference is a nonterminal.
func (s *Server) semanticTokenType(symbols *model.Model, occurrence model.Occurrence) int {
	for _, definition := range symbols.DefinitionsNamed(occurrence.Name) {
		if occurrence.Role == model.RoleReference || definition.Span == occurrence.Span {
			if definition.Kind == model.KindNonterminal {
				return semanticTokenNonterminal
			}
			return semanticTokenTerminal
		}
	}
	return semanticTokenNonterminal
}

// openSnapshot returns the snapshot of the document, or an error for a request on a document which is not open.
func (s *Server) openSnapshot(documents *Documents, uri protocol.DocumentUri) (*Snapshot, error) {
	snapshot, ok := documents.Snapshot(uri)
	if !ok {
		return nil, NewError(jsonrpc.ErrorCodeInvalidParams, "document not open: "+string(uri))
	}
	return snapshot, nil
}
