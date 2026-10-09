package jsonrpc_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/jsonrpc"
)

var _ = Describe("ID", func() {
	DescribeTable("round trips",
		func(input string, expected string) {
			var id jsonrpc.ID
			Expect(json.Unmarshal([]byte(input), &id)).To(Succeed())
			Expect(id.String()).To(Equal(expected))
			Expect(json.Marshal(id)).To(MatchJSON(input))
		},
		Entry("an integer", `7`, `7`),
		Entry("a negative integer", `-3`, `-3`),
		Entry("a string", `"abc"`, `"abc"`),
		Entry("a string of digits", `"7"`, `"7"`),
		Entry("null", `null`, `null`),
	)

	It("keeps an integer and a string of the same digits apart", func() {
		var number, text jsonrpc.ID
		Expect(json.Unmarshal([]byte(`7`), &number)).To(Succeed())
		Expect(json.Unmarshal([]byte(`"7"`), &text)).To(Succeed())
		Expect(number.String()).NotTo(Equal(text.String()))
	})

	It("forgets the previous value when decoded again", func() {
		var id jsonrpc.ID
		Expect(json.Unmarshal([]byte(`7`), &id)).To(Succeed())
		Expect(json.Unmarshal([]byte(`"s"`), &id)).To(Succeed())
		Expect(json.Marshal(id)).To(MatchJSON(`"s"`))
		Expect(json.Unmarshal([]byte(`null`), &id)).To(Succeed())
		Expect(json.Marshal(id)).To(MatchJSON(`null`))
	})

	It("marshals when held as a value", func() {
		var id jsonrpc.ID
		Expect(json.Unmarshal([]byte(`7`), &id)).To(Succeed())
		Expect(json.Marshal(struct{ ID jsonrpc.ID }{ID: id})).To(MatchJSON(`{"ID":7}`))
	})

	DescribeTable("rejects",
		func(input string) {
			var id jsonrpc.ID
			Expect(json.Unmarshal([]byte(input), &id)).NotTo(Succeed())
		},
		Entry("a fraction", `1.5`),
		Entry("a boolean", `true`),
		Entry("an object", `{}`),
		Entry("an array", `[]`),
	)
})
