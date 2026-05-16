package frontend_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/scannergen/frontend/dsl"
)

var _ = Describe("CharClass", func() {
	It("should convert to string with a single character", func() {
		expression := dsl.CharClass(
			dsl.CharRange('a', 'a'),
		)
		Expect(expression.String()).To(Equal("[a]"))
	})

	It("should convert to string with a single character negated", func() {
		expression := dsl.NegCharClass(
			dsl.CharRange('a', 'a'),
		)
		Expect(expression.String()).To(Equal("[^a]"))
	})

	It("should convert to string with two characters", func() {
		expression := dsl.CharClass(
			dsl.CharRange('a', 'a'),
			dsl.CharRange('b', 'b'),
		)
		Expect(expression.String()).To(Equal("[ab]"))
	})

	It("should convert to string with two characters negated", func() {
		expression := dsl.NegCharClass(
			dsl.CharRange('a', 'a'),
			dsl.CharRange('b', 'b'),
		)
		Expect(expression.String()).To(Equal("[^ab]"))
	})

	It("should convert to string with a single character range", func() {
		expression := dsl.CharClass(
			dsl.CharRange('a', 'z'),
		)
		Expect(expression.String()).To(Equal("[a-z]"))
	})

	It("should convert to string with a single character range negated", func() {
		expression := dsl.NegCharClass(
			dsl.CharRange('a', 'z'),
		)
		Expect(expression.String()).To(Equal("[^a-z]"))
	})

	It("should convert to string with two character ranges", func() {
		expression := dsl.CharClass(
			dsl.CharRange('a', 'z'),
			dsl.CharRange('0', '9'),
		)
		Expect(expression.String()).To(Equal("[a-z0-9]"))
	})

	It("should convert to string with two character ranges negated", func() {
		expression := dsl.NegCharClass(
			dsl.CharRange('a', 'z'),
			dsl.CharRange('0', '9'),
		)
		Expect(expression.String()).To(Equal("[^a-z0-9]"))
	})

	It("should provide the correct value for IsSingleNode", func() {
		expression := dsl.CharClass()
		Expect(expression.IsSingleNode()).To(BeTrue())
	})

	It("should succeed validation with zero value", func() {
		expression := dsl.CharClass()
		Expect(expression.Validate()).To(Succeed())
	})

	It("should fail validation with invalid character range", func() {
		expression := dsl.CharClass(
			dsl.CharRange(-1, 'a'),
		)
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should validate successfully", func() {
		expression := dsl.CharClass(
			dsl.CharRange('a', 'a'),
		)
		Expect(expression.Validate()).To(Succeed())
	})
})
