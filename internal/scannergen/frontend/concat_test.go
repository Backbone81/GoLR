package frontend_test

import (
	"golr/internal/scannergen/frontend"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Concat", func() {
	It("should convert to string", func() {
		expression := frontend.Concat{
			Children: []*frontend.Node{
				frontend.NewNodeLiteral("a"),
				frontend.NewNodeLiteral("b"),
			},
		}
		Expect(expression.String()).To(Equal("ab"))
	})

	It("should provide the correct value for IsSingleNode", func() {
		expression := frontend.Concat{}
		Expect(expression.IsSingleNode()).To(BeFalse())

		expression = frontend.Concat{
			Children: []*frontend.Node{
				frontend.NewNodeLiteral("a"),
			},
		}
		Expect(expression.IsSingleNode()).To(BeTrue())

		expression = frontend.Concat{
			Children: []*frontend.Node{
				frontend.NewNodeLiteral("a"),
				frontend.NewNodeLiteral("b"),
			},
		}
		Expect(expression.IsSingleNode()).To(BeFalse())
	})

	It("should fail validation for the zero value", func() {
		expression := frontend.Concat{}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation with invalid child", func() {
		expression := frontend.Concat{
			Children: []*frontend.Node{
				frontend.NewNodeLiteral("a"),
				frontend.NewNodeLiteral(""),
			},
		}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation with only one child", func() {
		expression := frontend.Concat{
			Children: []*frontend.Node{
				frontend.NewNodeLiteral("a"),
			},
		}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should successfully validate", func() {
		expression := frontend.Concat{
			Children: []*frontend.Node{
				frontend.NewNodeLiteral("a"),
				frontend.NewNodeLiteral("b"),
			},
		}
		Expect(expression.Validate()).To(Succeed())
	})
})
