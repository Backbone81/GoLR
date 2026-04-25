package frontend_test

import (
	"golr/internal/scannergen/frontend"
	"golr/internal/scannergen/frontend/dsl"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Or", func() {
	It("should convert to string", func() {
		expression := frontend.Or{
			Children: []*frontend.Node{
				dsl.Literal("a"),
				dsl.Literal("b"),
			},
		}
		Expect(expression.String()).To(Equal("a|b"))
	})

	It("should provide the correct value for IsSingleNode", func() {
		expression := frontend.Or{
			Children: []*frontend.Node{
				dsl.Literal("a"),
			},
		}
		Expect(expression.IsSingleNode()).To(BeTrue())

		expression = frontend.Or{
			Children: []*frontend.Node{
				dsl.Literal("a"),
				dsl.Literal("b"),
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
			Children: []*frontend.Node{
				dsl.Literal("a"),
				dsl.Literal(""),
			},
		}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation with only one child", func() {
		expression := frontend.Or{
			Children: []*frontend.Node{
				dsl.Literal("a"),
			},
		}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should successfully validate", func() {
		expression := frontend.Or{
			Children: []*frontend.Node{
				dsl.Literal("a"),
				dsl.Literal("b"),
			},
		}
		Expect(expression.Validate()).To(Succeed())
	})
})
