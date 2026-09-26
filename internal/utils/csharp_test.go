package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/utils"
)

var _ = Describe("CSharp", func() {
	Context("CSharpStringLiteral", func() {
		It("should quote a string which needs no escaping", func() {
			Expect(utils.CSharpStringLiteral("end of input")).To(Equal(`"end of input"`))
			Expect(utils.CSharpStringLiteral("")).To(Equal(`""`))
		})

		It("should escape the quote and the backslash", func() {
			Expect(utils.CSharpStringLiteral(`";"`)).To(Equal(`"\";\""`))
			Expect(utils.CSharpStringLiteral(`\`)).To(Equal(`"\\"`))
		})

		It("should escape control characters", func() {
			Expect(utils.CSharpStringLiteral("a\nb\tc\r")).To(Equal(`"a\nb\tc\r"`))
			// A \u escape takes exactly four digits, so the hex digit after it stays a character of its own.
			Expect(utils.CSharpStringLiteral("\x00a")).To(Equal(`"\u0000a"`))
			Expect(utils.CSharpStringLiteral("\x7f")).To(Equal(`"\u007f"`))
		})

		It("should keep UTF-8 as it is", func() {
			Expect(utils.CSharpStringLiteral("größer")).To(Equal(`"größer"`))
		})
	})
})
