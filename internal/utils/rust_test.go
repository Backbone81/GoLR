package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/utils"
)

var _ = Describe("Rust", func() {
	Context("RustStringLiteral", func() {
		It("should quote a string which needs no escaping", func() {
			Expect(utils.RustStringLiteral("end of input")).To(Equal(`"end of input"`))
			Expect(utils.RustStringLiteral("")).To(Equal(`""`))
		})

		It("should escape the quote and the backslash", func() {
			Expect(utils.RustStringLiteral(`";"`)).To(Equal(`"\";\""`))
			Expect(utils.RustStringLiteral(`\`)).To(Equal(`"\\"`))
		})

		It("should escape control characters", func() {
			Expect(utils.RustStringLiteral("a\nb\tc\r")).To(Equal(`"a\nb\tc\r"`))
			// A \x escape takes exactly two digits, so the hex digit after it stays a character of its own.
			Expect(utils.RustStringLiteral("\x00a")).To(Equal(`"\x00a"`))
			Expect(utils.RustStringLiteral("\x7f")).To(Equal(`"\x7f"`))
		})

		It("should keep UTF-8 as it is", func() {
			Expect(utils.RustStringLiteral("größer")).To(Equal(`"größer"`))
		})
	})
})
