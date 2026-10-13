package server_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/jsonrpc"
	"github.com/backbone81/golr/internal/lsp/lsptest"
)

var _ = Describe("Handler", func() {
	DescribeTable("formats a document",
		func(parallel bool) {
			responses, err := run(parallel,
				initializeRequest,
				`{"jsonrpc":"2.0","method":"initialized","params":{}}`,
				didOpenUnformattedNotification,
				formattingRequest,
				shutdownRequest,
				exitNotification,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(responses).To(ConsistOf(
				MatchJSON(initializeResponse),
				MatchJSON(`{"jsonrpc":"2.0","id":2,"result":[{
					"range":{"start":{"line":0,"character":0},"end":{"line":2,"character":18}},
					"newText":"`+formatted+`"}]}`),
				MatchJSON(shutdownResponse),
			))
		},
		Entry("inline", false),
		Entry("in parallel", true),
	)

	It("returns no edit for a formatted document", func() {
		responses, err := run(false,
			initializeRequest,
			didOpenUnformattedNotification,
			didChangeToFormattedNotification,
			formattingRequest,
			shutdownRequest,
			exitNotification,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(ContainElement(MatchJSON(`{"jsonrpc":"2.0","id":2,"result":[]}`)))
	})

	DescribeTable("counts columns in the negotiated position encoding",
		func(positionEncodings string, character int) {
			responses, err := run(false,
				`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"processId":null,"capabilities":{
					"general":{"positionEncodings":`+positionEncodings+`}}}}`,
				// The last line ends in a character of four UTF-8 bytes and two UTF-16 units.
				`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{
					"uri":"file:///a.golr","languageId":"golr","version":1,
					"text":"@scanner{A:\"a\";} // \ud83d\ude00"}}}`,
				formattingRequest, shutdownRequest, exitNotification)
			Expect(err).NotTo(HaveOccurred())
			Expect(responses[1]).To(ContainSubstring(fmt.Sprintf(`"end":{"line":0,"character":%d}`, character)))
		},
		Entry("UTF-8 when the client supports it", `["utf-16","utf-8"]`, 24),
		Entry("UTF-16 otherwise", `["utf-32"]`, 22),
	)

	It("rejects formatting a document which is not open", func() {
		responses, err := run(false,
			initializeRequest,
			didOpenUnformattedNotification,
			didCloseNotification,
			formattingRequest,
			shutdownRequest,
			exitNotification,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(ContainElement(lsptest.BeErrorResponse(2, jsonrpc.ErrorCodeInvalidParams)))
	})

	It("keeps the document on a didChange without changes", func() {
		didOpen := `{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{
			"uri":"file:///a.golr","version":1,"text":"@scanner{A:\"a\";}"}}}`
		didChange := `{"jsonrpc":"2.0","method":"textDocument/didChange","params":{
			"textDocument":{"uri":"file:///a.golr","version":2},"contentChanges":[]}}`
		formatting := `{"jsonrpc":"2.0","id":2,"method":"textDocument/formatting","params":{
			"textDocument":{"uri":"file:///a.golr"}}}`

		unchanged, err := run(false, initializeRequest, didOpen, formatting, shutdownRequest, exitNotification)
		Expect(err).NotTo(HaveOccurred())
		responses, err := run(false, initializeRequest, didOpen, didChange, formatting, shutdownRequest, exitNotification)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(HaveLen(3))
		Expect(responses[1]).To(ContainSubstring(`"newText"`))
		Expect(responses[1]).To(MatchJSON(unchanged[1]))
	})

	It("ignores a $/cancelRequest without id or with an unknown id", func() {
		responses, err := run(false,
			initializeRequest,
			`{"jsonrpc":"2.0","method":"$/cancelRequest","params":{}}`,
			`{"jsonrpc":"2.0","method":"$/cancelRequest"}`,
			`{"jsonrpc":"2.0","method":"$/cancelRequest","params":{"id":99}}`,
			unknownRequest,
			shutdownRequest,
			exitNotification,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(HaveLen(3))
		Expect(responses[1]).To(lsptest.BeErrorResponse(2, jsonrpc.ErrorCodeMethodNotFound))
	})

	DescribeTable("ignores a notification without uri",
		func(notification string) {
			responses, err := run(false,
				initializeRequest,
				notification,
				`{"jsonrpc":"2.0","id":2,"method":"textDocument/formatting","params":{"textDocument":{"uri":""}}}`,
				shutdownRequest,
				exitNotification,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(responses).To(HaveLen(3))
			Expect(responses[1]).To(lsptest.BeErrorResponse(2, jsonrpc.ErrorCodeInvalidParams))
		},
		Entry("didOpen", `{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{
			"version":1,"text":""}}}`),
		Entry("didChange", `{"jsonrpc":"2.0","method":"textDocument/didChange","params":{
			"textDocument":{"version":2},"contentChanges":[{"text":""}]}}`),
		Entry("didChange with null params", `{"jsonrpc":"2.0","method":"textDocument/didChange","params":null}`),
	)
})
