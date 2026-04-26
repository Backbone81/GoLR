package frontend_test

import (
	"golr/internal/scannergen/frontend"
	"golr/internal/scannergen/frontend/dsl"
	"unicode"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CharRange", func() {
	It("should convert to string with a single character", func() {
		expression := dsl.CharRange('a', 'a')
		Expect(expression.String()).To(Equal("a"))
	})

	It("should convert to string with a character range", func() {
		expression := dsl.CharRange('a', 'z')
		Expect(expression.String()).To(Equal("a-z"))
	})

	It("should convert to string with special characters", func() {
		expression := dsl.CharRange(' ', '\t')
		Expect(expression.String()).To(Equal("' '-\\t"))

		expression = dsl.CharRange('\r', '\n')
		Expect(expression.String()).To(Equal("\\r-\\n"))

		expression = dsl.CharRange('\u2116', '\u26a0')
		Expect(expression.String()).To(Equal("0x2116-0x26a0"))
	})

	It("should fail validation with out of range runes", func() {
		expression := dsl.CharRange(-1, 'c')
		Expect(expression.Validate()).ToNot(Succeed())

		expression = dsl.CharRange(unicode.MaxRune+1, 'c')
		Expect(expression.Validate()).ToNot(Succeed())

		expression = dsl.CharRange('a', -1)
		Expect(expression.Validate()).ToNot(Succeed())

		expression = dsl.CharRange('a', unicode.MaxRune+1)
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation when high is lower than low", func() {
		expression := dsl.CharRange('c', 'a')
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should validate successfully", func() {
		expression := dsl.CharRange('a', 'a')
		Expect(expression.Validate()).To(Succeed())

		expression = dsl.CharRange('a', 'b')
		Expect(expression.Validate()).To(Succeed())
	})

	Context("SortCharRanges", func() {
		It("should sort ascending", func() {
			characterRanges := []frontend.CharRange{
				dsl.CharRange('x', 'y'),
				dsl.CharRange('a', 'b'),
			}
			frontend.SortCharRanges(characterRanges)
			Expect(characterRanges).To(Equal([]frontend.CharRange{
				dsl.CharRange('a', 'b'),
				dsl.CharRange('x', 'y'),
			}))
		})

		It("should not change the order of already sorted character range", func() {
			characterRanges := []frontend.CharRange{
				dsl.CharRange('a', 'b'),
				dsl.CharRange('x', 'y'),
			}
			frontend.SortCharRanges(characterRanges)
			Expect(characterRanges).To(Equal([]frontend.CharRange{
				dsl.CharRange('a', 'b'),
				dsl.CharRange('x', 'y'),
			}))
		})

		It("should correctly deal with character ranges starting with the same rune", func() {
			characterRanges := []frontend.CharRange{
				dsl.CharRange('a', 'b'),
				dsl.CharRange('a', 'z'),
			}
			frontend.SortCharRanges(characterRanges)
			Expect(characterRanges).To(Equal([]frontend.CharRange{
				dsl.CharRange('a', 'b'),
				dsl.CharRange('a', 'z'),
			}))
		})
	})

	Context("SplitCharRanges", func() {
		It("should split in the middle", func() {
			characterRanges := []frontend.CharRange{
				dsl.CharRange('a', 'z'),
			}
			splitCharacterRanges := frontend.SplitCharRanges(characterRanges, 'f')
			Expect(splitCharacterRanges).To(Equal([]frontend.CharRange{
				dsl.CharRange('a', 'e'),
				dsl.CharRange('f', 'z'),
			}))
		})

		It("should not split on the start", func() {
			characterRanges := []frontend.CharRange{
				dsl.CharRange('a', 'z'),
			}
			splitCharacterRanges := frontend.SplitCharRanges(characterRanges, 'a')
			Expect(splitCharacterRanges).To(Equal([]frontend.CharRange{
				dsl.CharRange('a', 'z'),
			}))
		})

		It("should split on the end", func() {
			characterRanges := []frontend.CharRange{
				dsl.CharRange('a', 'z'),
			}
			splitCharacterRanges := frontend.SplitCharRanges(characterRanges, 'z')
			Expect(splitCharacterRanges).To(Equal([]frontend.CharRange{
				dsl.CharRange('a', 'y'),
				dsl.CharRange('z', 'z'),
			}))
		})

		It("should not split out of range at the end", func() {
			characterRanges := []frontend.CharRange{
				dsl.CharRange('a', 'f'),
			}
			splitCharacterRanges := frontend.SplitCharRanges(characterRanges, 'u')
			Expect(splitCharacterRanges).To(Equal([]frontend.CharRange{
				dsl.CharRange('a', 'f'),
			}))
		})

		It("should not split out of range at the start", func() {
			characterRanges := []frontend.CharRange{
				dsl.CharRange('u', 'w'),
			}
			splitCharacterRanges := frontend.SplitCharRanges(characterRanges, 'a')
			Expect(splitCharacterRanges).To(Equal([]frontend.CharRange{
				dsl.CharRange('u', 'w'),
			}))
		})

		It("should not split single character ranges", func() {
			characterRanges := []frontend.CharRange{
				dsl.CharRange('b', 'b'),
			}
			splitCharacterRanges := frontend.SplitCharRanges(characterRanges, 'b')
			Expect(splitCharacterRanges).To(Equal([]frontend.CharRange{
				dsl.CharRange('b', 'b'),
			}))
		})
	})

	Context("RemoveCharRanges", func() {
		It("should remove an exact match", func() {
			characterRanges := []frontend.CharRange{
				dsl.CharRange('a', 'c'),
				dsl.CharRange('u', 'v'),
				dsl.CharRange('x', 'z'),
			}
			Expect(frontend.RemoveCharRanges(characterRanges, dsl.CharRange('u', 'v'))).To(
				Equal([]frontend.CharRange{
					dsl.CharRange('a', 'c'),
					dsl.CharRange('x', 'z'),
				}),
			)
		})

		It("should remove on overlap", func() {
			characterRanges := []frontend.CharRange{
				dsl.CharRange('a', 'c'),
				dsl.CharRange('u', 'v'),
				dsl.CharRange('x', 'z'),
			}
			Expect(frontend.RemoveCharRanges(characterRanges, dsl.CharRange('t', 'w'))).To(Equal([]frontend.CharRange{
				dsl.CharRange('a', 'c'),
				dsl.CharRange('x', 'z'),
			}))
		})

		It("should remove multiples in range", func() {
			characterRanges := []frontend.CharRange{
				dsl.CharRange('b', 'c'),
				dsl.CharRange('u', 'v'),
				dsl.CharRange('x', 'z'),
			}
			Expect(frontend.RemoveCharRanges(characterRanges, dsl.CharRange('a', 'w'))).To(Equal([]frontend.CharRange{
				dsl.CharRange('x', 'z'),
			}))
		})

		It("should not remove on partial match", func() {
			characterRanges := []frontend.CharRange{
				dsl.CharRange('a', 'c'),
				dsl.CharRange('u', 'v'),
				dsl.CharRange('x', 'z'),
			}
			Expect(frontend.RemoveCharRanges(characterRanges, dsl.CharRange('b', 'e'))).To(Equal([]frontend.CharRange{
				dsl.CharRange('a', 'c'),
				dsl.CharRange('u', 'v'),
				dsl.CharRange('x', 'z'),
			}))
		})

		It("should not remove when no match", func() {
			characterRanges := []frontend.CharRange{
				dsl.CharRange('a', 'c'),
				dsl.CharRange('u', 'v'),
				dsl.CharRange('x', 'z'),
			}
			Expect(frontend.RemoveCharRanges(characterRanges, dsl.CharRange('e', 'g'))).To(Equal([]frontend.CharRange{
				dsl.CharRange('a', 'c'),
				dsl.CharRange('u', 'v'),
				dsl.CharRange('x', 'z'),
			}))
		})
	})

	Context("NegateCharRanges", func() {
		It("should correctly negate a single character range", func() {
			negatedRange := frontend.NegateCharRanges([]frontend.CharRange{
				dsl.CharRange('b', 'b'),
			})
			Expect(negatedRange).To(Equal([]frontend.CharRange{
				dsl.CharRange(0, 'a'),
				dsl.CharRange('c', unicode.MaxRune),
			}))
		})

		It("should correctly negate a multi character range", func() {
			negatedRange := frontend.NegateCharRanges([]frontend.CharRange{
				dsl.CharRange('c', 'h'),
			})
			Expect(negatedRange).To(Equal([]frontend.CharRange{
				dsl.CharRange(0, 'b'),
				dsl.CharRange('i', unicode.MaxRune),
			}))
		})

		It("should correctly negate multiple character ranges", func() {
			negatedRange := frontend.NegateCharRanges([]frontend.CharRange{
				dsl.CharRange('b', 'f'),
				dsl.CharRange('j', 'j'),
				dsl.CharRange('m', 'o'),
			})
			Expect(negatedRange).To(Equal([]frontend.CharRange{
				dsl.CharRange(0, 'a'),
				dsl.CharRange('g', 'i'),
				dsl.CharRange('k', 'l'),
				dsl.CharRange('p', unicode.MaxRune),
			}))
		})
	})
})
