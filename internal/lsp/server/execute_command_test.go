package server_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/jsonrpc"
	"github.com/backbone81/golr/internal/lsp/lsptest"
)

var _ = Describe("Execute command", func() {
	DescribeTable("does nothing for golr.noOp",
		func(params string) {
			responses, err := run(false,
				initializeRequest,
				`{"jsonrpc":"2.0","id":2,"method":"workspace/executeCommand","params":`+params+`}`,
				shutdownRequest,
				exitNotification,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(responses).To(HaveLen(3))
			Expect(responses[1]).To(MatchJSON(`{"jsonrpc":"2.0","id":2,"result":null}`))
		},
		Entry("without arguments", `{"command":"golr.noOp"}`),
		Entry("with arguments", `{"command":"golr.noOp","arguments":[1,"a",null]}`),
	)

	DescribeTable("rejects an unknown command",
		func(params string) {
			responses, err := run(false,
				initializeRequest,
				`{"jsonrpc":"2.0","id":2,"method":"workspace/executeCommand","params":`+params+`}`,
				shutdownRequest,
				exitNotification,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(responses).To(HaveLen(3))
			Expect(responses[1]).To(lsptest.BeErrorResponse(2, jsonrpc.ErrorCodeInvalidParams))
			Expect(responses[1]).To(ContainSubstring("unknown command"))
		},
		Entry("with another name", `{"command":"golr.referenceCount"}`),
		Entry("with an empty name", `{"command":""}`),
	)
})
