package backend

import (
	"errors"
	"fmt"
	"slices"
)

// ByteRange is specifying byte ranges for character classes.
// Low must always be smaller or equal to High.
// Set Low and High to the same byte to have a single byte
// Note that frontend.CharRange is working with unicode runes while ByteRange is working with bytes after unicode
// characters have been encoded as UTF-8 byte sequences.
type ByteRange struct {
	Low  byte `json:"low"  yaml:"low"`
	High byte `json:"high" yaml:"high"`
}

// ByteRange implements the [fmt.Stringer] interface.
var _ fmt.Stringer = (*ByteRange)(nil)

// String returns a string representation of this regular expression.
func (c *ByteRange) String() string {
	if c.Low != c.High {
		return fmt.Sprintf("%s-%s", c.printByte(c.Low), c.printByte(c.High))
	}
	return c.printByte(c.Low)
}

// printByte is a helper method for creating a readable representation of a byte range.
func (c *ByteRange) printByte(r byte) string {
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

// Validate reports if the byte range satisfies the required conditions to be considered valid.
// A nil return value indicates that the byte range is valid.
// An error return value provides details about the unmet condition.
func (c *ByteRange) Validate() error {
	if c.High < c.Low {
		return errors.New("low must always be smaller or equal to high")
	}
	return nil
}

// SplitByteRanges splits up byte ranges into two ranges when the given split point falls into them.
func SplitByteRanges(byteRanges []ByteRange, splitPoint byte) []ByteRange {
	var result []ByteRange
	for _, byteRange := range byteRanges {
		if byteRange.Low < byteRange.High &&
			byteRange.Low < splitPoint &&
			splitPoint <= byteRange.High {
			result = append(result, ByteRange{
				Low:  byteRange.Low,
				High: splitPoint - 1,
			})
			result = append(result, ByteRange{
				Low:  splitPoint,
				High: byteRange.High,
			})
		} else {
			result = append(result, byteRange)
		}
	}
	return result
}

// RemoveByteRanges removes byte ranges which reside fully inside the removeRange.
func RemoveByteRanges(byteRanges []ByteRange, removeRange ByteRange) []ByteRange {
	return slices.DeleteFunc(byteRanges, func(r ByteRange) bool {
		return removeRange.Low <= r.Low && r.High <= removeRange.High
	})
}
