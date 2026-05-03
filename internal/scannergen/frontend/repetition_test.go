package frontend_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"golr/internal/scannergen/frontend/dsl"
)

var _ = Describe("Repetition", func() {
	It("should convert to string with Any and a fixed repetition", func() {
		expression := dsl.Repetition(
			dsl.Any(),
			3,
			3,
		)
		Expect(expression.String()).To(Equal(".{3}"))
	})

	It("should convert to string with Any and a repetition range", func() {
		expression := dsl.Repetition(
			dsl.Any(),
			3,
			5,
		)
		Expect(expression.String()).To(Equal(".{3,5}"))
	})

	It("should convert to string with single character Literal and a fixed repetition", func() {
		expression := dsl.Repetition(
			dsl.Literal("a"),
			3,
			3,
		)
		Expect(expression.String()).To(Equal("a{3}"))
	})

	It("should convert to string with single character Literal and a repetition range", func() {
		expression := dsl.Repetition(
			dsl.Literal("a"),
			3,
			5,
		)
		Expect(expression.String()).To(Equal("a{3,5}"))
	})

	It("should convert to string with multi character Literal and a fixed repetition", func() {
		expression := dsl.Repetition(
			dsl.Literal("foo"),
			3,
			3,
		)
		Expect(expression.String()).To(Equal("(foo){3}"))
	})

	It("should convert to string with multi character Literal and a repetition range", func() {
		expression := dsl.Repetition(
			dsl.Literal("foo"),
			3,
			5,
		)
		Expect(expression.String()).To(Equal("(foo){3,5}"))
	})

	It("should convert to string with CharClass and a fixed repetition", func() {
		expression := dsl.Repetition(
			dsl.CharClass(
				dsl.CharRange('a', 'z'),
			),
			3,
			3,
		)
		Expect(expression.String()).To(Equal("[a-z]{3}"))
	})

	It("should convert to string with CharClass and a repetition range", func() {
		expression := dsl.Repetition(
			dsl.CharClass(
				dsl.CharRange('a', 'z'),
			),
			3,
			5,
		)
		Expect(expression.String()).To(Equal("[a-z]{3,5}"))
	})

	It("should provide the correct value for IsSingleNode", func() {
		expression := dsl.Repetition(nil, 0, 0)
		Expect(expression.IsSingleNode()).To(BeFalse())
	})

	It("should fail validation with the zero value", func() {
		expression := dsl.Repetition(nil, 0, 0)
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation with a negative minimum", func() {
		expression := dsl.Repetition(
			dsl.Literal("a"),
			-1,
			3,
		)
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation with a negative maximum", func() {
		expression := dsl.Repetition(
			dsl.Literal("a"),
			0,
			-1,
		)
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation with a maximum below the minimum", func() {
		expression := dsl.Repetition(
			dsl.Literal("a"),
			5,
			3,
		)
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation with a maximum and minimum to zero", func() {
		expression := dsl.Repetition(
			dsl.Literal("a"),
			0,
			0,
		)
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should successfully validate", func() {
		expression := dsl.Repetition(
			dsl.Literal("a"),
			1,
			3,
		)
		Expect(expression.Validate()).To(Succeed())

		expression = dsl.Repetition(
			dsl.Literal("a"),
			3,
			3,
		)
		Expect(expression.Validate()).To(Succeed())
	})
})
