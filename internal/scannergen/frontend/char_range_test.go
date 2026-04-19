package frontend_test

import (
	"golr/internal/scannergen/frontend"
	"unicode"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CharRange", func() {
	It("should convert to string with a single character", func() {
		expression := frontend.CharRange{
			Low:  'a',
			High: 'a',
		}
		Expect(expression.String()).To(Equal("a"))
	})

	It("should convert to string with a character range", func() {
		expression := frontend.CharRange{
			Low:  'a',
			High: 'z',
		}
		Expect(expression.String()).To(Equal("a-z"))
	})

	It("should convert to string with special characters", func() {
		expression := frontend.CharRange{
			Low:  ' ',
			High: '\t',
		}
		Expect(expression.String()).To(Equal("' '-\\t"))

		expression = frontend.CharRange{
			Low:  '\r',
			High: '\n',
		}
		Expect(expression.String()).To(Equal("\\r-\\n"))

		expression = frontend.CharRange{
			Low:  '\u2116',
			High: '\u26a0',
		}
		Expect(expression.String()).To(Equal("0x2116-0x26a0"))
	})

	It("should fail validation with out of range runes", func() {
		expression := frontend.CharRange{
			Low:  -1,
			High: 'c',
		}
		Expect(expression.Validate()).ToNot(Succeed())

		expression = frontend.CharRange{
			Low:  unicode.MaxRune + 1,
			High: 'c',
		}
		Expect(expression.Validate()).ToNot(Succeed())

		expression = frontend.CharRange{
			Low:  'a',
			High: -1,
		}
		Expect(expression.Validate()).ToNot(Succeed())

		expression = frontend.CharRange{
			Low:  'a',
			High: unicode.MaxRune + 1,
		}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation when high is lower than low", func() {
		expression := frontend.CharRange{
			Low:  'c',
			High: 'a',
		}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should validate successfully", func() {
		expression := frontend.CharRange{
			Low:  'a',
			High: 'a',
		}
		Expect(expression.Validate()).To(Succeed())

		expression = frontend.CharRange{
			Low:  'a',
			High: 'b',
		}
		Expect(expression.Validate()).To(Succeed())
	})

	Context("SortCharRanges", func() {
		It("should sort ascending", func() {
			characterRanges := []frontend.CharRange{
				{
					Low:  'x',
					High: 'y',
				},
				{
					Low:  'a',
					High: 'b',
				},
			}
			frontend.SortCharRanges(characterRanges)
			Expect(characterRanges).To(Equal([]frontend.CharRange{
				{
					Low:  'a',
					High: 'b',
				},
				{
					Low:  'x',
					High: 'y',
				},
			}))
		})

		It("should not change the order of already sorted character range", func() {
			characterRanges := []frontend.CharRange{
				{
					Low:  'a',
					High: 'b',
				},
				{
					Low:  'x',
					High: 'y',
				},
			}
			frontend.SortCharRanges(characterRanges)
			Expect(characterRanges).To(Equal([]frontend.CharRange{
				{
					Low:  'a',
					High: 'b',
				},
				{
					Low:  'x',
					High: 'y',
				},
			}))
		})

		It("should correctly deal with character ranges starting with the same rune", func() {
			characterRanges := []frontend.CharRange{
				{
					Low:  'a',
					High: 'b',
				},
				{
					Low:  'a',
					High: 'z',
				},
			}
			frontend.SortCharRanges(characterRanges)
			Expect(characterRanges).To(Equal([]frontend.CharRange{
				{
					Low:  'a',
					High: 'b',
				},
				{
					Low:  'a',
					High: 'z',
				},
			}))
		})
	})

	Context("SplitCharRanges", func() {
		It("should split in the middle", func() {
			characterRanges := []frontend.CharRange{
				{
					Low:  'a',
					High: 'z',
				},
			}
			splitCharacterRanges := frontend.SplitCharRanges(characterRanges, 'f')
			Expect(splitCharacterRanges).To(Equal([]frontend.CharRange{
				{
					Low:  'a',
					High: 'e',
				},
				{
					Low:  'f',
					High: 'z',
				},
			}))
		})

		It("should not split on the start", func() {
			characterRanges := []frontend.CharRange{
				{
					Low:  'a',
					High: 'z',
				},
			}
			splitCharacterRanges := frontend.SplitCharRanges(characterRanges, 'a')
			Expect(splitCharacterRanges).To(Equal([]frontend.CharRange{
				{
					Low:  'a',
					High: 'z',
				},
			}))
		})

		It("should split on the end", func() {
			characterRanges := []frontend.CharRange{
				{
					Low:  'a',
					High: 'z',
				},
			}
			splitCharacterRanges := frontend.SplitCharRanges(characterRanges, 'z')
			Expect(splitCharacterRanges).To(Equal([]frontend.CharRange{
				{
					Low:  'a',
					High: 'y',
				},
				{
					Low:  'z',
					High: 'z',
				},
			}))
		})

		It("should not split out of range at the end", func() {
			characterRanges := []frontend.CharRange{
				{
					Low:  'a',
					High: 'f',
				},
			}
			splitCharacterRanges := frontend.SplitCharRanges(characterRanges, 'u')
			Expect(splitCharacterRanges).To(Equal([]frontend.CharRange{
				{
					Low:  'a',
					High: 'f',
				},
			}))
		})

		It("should not split out of range at the start", func() {
			characterRanges := []frontend.CharRange{
				{
					Low:  'u',
					High: 'w',
				},
			}
			splitCharacterRanges := frontend.SplitCharRanges(characterRanges, 'a')
			Expect(splitCharacterRanges).To(Equal([]frontend.CharRange{
				{
					Low:  'u',
					High: 'w',
				},
			}))
		})

		It("should not split single character ranges", func() {
			characterRanges := []frontend.CharRange{
				{
					Low:  'b',
					High: 'b',
				},
			}
			splitCharacterRanges := frontend.SplitCharRanges(characterRanges, 'b')
			Expect(splitCharacterRanges).To(Equal([]frontend.CharRange{
				{
					Low:  'b',
					High: 'b',
				},
			}))
		})
	})

	Context("RemoveCharRanges", func() {
		It("should remove an exact match", func() {
			characterRanges := []frontend.CharRange{
				{
					Low:  'a',
					High: 'c',
				},
				{
					Low:  'u',
					High: 'v',
				},
				{
					Low:  'x',
					High: 'z',
				},
			}
			Expect(frontend.RemoveCharRanges(characterRanges, frontend.CharRange{
				Low:  'u',
				High: 'v',
			})).To(Equal([]frontend.CharRange{
				{
					Low:  'a',
					High: 'c',
				},
				{
					Low:  'x',
					High: 'z',
				},
			}))
		})

		It("should remove on overlap", func() {
			characterRanges := []frontend.CharRange{
				{
					Low:  'a',
					High: 'c',
				},
				{
					Low:  'u',
					High: 'v',
				},
				{
					Low:  'x',
					High: 'z',
				},
			}
			Expect(frontend.RemoveCharRanges(characterRanges, frontend.CharRange{
				Low:  't',
				High: 'w',
			})).To(Equal([]frontend.CharRange{
				{
					Low:  'a',
					High: 'c',
				},
				{
					Low:  'x',
					High: 'z',
				},
			}))
		})

		It("should remove multiples in range", func() {
			characterRanges := []frontend.CharRange{
				{
					Low:  'b',
					High: 'c',
				},
				{
					Low:  'u',
					High: 'v',
				},
				{
					Low:  'x',
					High: 'z',
				},
			}
			Expect(frontend.RemoveCharRanges(characterRanges, frontend.CharRange{
				Low:  'a',
				High: 'w',
			})).To(Equal([]frontend.CharRange{
				{
					Low:  'x',
					High: 'z',
				},
			}))
		})

		It("should not remove on partial match", func() {
			characterRanges := []frontend.CharRange{
				{
					Low:  'a',
					High: 'c',
				},
				{
					Low:  'u',
					High: 'v',
				},
				{
					Low:  'x',
					High: 'z',
				},
			}
			Expect(frontend.RemoveCharRanges(characterRanges, frontend.CharRange{
				Low:  'b',
				High: 'e',
			})).To(Equal([]frontend.CharRange{
				{
					Low:  'a',
					High: 'c',
				},
				{
					Low:  'u',
					High: 'v',
				},
				{
					Low:  'x',
					High: 'z',
				},
			}))
		})

		It("should not remove when no match", func() {
			characterRanges := []frontend.CharRange{
				{
					Low:  'a',
					High: 'c',
				},
				{
					Low:  'u',
					High: 'v',
				},
				{
					Low:  'x',
					High: 'z',
				},
			}
			Expect(frontend.RemoveCharRanges(characterRanges, frontend.CharRange{
				Low:  'e',
				High: 'g',
			})).To(Equal([]frontend.CharRange{
				{
					Low:  'a',
					High: 'c',
				},
				{
					Low:  'u',
					High: 'v',
				},
				{
					Low:  'x',
					High: 'z',
				},
			}))
		})
	})

	Context("NegateCharRanges", func() {
		It("should correctly negate a single character range", func() {
			negatedRange := frontend.NegateCharRanges([]frontend.CharRange{
				{
					Low:  'b',
					High: 'b',
				},
			})
			Expect(negatedRange).To(Equal([]frontend.CharRange{
				{
					Low:  0,
					High: 'a',
				},
				{
					Low:  'c',
					High: unicode.MaxRune,
				},
			}))
		})

		It("should correctly negate a multi character range", func() {
			negatedRange := frontend.NegateCharRanges([]frontend.CharRange{
				{
					Low:  'c',
					High: 'h',
				},
			})
			Expect(negatedRange).To(Equal([]frontend.CharRange{
				{
					Low:  0,
					High: 'b',
				},
				{
					Low:  'i',
					High: unicode.MaxRune,
				},
			}))
		})

		It("should correctly negate multiple character ranges", func() {
			negatedRange := frontend.NegateCharRanges([]frontend.CharRange{
				{
					Low:  'b',
					High: 'f',
				},
				{
					Low:  'j',
					High: 'j',
				},
				{
					Low:  'm',
					High: 'o',
				},
			})
			Expect(negatedRange).To(Equal([]frontend.CharRange{
				{
					Low:  0,
					High: 'a',
				},
				{
					Low:  'g',
					High: 'i',
				},
				{
					Low:  'k',
					High: 'l',
				},
				{
					Low:  'p',
					High: unicode.MaxRune,
				},
			}))
		})
	})
})
