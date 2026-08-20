package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/utils"
)

var _ = Describe("Python", func() {
	Context("PythonConstantName", func() {
		It("should spell an identifier the way Python spells a constant", func() {
			Expect(utils.PythonConstantName("TransitionBase")).To(Equal("TRANSITION_BASE"))
			Expect(utils.PythonConstantName("transitionBase")).To(Equal("TRANSITION_BASE"))
			Expect(utils.PythonConstantName("Token")).To(Equal("TOKEN"))
			Expect(utils.PythonConstantName("")).To(Equal(""))
		})

		It("should treat a digit as the end of a word", func() {
			Expect(utils.PythonConstantName("TokenUtf8Literal")).To(Equal("TOKEN_UTF8_LITERAL"))
		})
	})

	Context("NewPythonIntArray", func() {
		It("should leave the entries untyped, because Python has nothing to narrow", func() {
			Expect(utils.NewPythonIntArray([]int{0, 1, 300}).Type).To(BeEmpty())
			Expect(utils.NewPythonIntArray([]int{0, 1, 300}).Values).To(Equal([]int{0, 1, 300}))
		})
	})
})
