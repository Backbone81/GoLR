package server_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/jsonrpc"
	"github.com/backbone81/golr/internal/lsp/lsptest"
	"github.com/backbone81/golr/internal/lsp/protocol"
	"github.com/backbone81/golr/internal/lsp/server"
)

var _ = Describe("Request", func() {
	It("answers a request with broken params with InvalidParams and goes on", func() {
		responses, err := run(false,
			`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"capabilities":"broken"}}`,
			initializeRequest,
			shutdownRequest,
			exitNotification,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(HaveLen(3))
		Expect(responses[0]).To(lsptest.BeErrorResponse(1, jsonrpc.ErrorCodeInvalidParams))
		Expect(responses[1]).To(MatchJSON(initializeResponse))
	})

	It("runs a request without id and writes no response", func() {
		responses, err := run(false,
			initializeRequest,
			`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{
				"uri":"file:///a.golr","version":1,"text":""}}}`,
			`{"jsonrpc":"2.0","method":"textDocument/formatting","params":{"textDocument":{"uri":"file:///a.golr"}}}`,
			unknownRequest,
			shutdownRequest,
			exitNotification,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(ConsistOf(
			MatchJSON(initializeResponse),
			lsptest.BeErrorResponse(2, jsonrpc.ErrorCodeMethodNotFound),
			MatchJSON(shutdownResponse),
		))
	})

	It("answers a notification with id and applies it", func() {
		responses, err := run(false,
			initializeRequest,
			`{"jsonrpc":"2.0","id":3,"method":"textDocument/didOpen","params":{"textDocument":{
				"uri":"file:///a.golr","version":1,"text":""}}}`,
			`{"jsonrpc":"2.0","id":2,"method":"textDocument/formatting","params":{
				"textDocument":{"uri":"file:///a.golr"}}}`,
			shutdownRequest,
			exitNotification,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(HaveLen(4))
		Expect(responses[1]).To(MatchJSON(`{"jsonrpc":"2.0","id":3,"result":null}`))
		Expect(responses[2]).NotTo(ContainSubstring("document not open"))
	})

	It("ends the session on exit with id", func() {
		responses, err := run(false,
			initializeRequest,
			shutdownRequest,
			`{"jsonrpc":"2.0","id":3,"method":"exit"}`,
			unknownRequest,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(ConsistOf(MatchJSON(initializeResponse), MatchJSON(shutdownResponse)))
	})

	It("answers a request whose params are no object with InvalidParams", func() {
		responses, err := run(false,
			initializeRequest,
			`{"jsonrpc":"2.0","id":2,"method":"textDocument/formatting","params":"broken"}`,
			shutdownRequest,
			exitNotification,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(HaveLen(3))
		Expect(responses[1]).To(lsptest.BeErrorResponse(2, jsonrpc.ErrorCodeInvalidParams))
	})

	It("answers requests with RequestCancelled when the session is cancelled", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		var output bytes.Buffer
		srv := server.New(slog.New(slog.DiscardHandler), "", false)
		err := srv.Run(ctx, strings.NewReader(messages(initializeRequest, shutdownRequest, exitNotification)), &output)
		Expect(err).To(MatchError(server.ErrSessionEndedWithoutShutdown))
		Expect(readContents(&output)).To(ConsistOf(
			lsptest.BeErrorResponse(1, int(protocol.LSPErrorCodesRequestCancelled)),
			lsptest.BeErrorResponse(9, int(protocol.ErrorCodesServerNotInitialized)),
		))
	})

	DescribeTable("ends the session on a notification which fails",
		func(message string) {
			responses, err := run(false, initializeRequest, message, unknownRequest)
			Expect(err).To(HaveOccurred())
			Expect(responses).To(ConsistOf(MatchJSON(initializeResponse)))
		},
		Entry("with params which do not unmarshal",
			`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":"broken"}}`),
		Entry("with params which are no object",
			`{"jsonrpc":"2.0","method":"textDocument/didChange","params":"broken"}`),
		Entry("with $/cancelRequest params which are no object",
			`{"jsonrpc":"2.0","method":"$/cancelRequest","params":"broken"}`),
		Entry("as a request without id",
			`{"jsonrpc":"2.0","method":"textDocument/formatting","params":{"textDocument":{"uri":"file:///a.golr"}}}`),
	)
})
