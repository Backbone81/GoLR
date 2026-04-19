package frontend_test

import (
	"golr/internal/scannergen/frontend"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Optional", func() {
	It("should convert to string with Any", func() {
		expression := frontend.Optional{
			Child: &frontend.Any{},
		}
		Expect(expression.String()).To(Equal(".?"))
	})

	It("should convert to string with single character Literal", func() {
		expression := frontend.Optional{
			Child: &frontend.Literal{
				Text: "a",
			},
		}
		Expect(expression.String()).To(Equal("a?"))
	})

	It("should convert to string with multi character Literal", func() {
		expression := frontend.Optional{
			Child: &frontend.Literal{
				Text: "foo",
			},
		}
		Expect(expression.String()).To(Equal("(foo)?"))
	})

	It("should convert to string with CharClass", func() {
		expression := frontend.Optional{
			Child: &frontend.CharClass{
				Ranges: []frontend.CharRange{
					{
						Low:  'a',
						High: 'z',
					},
				},
			},
		}
		Expect(expression.String()).To(Equal("[a-z]?"))
	})

	It("should provide the correct value for IsSingleNode", func() {
		expression := frontend.Optional{}
		Expect(expression.IsSingleNode()).To(BeFalse())
	})

	It("should fail validation for the zero value", func() {
		expression := frontend.Optional{}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation for an invalid child", func() {
		expression := frontend.Optional{
			Child: &frontend.Literal{},
		}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should successfully validate", func() {
		expression := frontend.Optional{
			Child: &frontend.Literal{Text: "a"},
		}
		Expect(expression.Validate()).To(Succeed())
	})
})
