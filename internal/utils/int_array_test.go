package utils_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/utils"
)

var _ = Describe("IntArray", func() {
	It("should pick the narrowest type which holds every value", func() {
		Expect(utils.NewIntArray([]int{0, 1, 255}).Type).To(Equal("uint8"))
		Expect(utils.NewIntArray([]int{0, 1, 256}).Type).To(Equal("uint16"))
		Expect(utils.NewIntArray([]int{65535}).Type).To(Equal("uint16"))
		Expect(utils.NewIntArray([]int{65536}).Type).To(Equal("uint32"))
	})

	It("should pick a type for an empty table", func() {
		Expect(utils.NewIntArray(nil).Type).To(Equal("uint8"))
	})

	It("should keep the type it is given", func() {
		// Two tables which are compared against each other at runtime have to share a type, even where the values of
		// one of them would fit into a narrower one.
		result := utils.NewTypedIntArray("uint16", []int{1, 2, 3})
		Expect(result.Type).To(Equal("uint16"))
		Expect(result.Values).To(Equal([]int{1, 2, 3}))
	})

	It("should write every value into the literal", func() {
		result := utils.NewIntArray([]int{1, 2, 3}).Literal()
		Expect(strings.Fields(result)).To(Equal([]string{"1,", "2,", "3,"}))
	})

	It("should wrap the literal over several lines", func() {
		values := make([]int, 40)
		result := utils.NewIntArray(values).Literal()

		// The wrapping is what keeps a table of thousands of entries readable in the generated file.
		Expect(strings.Count(result, "\n")).To(BeNumerically(">", 1))
		Expect(strings.Fields(result)).To(HaveLen(len(values)))
	})

	It("should provide an empty literal for an empty table", func() {
		Expect(strings.Fields(utils.NewIntArray(nil).Literal())).To(BeEmpty())
	})
})
