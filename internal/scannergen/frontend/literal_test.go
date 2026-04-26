package frontend_test

import (
	"golr/internal/scannergen/frontend"
	"golr/internal/scannergen/frontend/dsl"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Literal", func() {
	It("should convert to string", func() {
		expression := dsl.Literal("foo")
		Expect(expression.String()).To(Equal("foo"))
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
