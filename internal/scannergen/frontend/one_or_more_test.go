package frontend_test

import (
	"golr/internal/scannergen/frontend"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("OneOrMore", func() {
	It("should convert to string with Any", func() {
		expression := frontend.OneOrMore{
			Child: &frontend.Node{
				Kind: frontend.KindAny,
			},
		}
		Expect(expression.String()).To(Equal(".+"))
	})

	It("should convert to string with single character Literal", func() {
		expression := frontend.OneOrMore{
			Child: frontend.NewNodeLiteral("a"),
		}
		Expect(expression.String()).To(Equal("a+"))
	})

	It("should convert to string with multi character Literal", func() {
		expression := frontend.OneOrMore{
			Child: frontend.NewNodeLiteral("foo"),
		}
		Expect(expression.String()).To(Equal("(foo)+"))
	})

	It("should convert to string with CharClass", func() {
		expression := frontend.OneOrMore{
			Child: frontend.NewNodeCharClass(frontend.CharClass{
				Ranges: []frontend.CharRange{
					{
						Low:  'a',
						High: 'z',
					},
				},
			}),
		}
		Expect(expression.String()).To(Equal("[a-z]+"))
	})

	It("should provide the correct value for IsSingleNode", func() {
		expression := frontend.OneOrMore{}
		Expect(expression.IsSingleNode()).To(BeFalse())
	})

	It("should fail validation for the zero value", func() {
		expression := frontend.OneOrMore{}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation for invalid child", func() {
		expression := frontend.OneOrMore{
			Child: frontend.NewNodeLiteral(""),
		}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should successfully validate", func() {
		expression := frontend.OneOrMore{
			Child: frontend.NewNodeLiteral("a"),
		}
		Expect(expression.Validate()).To(Succeed())
	})
})
