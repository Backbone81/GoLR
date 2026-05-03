package frontend_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"golr/internal/scannergen/frontend/dsl"
)

var _ = Describe("OneOrMore", func() {
	It("should convert to string with Any", func() {
		expression := dsl.OneOrMore(
			dsl.Any(),
		)
		Expect(expression.String()).To(Equal(".+"))
	})

	It("should convert to string with single character Literal", func() {
		expression := dsl.OneOrMore(
			dsl.Literal("a"),
		)
		Expect(expression.String()).To(Equal("a+"))
	})

	It("should convert to string with multi character Literal", func() {
		expression := dsl.OneOrMore(
			dsl.Literal("foo"),
		)
		Expect(expression.String()).To(Equal("(foo)+"))
	})

	It("should convert to string with CharClass", func() {
		expression := dsl.OneOrMore(
			dsl.CharClass(
				dsl.CharRange('a', 'z'),
			),
		)
		Expect(expression.String()).To(Equal("[a-z]+"))
	})

	It("should provide the correct value for IsSingleNode", func() {
		expression := dsl.OneOrMore(nil)
		Expect(expression.IsSingleNode()).To(BeFalse())
	})

	It("should fail validation for the zero value", func() {
		expression := dsl.OneOrMore(nil)
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation for invalid child", func() {
		expression := dsl.OneOrMore(
			dsl.Literal(""),
		)
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should successfully validate", func() {
		expression := dsl.OneOrMore(
			dsl.Literal("a"),
		)
		Expect(expression.Validate()).To(Succeed())
	})
})
