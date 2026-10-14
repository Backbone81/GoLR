package server_test

import (
	"context"
	"io"
	"log/slog"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/jsonrpc"
	"github.com/backbone81/golr/internal/lsp/lsptest"
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
})
