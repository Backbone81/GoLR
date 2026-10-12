package protocol_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/protocol"
)

var _ = Describe("Unions", func() {
	DescribeTable("should write the value of the arm",
		func(input any, expected string) {
			Expect(json.Marshal(input)).To(MatchJSON(expected))
		},
		Entry("integer", protocol.NewProgressToken(7), `7`),
		Entry("string", protocol.NewProgressToken("token"), `"token"`),
		Entry("structure", protocol.NewHoverContents(protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: "text",
		}), `{"kind":"markdown","value":"text"}`),
		Entry("slice", protocol.NewDefinition([]protocol.Location{}), `[]`),
		Entry("union", protocol.NewDefinitionRequestResult(protocol.NewDefinition(protocol.Location{
			Uri: "file:///a.golr",
		})), `{"uri":"file:///a.golr","range":{"start":{"line":0,"character":0},"end":{"line":0,"character":0}}}`),
		Entry("empty literal", protocol.NewSemanticTokensOptionsRange(struct{}{}), `{}`),
		Entry("zero value", protocol.DefinitionRequestResult{}, `null`),
	)

	It("should return the value of the arm", func() {
		Expect(protocol.NewProgressToken("token").Value()).To(Equal("token"))
		Expect(protocol.ProgressToken{}.Value()).To(BeNil())
	})

	It("should leave out absent optional properties", func() {
		hover := protocol.Hover{Contents: protocol.NewHoverContents(protocol.MarkupContent{
			Kind:  protocol.MarkupKindPlainText,
			Value: "text",
		})}
		Expect(json.Marshal(hover)).To(MatchJSON(`{"contents":{"kind":"plaintext","value":"text"}}`))

		hover.Range = protocol.NewOptional(protocol.Range{})
		Expect(json.Marshal(hover)).To(MatchJSON(`{
			"contents": {"kind":"plaintext","value":"text"},
			"range": {"start":{"line":0,"character":0},"end":{"line":0,"character":0}}
		}`))
	})
})
