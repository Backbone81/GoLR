package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/utils"
)

var _ = Describe("C", func() {
	Context("CStringLiteral", func() {
		It("should quote a string which needs no escaping", func() {
			Expect(utils.CStringLiteral("end of input")).To(Equal(`"end of input"`))
			Expect(utils.CStringLiteral("")).To(Equal(`""`))
		})

		It("should escape the quote and the backslash", func() {
			Expect(utils.CStringLiteral(`";"`)).To(Equal(`"\";\""`))
			Expect(utils.CStringLiteral(`\`)).To(Equal(`"\\"`))
		})

		It("should escape control bytes", func() {
			Expect(utils.CStringLiteral("a\nb\tc\r")).To(Equal(`"a\nb\tc\r"`))
			// An octal escape takes at most three digits, so the digit after it stays a digit of its own.
			Expect(utils.CStringLiteral("\x001")).To(Equal(`"\0001"`))
			Expect(utils.CStringLiteral("\x7f")).To(Equal(`"\177"`))
		})

		It("should keep a trigraph from forming", func() {
			Expect(utils.CStringLiteral("??=")).To(Equal(`"?\?="`))
			Expect(utils.CStringLiteral("?")).To(Equal(`"?"`))
		})

		It("should keep UTF-8 as it is", func() {
			Expect(utils.CStringLiteral("größer")).To(Equal(`"größer"`))
		})
	})
})
