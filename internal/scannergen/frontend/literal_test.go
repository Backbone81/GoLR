package frontend_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/scannergen/frontend"
	"github.com/backbone81/golr/internal/scannergen/frontend/dsl"
)

var _ = Describe("Literal", func() {
	It("should convert to string", func() {
		expression := dsl.Literal("foo")
		Expect(expression.String()).To(Equal("foo"))
	})

	It("should escape regex metacharacters in string output", func() {
		expression := dsl.Literal(".")
		Expect(expression.String()).To(Equal(`\.`))

		expression = dsl.Literal("a.b")
		Expect(expression.String()).To(Equal(`a\.b`))

		expression = dsl.Literal("+")
		Expect(expression.String()).To(Equal(`\+`))
	})

	It("should escape control characters in string output", func() {
		expression := dsl.Literal("\t")
		Expect(expression.String()).To(Equal(`\t`))

		expression = dsl.Literal("\n")
		Expect(expression.String()).To(Equal(`\n`))
	})

	It("should provide the correct value for IsSingleNode", func() {
		expression := dsl.Literal("a")
		Expect(expression.IsSingleNode()).To(BeTrue())

		expression = dsl.Literal("ab")
		Expect(expression.IsSingleNode()).To(BeFalse())
	})

	It("should fail validation with the zero value", func() {
		expression := frontend.Literal{}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation with the empty string", func() {
		expression := dsl.Literal("")
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should successfully validate", func() {
		expression := dsl.Literal("a")
		Expect(expression.Validate()).To(Succeed())
	})
})
