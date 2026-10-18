package lsp_test

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp"
	"github.com/backbone81/golr/internal/lsp/base"
	"github.com/backbone81/golr/internal/lsp/jsonrpc"
	"github.com/backbone81/golr/internal/lsp/lsptest"
	"github.com/backbone81/golr/internal/lsp/protocol"
	"github.com/backbone81/golr/internal/utils"
)

const (
	sessionURI = "file:///session.golr"

	// sessionInitializeParams announce what Neovim does: UTF-8 positions and prepareRename.
	sessionInitializeParams = `{"processId":null,"capabilities":{
		"general":{"positionEncodings":["utf-8","utf-16"]},
		"textDocument":{"rename":{"prepareSupport":true}}}}`

	sessionInitializeResponse = `{"jsonrpc":"2.0","id":1,"result":{
		"capabilities":{"positionEncoding":"utf-8","textDocumentSync":1,
		"codeLensProvider":{},"completionProvider":{"triggerCharacters":["@"]},
		"definitionProvider":true,"referencesProvider":true,"renameProvider":{"prepareProvider":true},
		"documentFormattingProvider":true,
		"semanticTokensProvider":{"legend":{"tokenTypes":["class","enum"],"tokenModifiers":["declaration"]},"full":true}},
		"serverInfo":{"name":"GoLR"}}}`
)

// The session follows an editor through the life of a document: it is opened, formatted, highlighted, navigated,
// edited with completion and rename, and closed. Each step works on the text the steps before left, which the client
// keeps as an editor does.
var _ = Describe("Session", func() {
	DescribeTable("serves an editor through a whole session",
		func(options ...lsp.Option) {
			client := newScriptedClient(options)
			Expect(client.initialize()).To(MatchJSON(sessionInitializeResponse))
			client.notify(protocol.MethodInitialized, protocol.InitializedParams{})

			By("opening an unformatted document")
			// The comment in front of INTEGER makes UTF-8 and UTF-16 positions differ on its line.
			client.open(utils.HereDoc(`
				@scanner{
				PLUS:"+";
				INTEGER:/[0-9]+/;
				}
				@parser{
				@start:expression;
				expression:expression PLUS term|term;
				term:/* 😀 */ INTEGER;
				}
			`))

			By("formatting it")
			client.change(client.formatting())
			Expect(client.text).To(Equal(utils.HereDoc(`
				@scanner {
				    PLUS:    "+";
				    INTEGER: /[0-9]+/;
				}

				@parser {
				    @start: expression;

				    expression
				        : expression PLUS term
				        | term
				        ;

				    term
				        : /* 😀 */ INTEGER
				        ;
				}
			`)))

			By("highlighting the symbols")
			Expect(client.semanticTokens()).To(Equal([]string{
				"2: PLUS enum declaration",
				"3: INTEGER enum declaration",
				"7: expression class",
				"9: expression class declaration",
				"10: expression class",
				"10: PLUS enum",
				"10: term class",
				"11: term class",
				"14: term class declaration",
				"15: INTEGER enum",
			}))

			By("showing the reference counts")
			Expect(client.codeLenses()).To(Equal([]string{
				"2: PLUS 1 reference",
				"3: INTEGER 1 reference",
				"9: expression 2 references",
				"14: term 2 references",
			}))

			By("going to the definition of a reference")
			Expect(client.definition(client.at(11, "term"))).To(Equal([]string{"14: term"}))

			By("finding the references of a definition")
			Expect(client.references(client.at(14, "term"), true)).To(Equal([]string{
				"10: term",
				"11: term",
				"14: term",
			}))

			By("completing a partially typed name")
			client.change([]protocol.TextEdit{{
				Range:   protocol.Range{Start: protocol.Position{Line: 11}, End: protocol.Position{Line: 11}},
				NewText: "        | ex\n",
			}})
			items := client.completion(client.after(12, "ex"))
			Expect(client.completionLabels(items)).To(Equal([]string{
				"@empty", "@error", "@precedence", "@name", "PLUS", "INTEGER", "expression", "term",
			}))
			client.change([]protocol.TextEdit{client.completionEdit(items, "expression")})
			Expect(client.line(12)).To(Equal("        | expression"))

			By("renaming a nonterminal")
			Expect(client.prepareRename(client.at(11, "term"))).To(Equal("11: term"))
			client.change(client.rename(client.at(11, "term"), "factor"))
			Expect(client.text).To(Equal(utils.HereDoc(`
				@scanner {
				    PLUS:    "+";
				    INTEGER: /[0-9]+/;
				}

				@parser {
				    @start: expression;

				    expression
				        : expression PLUS factor
				        | factor
				        | expression
				        ;

				    factor
				        : /* 😀 */ INTEGER
				        ;
				}
			`)))
			Expect(client.codeLenses()).To(ContainElement("9: expression 3 references"))

			By("rejecting a rename which would merge two terminals")
			response := client.request(protocol.MethodTextDocumentRename, protocol.RenameParams{
				TextDocument: protocol.TextDocumentIdentifier{Uri: sessionURI},
				Position:     client.at(2, "PLUS"),
				NewName:      "INTEGER",
			})
			Expect(response).To(lsptest.BeErrorResponse(client.id, int(protocol.LSPErrorCodesRequestFailed)))
			Expect(response).To(ContainSubstring("already defined"))

			By("closing the document")
			client.notify(protocol.MethodTextDocumentDidClose, protocol.DidCloseTextDocumentParams{
				TextDocument: protocol.TextDocumentIdentifier{Uri: sessionURI},
			})
			Expect(client.request(protocol.MethodTextDocumentCodeLens, protocol.CodeLensParams{
				TextDocument: protocol.TextDocumentIdentifier{Uri: sessionURI},
			})).To(lsptest.BeErrorResponse(client.id, jsonrpc.ErrorCodeInvalidParams))

			By("shutting down")
			Expect(client.request(protocol.MethodShutdown, nil)).To(MatchJSON(
				fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"result":null}`, client.id)))
			client.notify(protocol.MethodExit, nil)
			Expect(client.wait()).To(Succeed())
		},
		Entry("inline", lsp.WithoutParallelism()),
		Entry("in parallel"),
	)
})

// scriptedClient plays an editor in a session with the server: it sends one message at a time, waits for the response
// of every request and keeps the text of the one document it edits. It negotiates UTF-8 positions, so a character
// is a byte of the text.
type scriptedClient struct {
	input  *io.PipeWriter
	output *base.ProtocolParser
	done   chan error

	// id is the id of the last request.
	id      int
	legend  protocol.SemanticTokensLegend
	text    string
	version int

	// messages and responses are the contents sent and received, which are checked for decodable results at the end.
	messages  []string
	responses []string
}

// newScriptedClient starts a server with the options on pipes to the client.
func newScriptedClient(options []lsp.Option) *scriptedClient {
	serverInput, clientOutput := io.Pipe()
	clientInput, serverOutput := io.Pipe()
	client := &scriptedClient{
		input:  clientOutput,
		output: base.NewProtocolParser(clientInput),
		done:   make(chan error, 1),
	}
	DeferCleanup(clientInput.Close)
	DeferCleanup(clientOutput.Close)

	options = append(options,
		lsp.WithStreams(serverInput, serverOutput),
		lsp.WithLogger(slog.New(slog.DiscardHandler)),
	)
	go client.serve(options)
	return client
}

// serve runs the server until the session ends.
func (c *scriptedClient) serve(options []lsp.Option) {
	c.done <- lsp.Serve(context.Background(), options...)
}

// wait returns the error of the server after the session ended and checks every result the server sent.
func (c *scriptedClient) wait() error {
	err := <-c.done
	lsptest.ExpectDecodableResults(c.messages, c.responses)
	return err
}

// initialize sends the initialize request and keeps the legend of the semantic tokens.
func (c *scriptedClient) initialize() string {
	response := c.request(protocol.MethodInitialize, json.RawMessage(sessionInitializeParams))
	result, err := protocol.DecodeInitializeResult(c.result(response))
	Expect(err).NotTo(HaveOccurred())
	options, ok := result.Capabilities.SemanticTokensProvider.Value().(protocol.SemanticTokensOptions)
	Expect(ok).To(BeTrue())
	c.legend = options.Legend
	return response
}

// open opens the document with the text.
func (c *scriptedClient) open(text string) {
	c.text = text
	c.version = 1
	c.notify(protocol.MethodTextDocumentDidOpen, protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{Uri: sessionURI, LanguageId: "golr", Version: c.version, Text: c.text},
	})
}

// change applies the edits to the text and sends the whole text with the next version.
func (c *scriptedClient) change(edits []protocol.TextEdit) {
	// Applied from the end, the offsets of the edits before stay valid.
	edits = slices.Clone(edits)
	slices.SortFunc(edits, c.compareEditsFromEnd)
	for _, edit := range edits {
		c.text = c.text[:c.offset(edit.Range.Start)] + edit.NewText + c.text[c.offset(edit.Range.End):]
	}

	c.version++
	c.notify(protocol.MethodTextDocumentDidChange, protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{Uri: sessionURI, Version: c.version},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			protocol.NewTextDocumentContentChangeEvent(protocol.TextDocumentContentChangeWholeDocument{Text: c.text}),
		},
	})
}

// compareEditsFromEnd orders edits by their start, the last first.
func (c *scriptedClient) compareEditsFromEnd(a protocol.TextEdit, b protocol.TextEdit) int {
	return cmp.Compare(c.offset(b.Range.Start), c.offset(a.Range.Start))
}

// formatting returns the edits which format the document.
func (c *scriptedClient) formatting() []protocol.TextEdit {
	result := c.call(protocol.MethodTextDocumentFormatting, protocol.DocumentFormattingParams{
		TextDocument: protocol.TextDocumentIdentifier{Uri: sessionURI},
		Options:      protocol.FormattingOptions{TabSize: 4, InsertSpaces: true},
	})
	var edits []protocol.TextEdit
	for _, element := range c.elements(result) {
		edit, err := protocol.DecodeTextEdit(element)
		Expect(err).NotTo(HaveOccurred())
		edits = append(edits, edit)
	}
	return edits
}

// semanticTokens returns the semantic tokens as "<line>: <text> <type>[ <modifier>...]".
func (c *scriptedClient) semanticTokens() []string {
	result := c.call(protocol.MethodTextDocumentSemanticTokensFull, protocol.SemanticTokensParams{
		TextDocument: protocol.TextDocumentIdentifier{Uri: sessionURI},
	})
	tokens, err := protocol.DecodeSemanticTokens(result)
	Expect(err).NotTo(HaveOccurred())
	Expect(len(tokens.Data) % 5).To(BeZero())

	var descriptions []string
	var start protocol.Position
	for idx := 0; idx < len(tokens.Data); idx += 5 {
		deltaLine, deltaStart, length := tokens.Data[idx], tokens.Data[idx+1], tokens.Data[idx+2]
		if deltaLine > 0 {
			start = protocol.Position{Line: start.Line + deltaLine}
		}
		start.Character += deltaStart
		end := protocol.Position{Line: start.Line, Character: start.Character + length}

		parts := []string{c.describe(protocol.Range{Start: start, End: end}), c.legend.TokenTypes[tokens.Data[idx+3]]}
		for bit, modifier := range c.legend.TokenModifiers {
			if tokens.Data[idx+4]&(1<<bit) != 0 {
				parts = append(parts, modifier)
			}
		}
		descriptions = append(descriptions, strings.Join(parts, " "))
	}
	return descriptions
}

// codeLenses returns the code lenses as "<line>: <text> <title>".
func (c *scriptedClient) codeLenses() []string {
	result := c.call(protocol.MethodTextDocumentCodeLens, protocol.CodeLensParams{
		TextDocument: protocol.TextDocumentIdentifier{Uri: sessionURI},
	})
	var descriptions []string
	for _, element := range c.elements(result) {
		lens, err := protocol.DecodeCodeLens(element)
		Expect(err).NotTo(HaveOccurred())
		command, ok := lens.Command.Get()
		Expect(ok).To(BeTrue())
		Expect(command.Command).To(BeEmpty())
		descriptions = append(descriptions, c.describe(lens.Range)+" "+command.Title)
	}
	return descriptions
}

// definition returns the locations of the declarations of the name at the position, see describe.
func (c *scriptedClient) definition(position protocol.Position) []string {
	return c.locations(c.call(protocol.MethodTextDocumentDefinition, protocol.DefinitionParams{
		TextDocument: protocol.TextDocumentIdentifier{Uri: sessionURI},
		Position:     position,
	}))
}

// references returns the locations of the references of the name at the position, see describe.
func (c *scriptedClient) references(position protocol.Position, includeDeclaration bool) []string {
	return c.locations(c.call(protocol.MethodTextDocumentReferences, protocol.ReferenceParams{
		TextDocument: protocol.TextDocumentIdentifier{Uri: sessionURI},
		Position:     position,
		Context:      protocol.ReferenceContext{IncludeDeclaration: includeDeclaration},
	}))
}

// completion returns the completion items at the position.
func (c *scriptedClient) completion(position protocol.Position) []protocol.CompletionItem {
	result, err := protocol.DecodeCompletionRequestResult(c.call(protocol.MethodTextDocumentCompletion,
		protocol.CompletionParams{
			TextDocument: protocol.TextDocumentIdentifier{Uri: sessionURI},
			Position:     position,
		}))
	Expect(err).NotTo(HaveOccurred())
	items, ok := result.Value().([]protocol.CompletionItem)
	Expect(ok).To(BeTrue())
	return items
}

// completionLabels returns the labels of the items.
func (c *scriptedClient) completionLabels(items []protocol.CompletionItem) []string {
	labels := make([]string, 0, len(items))
	for _, item := range items {
		labels = append(labels, item.Label)
	}
	return labels
}

// completionEdit returns the edit of the item with the label.
func (c *scriptedClient) completionEdit(items []protocol.CompletionItem, label string) protocol.TextEdit {
	for _, item := range items {
		if item.Label != label {
			continue
		}
		edit, ok := item.TextEdit.Value().(protocol.TextEdit)
		Expect(ok).To(BeTrue())
		return edit
	}
	Fail("no completion item " + label)
	return protocol.TextEdit{}
}

// prepareRename returns the range to rename at the position, see describe.
func (c *scriptedClient) prepareRename(position protocol.Position) string {
	result, err := protocol.DecodePrepareRenameResult(c.call(protocol.MethodTextDocumentPrepareRename,
		protocol.PrepareRenameParams{
			TextDocument: protocol.TextDocumentIdentifier{Uri: sessionURI},
			Position:     position,
		}))
	Expect(err).NotTo(HaveOccurred())
	renamed, ok := result.Value().(protocol.Range)
	Expect(ok).To(BeTrue())
	return c.describe(renamed)
}

// rename returns the edits of the document which rename the name at the position.
func (c *scriptedClient) rename(position protocol.Position, newName string) []protocol.TextEdit {
	edit, err := protocol.DecodeWorkspaceEdit(c.call(protocol.MethodTextDocumentRename, protocol.RenameParams{
		TextDocument: protocol.TextDocumentIdentifier{Uri: sessionURI},
		Position:     position,
		NewName:      newName,
	}))
	Expect(err).NotTo(HaveOccurred())
	Expect(edit.Changes).To(HaveLen(1))
	return edit.Changes[sessionURI]
}

// locations returns the locations of the result, see describe.
func (c *scriptedClient) locations(result any) []string {
	var descriptions []string
	for _, element := range c.elements(result) {
		location, err := protocol.DecodeLocation(element)
		Expect(err).NotTo(HaveOccurred())
		Expect(location.Uri).To(BeEquivalentTo(sessionURI))
		descriptions = append(descriptions, c.describe(location.Range))
	}
	return descriptions
}

// call sends a request and returns its result, which has to be no error.
func (c *scriptedClient) call(method string, params any) any {
	return c.result(c.request(method, params))
}

// request sends a request with the next id and returns its response.
func (c *scriptedClient) request(method string, params any) string {
	c.id++
	c.send(map[string]any{"jsonrpc": "2.0", "id": c.id, "method": method, "params": params})
	Expect(c.output.Next()).To(BeTrue(), "response to %s", method)
	response := string(c.output.Content())
	c.responses = append(c.responses, response)
	return response
}

// notify sends a notification.
func (c *scriptedClient) notify(method string, params any) {
	c.send(map[string]any{"jsonrpc": "2.0", "method": method, "params": params})
}

// send writes the message to the server.
func (c *scriptedClient) send(message map[string]any) {
	if message["params"] == nil {
		delete(message, "params")
	}
	content, err := json.Marshal(message)
	Expect(err).NotTo(HaveOccurred())
	c.messages = append(c.messages, string(content))
	Expect(base.WriteMessage(c.input, content)).To(Succeed())
}

// result returns the result of the response as decoded JSON, which the decode functions of protocol take.
func (c *scriptedClient) result(response string) any {
	var responseObject jsonrpc.ResponseObject
	Expect(json.Unmarshal([]byte(response), &responseObject)).To(Succeed())
	Expect(responseObject.Error).To(BeNil(), response)

	decoder := json.NewDecoder(bytes.NewReader(responseObject.Result))
	decoder.UseNumber()
	var result any
	Expect(decoder.Decode(&result)).To(Succeed())
	return result
}

// elements returns the elements of a result which is an array.
func (c *scriptedClient) elements(result any) []any {
	elements, ok := result.([]any)
	Expect(ok).To(BeTrue(), "result is no array: %v", result)
	return elements
}

// describe returns the range as "<line>: <text>", with the line counted from 1 as an editor shows it.
func (c *scriptedClient) describe(textRange protocol.Range) string {
	return fmt.Sprintf("%d: %s", textRange.Start.Line+1, c.text[c.offset(textRange.Start):c.offset(textRange.End)])
}

// at returns the position of the first occurrence of the needle in the line, counted from 1.
func (c *scriptedClient) at(line int, needle string) protocol.Position {
	character := strings.Index(c.line(line), needle)
	Expect(character).To(BeNumerically(">=", 0), "%q not in line %d", needle, line)
	return protocol.Position{Line: line - 1, Character: character}
}

// after returns the position right after the first occurrence of the needle in the line, counted from 1.
func (c *scriptedClient) after(line int, needle string) protocol.Position {
	position := c.at(line, needle)
	position.Character += len(needle)
	return position
}

// line returns the line of the text, counted from 1, without its line ending.
func (c *scriptedClient) line(line int) string {
	lines := strings.Split(c.text, "\n")
	Expect(line).To(BeNumerically("<=", len(lines)))
	return lines[line-1]
}

// offset returns the byte offset of the UTF-8 position in the text.
func (c *scriptedClient) offset(position protocol.Position) int {
	offset := 0
	for range position.Line {
		next := strings.IndexByte(c.text[offset:], '\n')
		Expect(next).To(BeNumerically(">=", 0), "line %d beyond the text", position.Line)
		offset += next + 1
	}
	return offset + position.Character
}
