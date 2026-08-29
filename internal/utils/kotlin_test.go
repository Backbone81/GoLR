package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/utils"
)

var _ = Describe("Kotlin", func() {
	Context("KotlinIntType", func() {
		It("should pick the narrowest type which holds every value", func() {
			// Kotlin inherits the signed integer types of the virtual machine, so the boundaries are the positive
			// range of each and not its width. A table of 200 entries therefore needs a Short, where an unsigned byte
			// would have done.
			Expect(utils.KotlinIntType(0)).To(Equal("Byte"))
			Expect(utils.KotlinIntType(127)).To(Equal("Byte"))
			Expect(utils.KotlinIntType(128)).To(Equal("Short"))
			Expect(utils.KotlinIntType(32767)).To(Equal("Short"))
			Expect(utils.KotlinIntType(32768)).To(Equal("Int"))
		})
	})

	Context("NewKotlinIntArray", func() {
		It("should pick the type from the largest value it has to hold", func() {
			Expect(utils.NewKotlinIntArray([]int{0, 1, 127}).Type).To(Equal("Byte"))
			Expect(utils.NewKotlinIntArray([]int{0, 1, 128}).Type).To(Equal("Short"))
			Expect(utils.NewKotlinIntArray(nil).Type).To(Equal("Byte"))
		})
	})

	Context("KotlinConstantName", func() {
		It("should spell an identifier the way Kotlin spells a constant", func() {
			Expect(utils.KotlinConstantName("TransitionBase")).To(Equal("TRANSITION_BASE"))
			Expect(utils.KotlinConstantName("transitionBase")).To(Equal("TRANSITION_BASE"))
			Expect(utils.KotlinConstantName("Token")).To(Equal("TOKEN"))
			Expect(utils.KotlinConstantName("")).To(Equal(""))
		})
	})

	Context("KotlinString", func() {
		It("should quote a plain name", func() {
			Expect(utils.KotlinString("number")).To(Equal(`"number"`))
		})

		It("should escape a dollar sign, which Kotlin would read as a template expression", func() {
			// The augmented start symbol and the end of input terminal both carry one, so this is not a corner case
			// but every single grammar.
			Expect(utils.KotlinString("$accept")).To(Equal(`"\$accept"`))
			Expect(utils.KotlinString("$end")).To(Equal(`"\$end"`))
		})

		It("should escape what would end the literal", func() {
			Expect(utils.KotlinString(`a"b`)).To(Equal(`"a\"b"`))
			Expect(utils.KotlinString(`a\b`)).To(Equal(`"a\\b"`))
		})
	})

	Context("NewKotlinArray", func() {
		It("should spell the array of an entry type", func() {
			Expect(utils.NewKotlinArray("Byte").Type).To(Equal("ByteArray"))
			Expect(utils.NewKotlinArray("Byte").ArrayOf).To(Equal("byteArrayOf"))
			Expect(utils.NewKotlinArray("Short").Type).To(Equal("ShortArray"))
			Expect(utils.NewKotlinArray("Short").ArrayOf).To(Equal("shortArrayOf"))
			Expect(utils.NewKotlinArray("Int").Type).To(Equal("IntArray"))
			Expect(utils.NewKotlinArray("Int").ArrayOf).To(Equal("intArrayOf"))
		})

		It("should widen every entry which is not an Int already", func() {
			// Kotlin widens no integer type on its own, so a lookup which is compared with an index or added to one
			// carries the conversion. An array of Int needs none, and writing one there would be noise.
			Expect(utils.NewKotlinArray("Byte").ToInt).To(Equal(".toInt()"))
			Expect(utils.NewKotlinArray("Short").ToInt).To(Equal(".toInt()"))
			Expect(utils.NewKotlinArray("Int").ToInt).To(BeEmpty())
		})

		It("should narrow every Int written into an array which does not hold one", func() {
			// The counterpart of the widening above, for a value the generated code writes as an Int literal into an
			// array of a narrower type.
			Expect(utils.NewKotlinArray("Byte").FromInt).To(Equal(".toByte()"))
			Expect(utils.NewKotlinArray("Short").FromInt).To(Equal(".toShort()"))
			Expect(utils.NewKotlinArray("Int").FromInt).To(BeEmpty())
		})
	})

	Context("NewKotlinTable", func() {
		It("should name the property after the function", func() {
			result := utils.NewKotlinTable("transitionBase", utils.NewKotlinIntArray([]int{1, 2, 3}))

			Expect(result.Name).To(Equal("TRANSITION_BASE"))
			Expect(result.Function).To(Equal("transitionBase"))
			Expect(result.Type).To(Equal("ByteArray"))
			Expect(result.ArrayOf).To(Equal("byteArrayOf"))
		})

		It("should keep a table which fits in a single chunk", func() {
			result := utils.NewKotlinTable("transitionBase", utils.NewKotlinIntArray(make([]int, 4000)))

			Expect(result.Chunks).To(HaveLen(1))
			Expect(result.Chunks[0].Values).To(HaveLen(4000))
		})

		It("should split a table no single function could hold", func() {
			// The virtual machine limits a method to 64 KB of bytecode, which an array literal of much more than eight
			// thousand entries exceeds. A real grammar reaches that easily, so the split is what lets one be generated
			// at all.
			values := make([]int, 9001)
			for i := range values {
				values[i] = i % 128
			}

			result := utils.NewKotlinTable("transitionNext", utils.NewKotlinIntArray(values))

			Expect(result.Chunks).To(HaveLen(3))
			Expect(result.Chunks[0].Values).To(HaveLen(4000))
			Expect(result.Chunks[1].Values).To(HaveLen(4000))
			Expect(result.Chunks[2].Values).To(HaveLen(1001))

			// Every chunk is written as the type of the whole table, so the parts of one table can be joined.
			for _, chunk := range result.Chunks {
				Expect(chunk.Type).To(Equal("Byte"))
			}
		})

		It("should give an empty table the one function its property is built from", func() {
			result := utils.NewKotlinTable("transitionBase", utils.NewKotlinIntArray(nil))

			Expect(result.Chunks).To(HaveLen(1))
			Expect(result.Chunks[0].Values).To(BeEmpty())
		})

		It("should hold every value of the table in order", func() {
			values := make([]int, 8001)
			for i := range values {
				values[i] = i % 128
			}

			result := utils.NewKotlinTable("transitionNext", utils.NewKotlinIntArray(values))

			var joined []int
			for _, chunk := range result.Chunks {
				joined = append(joined, chunk.Values...)
			}
			Expect(joined).To(Equal(values))
		})
	})

	Context("NewKotlinNameTable", func() {
		It("should split the entries the same way a numeric table is split", func() {
			names := make([]string, 4001)
			for i := range names {
				names[i] = "INVALID_TOKEN"
			}

			result := utils.NewKotlinNameTable("acceptTokenByState", "Token", names)

			Expect(result.Name).To(Equal("ACCEPT_TOKEN_BY_STATE"))
			Expect(result.Type).To(Equal("Array<Token>"))
			Expect(result.ElementType).To(Equal("Token"))
			Expect(result.Chunks).To(HaveLen(2))
			Expect(result.Chunks[0]).To(HaveLen(4000))
			Expect(result.Chunks[1]).To(HaveLen(1))
		})

		It("should give an empty table the one function its property is built from", func() {
			result := utils.NewKotlinNameTable("acceptTokenByState", "Token", nil)

			Expect(result.Chunks).To(HaveLen(1))
			Expect(result.Chunks[0]).To(BeEmpty())
		})
	})
})
