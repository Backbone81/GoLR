package server_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"testing"

	"github.com/onsi/gomega/format"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/base"
	"github.com/backbone81/golr/internal/lsp/lsptest"
	"github.com/backbone81/golr/internal/lsp/server"
)

const (
	initializeRequest  = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"capabilities":{}}}`
	initializeResponse = `{"jsonrpc":"2.0","id":1,"result":{
		"capabilities":{"positionEncoding":"utf-16","textDocumentSync":1,
		"codeLensProvider":{},"completionProvider":{"triggerCharacters":["@"]},
		"definitionProvider":true,"referencesProvider":true,"renameProvider":true,
		"documentFormattingProvider":true,
		"semanticTokensProvider":{"legend":{"tokenTypes":["class","enum"],"tokenModifiers":["declaration"]},"full":true}},
		"serverInfo":{"name":"GoLR"}}}`
	unknownRequest   = `{"jsonrpc":"2.0","id":2,"method":"unknown/method"}`
	shutdownRequest  = `{"jsonrpc":"2.0","id":9,"method":"shutdown"}`
	shutdownResponse = `{"jsonrpc":"2.0","id":9,"result":null}`
	exitNotification = `{"jsonrpc":"2.0","method":"exit"}`

	// didOpenUnformattedNotification opens a document with CRLF line endings and a character outside the Basic
	// Multilingual Plane, which formats to formatted.
	didOpenUnformattedNotification = `{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{
		"uri":"file:///a.golr","languageId":"golr","version":1,
		"text":"@scanner{\r\nPLUS:\"+\"; // 😀\r\nINTEGER:/[0-9]+/;}@parser{s:INTEGER;}"}}}`
	formatted = `@scanner {\n    PLUS:    \"+\"; // 😀\n    INTEGER: /[0-9]+/;\n}\n\n@parser ` +
		`{\n    s\n        : INTEGER\n        ;\n}\n`
	didChangeToFormattedNotification = `{"jsonrpc":"2.0","method":"textDocument/didChange","params":{
		"textDocument":{"uri":"file:///a.golr","version":2},"contentChanges":[{"text":"` + formatted + `"}]}}`
	didCloseNotification = `{"jsonrpc":"2.0","method":"textDocument/didClose","params":{
		"textDocument":{"uri":"file:///a.golr"}}}`
	formattingRequest = `{"jsonrpc":"2.0","id":2,"method":"textDocument/formatting","params":{
		"textDocument":{"uri":"file:///a.golr"},"options":{"tabSize":4,"insertSpaces":true}}}`
)

func TestSuite(t *testing.T) {
	format.MaxLength = 0
	RegisterFailHandler(Fail)
	RunSpecs(t, "Server Suite")
}

// errBrokenOutput is the error of every write to brokenWriter.
var errBrokenOutput = errors.New("broken output")

// brokenWriter fails every write.
type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) {
	return 0, errBrokenOutput
}

// message returns a message with the extra header field.
func message(header string, content string) string {
	return "Content-Length: " + strconv.Itoa(len(content)) + "\r\n" + header + "\r\n\r\n" + content
}

// messages returns the messages with the given contents.
func messages(contents ...string) string {
	var input bytes.Buffer
	for _, content := range contents {
		Expect(base.WriteMessage(&input, []byte(content))).To(Succeed())
	}
	return input.String()
}

// run runs a session on the given message contents and returns the contents the server wrote with the error of Run.
func run(parallel bool, contents ...string) ([]string, error) {
	return runInput(parallel, messages(contents...))
}

// runInput runs a session on the raw input and returns the contents the server wrote with the error of Run.
func runInput(parallel bool, input string) ([]string, error) {
	var output bytes.Buffer
	srv := server.New(slog.New(slog.DiscardHandler), "", parallel, 0)
	err := srv.Run(context.Background(), strings.NewReader(input), &output)
	responses := readContents(&output)
	lsptest.ExpectDecodableResults(requestContents(input), responses)
	return responses, err
}

// requestContents returns the contents of the messages in the input up to the first broken one.
func requestContents(input string) []string {
	parser := base.NewProtocolParser(strings.NewReader(input))
	var contents []string
	for parser.Next() {
		contents = append(contents, string(parser.Content()))
	}
	return contents
}

// readContents returns the contents of the messages the server wrote.
func readContents(output io.Reader) []string {
	parser := base.NewProtocolParser(output)
	var contents []string
	for parser.Next() {
		contents = append(contents, string(parser.Content()))
	}
	Expect(parser.Err()).NotTo(HaveOccurred())
	return contents
}
