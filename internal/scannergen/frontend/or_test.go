package frontend_test

import (
	"golr/internal/scannergen/frontend"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Or", func() {
	It("should convert to string", func() {
		expression := frontend.Or{
			Children: []frontend.Node{
				&frontend.Literal{
					Text: "a",
				},
				&frontend.Literal{
					Text: "b",
				},
			},
		}
		Expect(expression.String()).To(Equal("a|b"))
	})

	It("should provide the correct value for IsSingleNode", func() {
		expression := frontend.Or{
			Children: []frontend.Node{
				&frontend.Literal{Text: "a"},
			},
		}
		Expect(expression.IsSingleNode()).To(BeTrue())

		expression = frontend.Or{
			Children: []frontend.Node{
				&frontend.Literal{Text: "a"},
				&frontend.Literal{Text: "b"},
			},
		}
		Expect(expression.IsSingleNode()).To(BeFalse())
	})

	It("should fail validation for the zero value", func() {
		expression := frontend.Or{}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation for an invalid child", func() {
		expression := frontend.Or{
			Children: []frontend.Node{
				&frontend.Literal{Text: "a"},
				&frontend.Literal{},
			},
		}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation with only one child", func() {
		expression := frontend.Or{
			Children: []frontend.Node{
				&frontend.Literal{Text: "a"},
			},
		}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should successfully validate", func() {
		expression := frontend.Or{
			Children: []frontend.Node{
				&frontend.Literal{Text: "a"},
				&frontend.Literal{Text: "b"},
			},
		}
		Expect(expression.Validate()).To(Succeed())
	})
})
