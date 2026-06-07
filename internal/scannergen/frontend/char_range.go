package frontend

import (
	"errors"
	"fmt"
	"slices"
	"unicode"
)

// CharRange is specifying character ranges for character classes.
// The character range must be in the range of valid Unicode characters from 0x00 to [unicode.MaxRune].
// Low must always be smaller or equal to High.
// Set Low and High to the same character to have a single character.
type CharRange struct {
	Low  rune `json:"low"  yaml:"low"`
	High rune `json:"high" yaml:"high"`
}

// CharRange implements the [fmt.Stringer] interface.
var _ fmt.Stringer = (*CharRange)(nil)

// String returns a string representation of this regular expression.
func (c *CharRange) String() string {
	if c.Low != c.High {
		return fmt.Sprintf("%s-%s", c.printRune(c.Low), c.printRune(c.High))
	}
	return c.printRune(c.Low)
}

// printRune is a helper method for creating a readable representation of a character range.
func (c *CharRange) printRune(r rune) string {
	switch r {
	case '\t':
		return `\t`
	case '\n':
		return `\n`
	case '\r':
		return `\r`
	case '\v':
		return `\v`
	case '\f':
		return `\f`
	case 0:
		return `\0`
	case ']', '\\', '-', '^':
		return `\` + string(r)
	default:
		return string(r)
	}
}

// Validate reports if the character range satisfies the required conditions to be considered valid.
// A nil return value indicates that the character range is valid.
// An error return value provides details about the unmet condition.
func (c *CharRange) Validate() error {
	if c.Low < 0 || unicode.MaxRune < c.Low {
		return errors.New("low must be a valid Unicode character in the range from 0x00 to unicode.MaxRune")
	}
	if c.High < 0 || unicode.MaxRune < c.High {
		return errors.New("high must be a valid Unicode character in the range from 0x00 to unicode.MaxRune")
	}
	if c.High < c.Low {
		return errors.New("low must always be smaller or equal to high")
	}
	return nil
}

// SortCharRanges sorts the character ranges ascending by the [CharRange.Low] value of the character range.
func SortCharRanges(charRanges []CharRange) {
	slices.SortStableFunc(charRanges, func(a, b CharRange) int {
		switch {
		case a.Low < b.Low:
			return -1
		case a.Low > b.Low:
			return 1
		default:
			return 0
		}
	})
}

// SplitCharRanges splits up character ranges into two ranges when the given split point falls into them.
func SplitCharRanges(charRanges []CharRange, splitPoint rune) []CharRange {
	var result []CharRange
	for _, characterRange := range charRanges {
		if characterRange.Low < characterRange.High &&
			characterRange.Low < splitPoint &&
			splitPoint <= characterRange.High {
			result = append(result, CharRange{
				Low:  characterRange.Low,
				High: splitPoint - 1,
			})
			result = append(result, CharRange{
				Low:  splitPoint,
				High: characterRange.High,
			})
		} else {
			result = append(result, characterRange)
		}
	}
	return result
}

// RemoveCharRanges removes character ranges which reside fully inside the removeRange.
func RemoveCharRanges(charRanges []CharRange, removeRange CharRange) []CharRange {
	return slices.DeleteFunc(charRanges, func(r CharRange) bool {
		return removeRange.Low <= r.Low && r.High <= removeRange.High
	})
}

// NegateCharRanges creates the inverse from the given character ranges.
func NegateCharRanges(charRanges []CharRange) []CharRange {
	negatedRanges := []CharRange{
		{
			Low:  0,
			High: unicode.MaxRune,
		},
	}
	for _, characterRange := range charRanges {
		negatedRanges = SplitCharRanges(negatedRanges, characterRange.Low)
		negatedRanges = SplitCharRanges(negatedRanges, characterRange.High+1)
		negatedRanges = RemoveCharRanges(negatedRanges, characterRange)
	}
	return negatedRanges
}
