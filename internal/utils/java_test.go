package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/utils"
)

var _ = Describe("Java", func() {
	Context("JavaIntType", func() {
		It("should pick the narrowest type which holds every value", func() {
			// Every Java integer type is signed, so the boundaries are the positive range of each and not its
			// width. A table of 200 entries therefore needs a short, where an unsigned byte would have done.
			Expect(utils.JavaIntType(0)).To(Equal("byte"))
			Expect(utils.JavaIntType(127)).To(Equal("byte"))
			Expect(utils.JavaIntType(128)).To(Equal("short"))
			Expect(utils.JavaIntType(32767)).To(Equal("short"))
			Expect(utils.JavaIntType(32768)).To(Equal("int"))
		})
	})

	Context("NewJavaIntArray", func() {
		It("should pick the type from the largest value it has to hold", func() {
			Expect(utils.NewJavaIntArray([]int{0, 1, 127}).Type).To(Equal("byte"))
			Expect(utils.NewJavaIntArray([]int{0, 1, 128}).Type).To(Equal("short"))
			Expect(utils.NewJavaIntArray(nil).Type).To(Equal("byte"))
		})
	})

	Context("JavaConstantName", func() {
		It("should spell an identifier the way Java spells a constant", func() {
			Expect(utils.JavaConstantName("TransitionBase")).To(Equal("TRANSITION_BASE"))
			Expect(utils.JavaConstantName("transitionBase")).To(Equal("TRANSITION_BASE"))
			Expect(utils.JavaConstantName("Token")).To(Equal("TOKEN"))
			Expect(utils.JavaConstantName("")).To(Equal(""))
		})

		It("should treat a digit as the end of a word", func() {
			Expect(utils.JavaConstantName("TokenUtf8Literal")).To(Equal("TOKEN_UTF8_LITERAL"))
		})
	})

	Context("NewJavaTable", func() {
		It("should name the constant after the method", func() {
			result := utils.NewJavaTable("transitionBase", utils.NewJavaIntArray([]int{1, 2, 3}))

			Expect(result.Name).To(Equal("TRANSITION_BASE"))
			Expect(result.Method).To(Equal("transitionBase"))
			Expect(result.Type).To(Equal("byte"))
		})

		It("should keep a table which fits in a single chunk", func() {
			result := utils.NewJavaTable("transitionBase", utils.NewJavaIntArray(make([]int, 4000)))

			Expect(result.Chunks).To(HaveLen(1))
			Expect(result.Chunks[0].Values).To(HaveLen(4000))
		})

		It("should split a table no single method could hold", func() {
			// The virtual machine limits a method to 64 KB of bytecode, which an array literal of much more than
			// eight thousand entries exceeds. A real grammar reaches that easily, so the split is what lets one be
			// generated at all.
			values := make([]int, 9001)
			for i := range values {
				values[i] = i % 128
			}

			result := utils.NewJavaTable("transitionNext", utils.NewJavaIntArray(values))

			Expect(result.Chunks).To(HaveLen(3))
			Expect(result.Chunks[0].Values).To(HaveLen(4000))
			Expect(result.Chunks[1].Values).To(HaveLen(4000))
			Expect(result.Chunks[2].Values).To(HaveLen(1001))

			// Every chunk is written as the type of the whole table, so the parts of one table can be joined.
			for _, chunk := range result.Chunks {
				Expect(chunk.Type).To(Equal(result.Type))
			}
		})

		It("should give an empty table the one method its constant is built from", func() {
			result := utils.NewJavaTable("transitionBase", utils.NewJavaIntArray(nil))

			Expect(result.Chunks).To(HaveLen(1))
			Expect(result.Chunks[0].Values).To(BeEmpty())
		})

		It("should hold every value of the table in order", func() {
			values := make([]int, 8001)
			for i := range values {
				values[i] = i % 128
			}

			result := utils.NewJavaTable("transitionNext", utils.NewJavaIntArray(values))

			var joined []int
			for _, chunk := range result.Chunks {
				joined = append(joined, chunk.Values...)
			}
			Expect(joined).To(Equal(values))
		})
	})

	Context("NewJavaNameTable", func() {
		It("should split the entries the same way a numeric table is split", func() {
			names := make([]string, 4001)
			for i := range names {
				names[i] = "INVALID_TOKEN"
			}

			result := utils.NewJavaNameTable("acceptTokenByState", "Token", names)

			Expect(result.Name).To(Equal("ACCEPT_TOKEN_BY_STATE"))
			Expect(result.Type).To(Equal("Token"))
			Expect(result.Chunks).To(HaveLen(2))
			Expect(result.Chunks[0]).To(HaveLen(4000))
			Expect(result.Chunks[1]).To(HaveLen(1))
		})

		It("should give an empty table the one method its constant is built from", func() {
			result := utils.NewJavaNameTable("acceptTokenByState", "Token", nil)

			Expect(result.Chunks).To(HaveLen(1))
			Expect(result.Chunks[0]).To(BeEmpty())
		})
	})
})
