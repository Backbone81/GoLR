package lsp_test

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/onsi/gomega/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp"
	"github.com/backbone81/golr/internal/lsp/base"
	"github.com/backbone81/golr/internal/lsp/lsptest"
)

const (
	initializeRequest = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"processId":null,"capabilities":{}}}`
	shutdownRequest   = `{"jsonrpc":"2.0","id":99,"method":"shutdown"}`
	exitNotification  = `{"jsonrpc":"2.0","method":"exit"}`

	initializeResponseUTF16 = `{"jsonrpc":"2.0","id":1,"result":{
		"capabilities":{"positionEncoding":"utf-16","textDocumentSync":1,
		"codeLensProvider":{},"completionProvider":{"triggerCharacters":["@"]},
		"definitionProvider":true,"referencesProvider":true,"renameProvider":true,
		"documentFormattingProvider":true,
		"semanticTokensProvider":{"legend":{"tokenTypes":["class","enum"],"tokenModifiers":["declaration"]},"full":true}},
		"serverInfo":{"name":"GoLR"}}}`
	shutdownResponse = `{"jsonrpc":"2.0","id":99,"result":null}`

	// didOpenUnformattedNotification opens a document with CRLF line endings and a character outside the Basic
	// Multilingual Plane, which formats to formatted.
	didOpenUnformattedNotification = `{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{
		"uri":"file:///a.golr","languageId":"golr","version":1,
		"text":"@scanner{\r\nPLUS:\"+\"; // 😀\r\nINTEGER:/[0-9]+/;}@parser{s:INTEGER;}"}}}`
	formatted = `@scanner {\n    PLUS:    \"+\"; // 😀\n    INTEGER: /[0-9]+/;\n}\n\n@parser {\n    s\n        : ` +
		`INTEGER\n        ;\n}\n`

	formattingRequest = `{"jsonrpc":"2.0","id":2,"method":"textDocument/formatting","params":{
		"textDocument":{"uri":"file:///a.golr"},"options":{"tabSize":4,"insertSpaces":true}}}`
)

var (
	// formattingSession opens a document, formats it and ends the session.
	formattingSession = []string{
		initializeRequest,
		`{"jsonrpc":"2.0","method":"initialized","params":{}}`,
		didOpenUnformattedNotification,
		formattingRequest,
		shutdownRequest,
		exitNotification,
	}

	// formattingSessionResponses are the responses to formattingSession.
	formattingSessionResponses = []types.GomegaMatcher{
		MatchJSON(initializeResponseUTF16),
		MatchJSON(`{"jsonrpc":"2.0","id":2,"result":[{
			"range":{"start":{"line":0,"character":0},"end":{"line":2,"character":37}},
			"newText":"` + formatted + `"}]}`),
		MatchJSON(shutdownResponse),
	}
)

var _ = Describe("Serve", func() {
	DescribeTable("formats a document",
		func(options ...lsp.Option) {
			responses, err := serve(options, formattingSession...)
			Expect(err).NotTo(HaveOccurred())
			Expect(responses).To(ConsistOf(formattingSessionResponses))
		},
		Entry("inline", lsp.WithoutParallelism()),
		Entry("in parallel"),
	)
})

// serve runs a session on the given message bodies and returns the bodies the server wrote.
func serve(options []lsp.Option, messages ...string) ([]string, error) {
	var input bytes.Buffer
	for _, message := range messages {
		Expect(base.WriteMessage(&input, []byte(message))).To(Succeed())
	}
	var output bytes.Buffer
	options = append(options, lsp.WithStreams(&input, &output), lsp.WithLogger(slog.New(slog.DiscardHandler)))
	err := lsp.Serve(context.Background(), options...)

	bodies := readBodies(&output)
	lsptest.ExpectDecodableResults(messages, bodies)
	return bodies, err
}
