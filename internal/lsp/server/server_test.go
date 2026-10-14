package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/base"
	"github.com/backbone81/golr/internal/lsp/jsonrpc"
	"github.com/backbone81/golr/internal/lsp/lsptest"
	"github.com/backbone81/golr/internal/lsp/protocol"
	"github.com/backbone81/golr/internal/lsp/server"
)

var _ = Describe("Run", func() {
	DescribeTable("answers requests",
		func(parallel bool) {
			responses, err := run(parallel, initializeRequest, unknownRequest, shutdownRequest, exitNotification)
			Expect(err).NotTo(HaveOccurred())
			Expect(responses).To(ConsistOf(
				MatchJSON(initializeResponse),
				lsptest.BeErrorResponse(2, jsonrpc.ErrorCodeMethodNotFound),
				MatchJSON(shutdownResponse),
			))
		},
		Entry("inline", false),
		Entry("in parallel", true),
	)

	It("ignores unknown notifications", func() {
		responses, err := run(false, initializeRequest, `{"jsonrpc":"2.0","method":"$/unknown","params":{}}`, unknownRequest,
			shutdownRequest, exitNotification)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(HaveLen(3))
		Expect(responses[1]).To(lsptest.BeErrorResponse(2, jsonrpc.ErrorCodeMethodNotFound))
	})

	It("drops notifications before initialize", func() {
		responses, err := run(false,
			`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{
				"uri":"file:///a.golr","version":1,"text":""}}}`,
			initializeRequest,
			`{"jsonrpc":"2.0","id":2,"method":"textDocument/formatting","params":{
				"textDocument":{"uri":"file:///a.golr"}}}`,
			shutdownRequest,
			exitNotification,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(HaveLen(3))
		Expect(responses[1]).To(lsptest.BeErrorResponse(2, jsonrpc.ErrorCodeInvalidParams))
	})

	It("ignores a response of the client", func() {
		responses, err := run(false, initializeRequest, `{"jsonrpc":"2.0","id":5,"result":null}`, unknownRequest,
			shutdownRequest, exitNotification)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(HaveLen(3))
	})

	It("stops reading after exit", func() {
		responses, err := run(false, initializeRequest, shutdownRequest, exitNotification, unknownRequest)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(ConsistOf(MatchJSON(initializeResponse), MatchJSON(shutdownResponse)))
	})

	It("fails when the client exits without shutdown", func() {
		_, err := run(false, initializeRequest, exitNotification)
		Expect(err).To(MatchError(server.ErrSessionEndedWithoutShutdown))
	})

	It("fails when the input ends without shutdown", func() {
		_, err := run(false, initializeRequest)
		Expect(err).To(MatchError(server.ErrSessionEndedWithoutShutdown))
	})

	DescribeTable("ends the session on a message it cannot handle",
		func(message string) {
			responses, err := run(false, initializeRequest, message, unknownRequest)
			Expect(err).To(HaveOccurred())
			Expect(responses).To(ConsistOf(MatchJSON(initializeResponse)))
		},
		Entry("broken JSON", `{"jsonrpc":`),
		Entry("a batch", `[{"jsonrpc":"2.0","method":"initialized","params":{}}]`),
		Entry("another JSON-RPC version", `{"jsonrpc":"1.0","method":"initialized"}`),
		Entry("no JSON-RPC version", `{"method":"initialized"}`),
		Entry("neither method nor id", `{"jsonrpc":"2.0"}`),
		Entry("an id of the wrong type", `{"jsonrpc":"2.0","id":true,"method":"initialize"}`),
	)

	It("answers a $/ request with MethodNotFound", func() {
		responses, err := run(false,
			initializeRequest,
			`{"jsonrpc":"2.0","id":3,"method":"$/cancelRequest","params":{"id":2}}`,
			shutdownRequest,
			exitNotification,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(HaveLen(3))
		Expect(responses[1]).To(lsptest.BeErrorResponse(3, jsonrpc.ErrorCodeMethodNotFound))
	})

	It("rejects a second initialize", func() {
		responses, err := run(false, initializeRequest, initializeRequest, shutdownRequest, exitNotification)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses[1]).To(lsptest.BeErrorResponse(1, jsonrpc.ErrorCodeInvalidRequest))
	})

	It("rejects requests after shutdown", func() {
		responses, err := run(false,
			initializeRequest, didOpenUnformattedNotification, shutdownRequest, formattingRequest, exitNotification)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses[2]).To(lsptest.BeErrorResponse(2, jsonrpc.ErrorCodeInvalidRequest))
	})

	It("ends the session on an unsupported charset", func() {
		input := message("Content-Type: application/vscode-jsonrpc; charset=latin1", initializeRequest)
		responses, err := runInput(false, input)
		Expect(err).To(MatchError(server.ErrUnsupportedCharset))
		Expect(responses).To(BeEmpty())
	})

	DescribeTable("accepts the charset",
		func(contentType string) {
			input := message("Content-Type: "+contentType, initializeRequest) + messages(shutdownRequest, exitNotification)
			responses, err := runInput(false, input)
			Expect(err).NotTo(HaveOccurred())
			Expect(responses).To(ConsistOf(MatchJSON(initializeResponse), MatchJSON(shutdownResponse)))
		},
		Entry("utf-8", "application/vscode-jsonrpc; charset=utf-8"),
		Entry("utf8 for backwards compatibility", "application/vscode-jsonrpc; charset=utf8"),
		Entry("quoted in any case", `application/vscode-jsonrpc; charset="UTF-8"`),
		Entry("left out", "application/vscode-jsonrpc"),
	)

	It("ends the session when a response cannot be written", func() {
		srv := server.New(slog.New(slog.DiscardHandler), "", false, 0)
		input := messages(initializeRequest, shutdownRequest, exitNotification)
		err := srv.Run(context.Background(), strings.NewReader(input), brokenWriter{})
		Expect(err).To(MatchError(errBrokenOutput))
	})

	It("ends the session when the stream breaks", func() {
		_, err := runInput(false, "Content-Length: 5\r\n\r\n{}")
		Expect(err).To(MatchError(io.ErrUnexpectedEOF))
	})

	It("handles a stream of messages in parallel", func() {
		// The race detector reports conflicting accesses which happen during the session. Passing does not prove the
		// absence of data races, but a stream of every handled message makes a flaw in the parallel path likely to show.
		srv := server.New(slog.New(slog.DiscardHandler), "", true, 0)
		Expect(srv.Run(context.Background(), newMessageStream(2*time.Second), io.Discard)).To(Succeed())
	})
})

// messageStream is a session which repeats rounds of messages until its time is up, and then shuts the server down.
type messageStream struct {
	deadline time.Time
	buffer   bytes.Buffer
	nextID   int
	version  int
	ended    bool
}

// newMessageStream returns a stream which initializes the server and opens a document, and repeats its rounds for
// the duration.
func newMessageStream(duration time.Duration) *messageStream {
	stream := &messageStream{
		deadline: time.Now().Add(duration),
		nextID:   1,
	}
	stream.writeRequest(protocol.MethodInitialize, protocol.InitializeParams{})
	stream.writeNotification(protocol.MethodInitialized, protocol.InitializedParams{})
	stream.writeNotification(protocol.MethodTextDocumentDidOpen, protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{Uri: "file:///a.golr", LanguageId: "golr", Text: `@scanner{PLUS:"+";}`},
	})
	return stream
}

// Read reads the messages, writing the next round when the previous one is read.
func (m *messageStream) Read(p []byte) (int, error) {
	if m.buffer.Len() == 0 {
		switch {
		case m.ended:
			return 0, io.EOF
		case time.Now().After(m.deadline):
			m.writeRequest(protocol.MethodShutdown, nil)
			m.writeNotification(protocol.MethodExit, nil)
			m.ended = true
		default:
			m.writeRound()
		}
	}
	return m.buffer.Read(p)
}

// writeRound writes one round: a change of the open document with requests on it, one of them cancelled, a request
// without handler, and a second document which is opened, formatted and closed.
func (m *messageStream) writeRound() {
	m.version++
	text := `@scanner{PLUS:"+";}`
	if m.version%2 == 0 {
		text = "@scanner {\n    PLUS: \"+\";\n}\n"
	}
	m.writeNotification(protocol.MethodTextDocumentDidChange, protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{Uri: "file:///a.golr", Version: m.version},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			protocol.NewTextDocumentContentChangeEvent(protocol.TextDocumentContentChangeWholeDocument{Text: text}),
		},
	})
	cancelledID := m.writeRequest(protocol.MethodTextDocumentFormatting, protocol.DocumentFormattingParams{
		TextDocument: protocol.TextDocumentIdentifier{Uri: "file:///a.golr"},
		Options:      protocol.FormattingOptions{TabSize: 4, InsertSpaces: true},
	})
	m.writeRequest(protocol.MethodTextDocumentFormatting, protocol.DocumentFormattingParams{
		TextDocument: protocol.TextDocumentIdentifier{Uri: "file:///a.golr"},
		Options:      protocol.FormattingOptions{TabSize: 4, InsertSpaces: true},
	})
	m.writeRequest(protocol.MethodTextDocumentFormatting, protocol.DocumentFormattingParams{
		TextDocument: protocol.TextDocumentIdentifier{Uri: "file:///a.golr"},
		Options:      protocol.FormattingOptions{TabSize: 4, InsertSpaces: true},
	})
	m.writeNotification(protocol.MethodCancelRequest, protocol.CancelParams{
		Id: protocol.NewCancelParamsId(cancelledID),
	})
	m.writeRequest("unknown/method", nil)
	m.writeNotification(protocol.MethodTextDocumentDidOpen, protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{Uri: "file:///b.golr", LanguageId: "golr", Text: `@scanner{PLUS:"+";}`},
	})
	m.writeRequest(protocol.MethodTextDocumentFormatting, protocol.DocumentFormattingParams{
		TextDocument: protocol.TextDocumentIdentifier{Uri: "file:///b.golr"},
		Options:      protocol.FormattingOptions{TabSize: 4, InsertSpaces: true},
	})
	m.writeNotification(protocol.MethodTextDocumentDidClose, protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{Uri: "file:///b.golr"},
	})
}

// writeRequest writes a request with an id no other request of the stream has, and returns the id. Nil params are
// left out.
func (m *messageStream) writeRequest(method string, params any) int {
	id := m.nextID
	m.nextID++
	requestID := jsonrpc.NewNumberID(id)
	m.write(jsonrpc.RequestObject{
		JSONRPC: jsonrpc.Version,
		Method:  &method,
		Params:  m.marshal(params),
		ID:      &requestID,
	})
	return id
}

// writeNotification writes a notification. Nil params are left out.
func (m *messageStream) writeNotification(method string, params any) {
	m.write(jsonrpc.RequestObject{
		JSONRPC: jsonrpc.Version,
		Method:  &method,
		Params:  m.marshal(params),
	})
}

// marshal returns the JSON of the params, nil for nil params.
func (m *messageStream) marshal(params any) json.RawMessage {
	if params == nil {
		return nil
	}
	content, err := json.Marshal(params)
	Expect(err).NotTo(HaveOccurred())
	return content
}

// write writes the message with its header.
func (m *messageStream) write(message jsonrpc.RequestObject) {
	content, err := json.Marshal(message)
	Expect(err).NotTo(HaveOccurred())
	Expect(base.WriteMessage(&m.buffer, content)).To(Succeed())
}
