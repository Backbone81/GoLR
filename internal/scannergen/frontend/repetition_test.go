package frontend_test

import (
	"golr/internal/scannergen/frontend"
	"golr/internal/scannergen/frontend/dsl"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Repetition", func() {
	It("should convert to string with Any and a fixed repetition", func() {
		expression := frontend.Repetition{
			Minimum: 3,
			Maximum: 3,
			Child:   dsl.Any(),
		}
		Expect(expression.String()).To(Equal(".{3}"))
	})

	It("should convert to string with Any and a repetition range", func() {
		expression := frontend.Repetition{
			Minimum: 3,
			Maximum: 5,
			Child:   dsl.Any(),
		}
		Expect(expression.String()).To(Equal(".{3,5}"))
	})

	It("should convert to string with single character Literal and a fixed repetition", func() {
		expression := frontend.Repetition{
			Minimum: 3,
			Maximum: 3,
			Child:   dsl.Literal("a"),
		}
		Expect(expression.String()).To(Equal("a{3}"))
	})

	It("should convert to string with single character Literal and a repetition range", func() {
		expression := frontend.Repetition{
			Minimum: 3,
			Maximum: 5,
			Child:   dsl.Literal("a"),
		}
		Expect(expression.String()).To(Equal("a{3,5}"))
	})

	It("should convert to string with multi character Literal and a fixed repetition", func() {
		expression := frontend.Repetition{
			Minimum: 3,
			Maximum: 3,
			Child:   dsl.Literal("foo"),
		}
		Expect(expression.String()).To(Equal("(foo){3}"))
	})

	It("should convert to string with multi character Literal and a repetition range", func() {
		expression := frontend.Repetition{
			Minimum: 3,
			Maximum: 5,
			Child:   dsl.Literal("foo"),
		}
		Expect(expression.String()).To(Equal("(foo){3,5}"))
	})

	It("should convert to string with CharClass and a fixed repetition", func() {
		expression := frontend.Repetition{
			Minimum: 3,
			Maximum: 3,
			Child: dsl.CharClass(
				dsl.CharRange('a', 'z'),
			),
		}
		Expect(expression.String()).To(Equal("[a-z]{3}"))
	})

	It("should convert to string with CharClass and a repetition range", func() {
		expression := frontend.Repetition{
			Minimum: 3,
			Maximum: 5,
			Child: dsl.CharClass(
				dsl.CharRange('a', 'z'),
			),
		}
		Expect(expression.String()).To(Equal("[a-z]{3,5}"))
	})

	It("should provide the correct value for IsSingleNode", func() {
		expression := frontend.Repetition{}
		Expect(expression.IsSingleNode()).To(BeFalse())
	})

	It("should fail validation with the zero value", func() {
		expression := frontend.Repetition{}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation with a negative minimum", func() {
		expression := frontend.Repetition{
			Minimum: -1,
			Maximum: 3,
			Child:   dsl.Literal("a"),
		}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation with a negative maximum", func() {
		expression := frontend.Repetition{
			Minimum: 0,
			Maximum: -1,
			Child:   dsl.Literal("a"),
		}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation with a maximum below the minimum", func() {
		expression := frontend.Repetition{
			Minimum: 5,
			Maximum: 3,
			Child:   dsl.Literal("a"),
		}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation with a maximum and minimum to zero", func() {
		expression := frontend.Repetition{
			Minimum: 0,
			Maximum: 0,
			Child:   dsl.Literal("a"),
		}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should successfully validate", func() {
		expression := frontend.Repetition{
			Minimum: 1,
			Maximum: 3,
			Child:   dsl.Literal("a"),
		}
		Expect(expression.Validate()).To(Succeed())

		expression = frontend.Repetition{
			Minimum: 3,
			Maximum: 3,
			Child:   dsl.Literal("a"),
		}
		Expect(expression.Validate()).To(Succeed())
	})
})
