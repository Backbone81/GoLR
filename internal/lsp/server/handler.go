//nolint:nilnil // A handler without result returns nil, which is the null result of JSON-RPC.
package server

import (
	"bytes"
	"context"
	"slices"

	golrfmt "github.com/backbone81/golr/internal/fmt"
	"github.com/backbone81/golr/internal/lsp/jsonrpc"
	"github.com/backbone81/golr/internal/lsp/protocol"
	golrparser "github.com/backbone81/golr/internal/parsergen/frontend/golr/parser"
)

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
			PositionEncoding:           protocol.NewOptional(s.encoding),
			TextDocumentSync:           protocol.NewServerCapabilitiesTextDocumentSync(protocol.TextDocumentSyncKindFull),
			DocumentFormattingProvider: protocol.NewServerCapabilitiesDocumentFormattingProvider(true),
		},
		ServerInfo: protocol.NewOptional(serverInfo),
	}, nil
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
	snapshot, ok := documents.Snapshot(params.TextDocument.Uri)
	if !ok {
		return nil, NewError(jsonrpc.ErrorCodeInvalidParams, "document not open: "+string(params.TextDocument.Uri))
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
