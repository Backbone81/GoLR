package frontend_test

import (
	"golr/internal/scannergen/frontend/dsl"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Optional", func() {
	It("should convert to string with Any", func() {
		expression := dsl.Optional(
			dsl.Any(),
		)
		Expect(expression.String()).To(Equal(".?"))
	})

	It("should convert to string with single character Literal", func() {
		expression := dsl.Optional(
			dsl.Literal("a"),
		)
		Expect(expression.String()).To(Equal("a?"))
	})

	It("should convert to string with multi character Literal", func() {
		expression := dsl.Optional(
			dsl.Literal("foo"),
		)
		Expect(expression.String()).To(Equal("(foo)?"))
	})

	It("should convert to string with CharClass", func() {
		expression := dsl.Optional(
			dsl.CharClass(
				dsl.CharRange('a', 'z'),
			),
		)
		Expect(expression.String()).To(Equal("[a-z]?"))
	})

	It("should provide the correct value for IsSingleNode", func() {
		expression := dsl.Optional(nil)
		Expect(expression.IsSingleNode()).To(BeFalse())
	})

	It("should fail validation for the zero value", func() {
		expression := dsl.Optional(nil)
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation for an invalid child", func() {
		expression := dsl.Optional(
			dsl.Literal(""),
		)
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should successfully validate", func() {
		expression := dsl.Optional(
			dsl.Literal("a"),
		)
		Expect(expression.Validate()).To(Succeed())
	})
})
