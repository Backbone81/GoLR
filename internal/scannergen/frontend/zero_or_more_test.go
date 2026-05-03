package frontend_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"golr/internal/scannergen/frontend/dsl"
)

var _ = Describe("ZeroOrMore", func() {
	It("should convert to string with Any", func() {
		expression := dsl.ZeroOrMore(
			dsl.Any(),
		)
		Expect(expression.String()).To(Equal(".*"))
	})

	It("should convert to string with single character Literal", func() {
		expression := dsl.ZeroOrMore(
			dsl.Literal("a"),
		)
		Expect(expression.String()).To(Equal("a*"))
	})

	It("should convert to string with multi character Literal", func() {
		expression := dsl.ZeroOrMore(
			dsl.Literal("foo"),
		)
		Expect(expression.String()).To(Equal("(foo)*"))
	})

	It("should convert to string with CharClass", func() {
		expression := dsl.ZeroOrMore(
			dsl.CharClass(
				dsl.CharRange('a', 'z'),
			),
		)
		Expect(expression.String()).To(Equal("[a-z]*"))
	})

	It("should provide the correct value for IsSingleNode", func() {
		expression := dsl.ZeroOrMore(nil)
		Expect(expression.IsSingleNode()).To(BeFalse())
	})

	It("should fail validation with the zero value", func() {
		expression := dsl.ZeroOrMore(nil)
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation with an invalid child", func() {
		expression := dsl.ZeroOrMore(
			dsl.Literal(""),
		)
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should successfully validate", func() {
		expression := dsl.ZeroOrMore(
			dsl.Literal("a"),
		)
		Expect(expression.Validate()).To(Succeed())
	})
})
