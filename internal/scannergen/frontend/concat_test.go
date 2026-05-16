package frontend_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/scannergen/frontend/dsl"
)

var _ = Describe("Concat", func() {
	It("should convert to string", func() {
		expression := dsl.Concat(
			dsl.Literal("a"),
			dsl.Literal("b"),
		)
		Expect(expression.String()).To(Equal("ab"))
	})

	It("should provide the correct value for IsSingleNode", func() {
		expression := dsl.Concat()
		Expect(expression.IsSingleNode()).To(BeFalse())

		expression = dsl.Concat(
			dsl.Literal("a"),
		)
		Expect(expression.IsSingleNode()).To(BeTrue())

		expression = dsl.Concat(
			dsl.Literal("a"),
			dsl.Literal("b"),
		)
		Expect(expression.IsSingleNode()).To(BeFalse())
	})

	It("should fail validation for the zero value", func() {
		expression := dsl.Concat()
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation with invalid child", func() {
		expression := dsl.Concat(
			dsl.Literal("a"),
			dsl.Literal(""),
		)
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation with only one child", func() {
		expression := dsl.Concat(
			dsl.Literal("a"),
		)
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should successfully validate", func() {
		expression := dsl.Concat(
			dsl.Literal("a"),
			dsl.Literal("b"),
		)
		Expect(expression.Validate()).To(Succeed())
	})
})
