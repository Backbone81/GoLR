package frontend_test

import (
	"golr/internal/scannergen/frontend"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Literal", func() {
	It("should convert to string", func() {
		expression := frontend.Literal{
			Text: "foo",
		}
		Expect(expression.String()).To(Equal("foo"))
	})

	It("should provide the correct value for IsSingleNode", func() {
		expression := frontend.Literal{
			Text: "a",
		}
		Expect(expression.IsSingleNode()).To(BeTrue())

		expression = frontend.Literal{
			Text: "ab",
		}
		Expect(expression.IsSingleNode()).To(BeFalse())
	})

	It("should fail validation with the zero value", func() {
		expression := frontend.Literal{}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation with the empty string", func() {
		expression := frontend.Literal{Text: ""}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should successfully validate", func() {
		expression := frontend.Literal{Text: "a"}
		Expect(expression.Validate()).To(Succeed())
	})
})
