package protocol_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/protocol"
)

var _ = Describe("Optional", func() {
	It("should be absent as zero value", func() {
		var optional protocol.Optional[int]
		value, ok := optional.Get()
		Expect(ok).To(BeFalse())
		Expect(value).To(BeZero())
		Expect(optional.IsZero()).To(BeTrue())
	})

	It("should hold a zero value as present", func() {
		optional := protocol.NewOptional(0)
		value, ok := optional.Get()
		Expect(ok).To(BeTrue())
		Expect(value).To(BeZero())
		Expect(optional.IsZero()).To(BeFalse())
	})

	DescribeTable("should write the value",
		func(input json.Marshaler, expected string) {
			Expect(input.MarshalJSON()).To(MatchJSON(expected))
		},
		Entry("integer", protocol.NewOptional(7), `7`),
		Entry("zero integer", protocol.NewOptional(0), `0`),
		Entry("empty string", protocol.NewOptional(""), `""`),
		Entry("nil pointer", protocol.NewOptional[*int](nil), `null`),
		Entry("structure", protocol.NewOptional(protocol.Position{Line: 1, Character: 2}),
			`{"line":1,"character":2}`),
		Entry("absent", protocol.Optional[int]{}, `0`),
	)
})
