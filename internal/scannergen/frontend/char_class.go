package frontend

import (
	"strings"
)

// CharClass is a regular expression matching a class of characters specified by character ranges.
type CharClass struct {
	// Negate reports if the character class should be negated/inverted. A negated character class matches all
	// characters not specified in the character ranges. The characters must still be in the range of valid Unicode
	// characters from 0x00 to [unicode.MaxRune].
	Negate bool        `json:"negate" yaml:"negate"`
	Ranges []CharRange `json:"ranges" yaml:"ranges"`
}

// String returns a string representation of this regular expression.
func (c *CharClass) String() string {
	var result strings.Builder
	result.WriteString("[")
	if c.Negate {
		result.WriteString("^")
	}
	for i := range c.Ranges {
		result.WriteString(c.Ranges[i].String())
	}
	result.WriteString("]")
	return result.String()
}

// IsSingleNode reports if this regular expression can be represented as a single node when converted to a string.
// Regular expressions which cannot be represented as a single node need to have parenthesis placed around them to
// form a subexpression.
func (c *CharClass) IsSingleNode() bool {
	return true
}

// Validate reports if the regular expression satisfies the required conditions to be considered valid.
// A nil return value indicates that the regular expression is valid.
// An error return value provides details about the unmet condition.
// In situations where the regular expression has children, all children are checked for validity recursively. If
// any child is not valid, this regular expression is also not valid.
func (c *CharClass) Validate() error {
	// We explicitly allow empty char classes for situations where a token needs to be declared but the scanner should
	// not match anything. This is helpful for situations where not all tokens can reliably be derived by a DFA but
	// instead have some overlay mechanic over the base scanner.
	for _, charRange := range c.Ranges {
		if err := charRange.Validate(); err != nil {
			return err
		}
	}
	return nil
}
