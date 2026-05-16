package dsl

import (
	"unicode"

	"github.com/backbone81/golr/internal/scannergen/frontend"
)

// Rule constructs a rule for a regular expression with a name.
func Rule(name string, node *frontend.Node) frontend.Rule {
	return frontend.Rule{
		Name:  name,
		Regex: *node,
	}
}

// Any constructs a regular expression matching any character.
func Any() *frontend.Node {
	return &frontend.Node{
		Kind: frontend.KindAny,
	}
}

// CharClass constructs a regular expression matching a class of characters specified by character ranges.
func CharClass(charRanges ...frontend.CharRange) *frontend.Node {
	return &frontend.Node{
		Kind: frontend.KindCharClass,
		CharClass: frontend.CharClass{
			Ranges: charRanges,
		},
	}
}

// NegCharClass constructs a regular expression matching a class of characters outside the specified character ranges.
// The characters must still be in the range of valid Unicode characters from 0x00 to [unicode.MaxRune].
func NegCharClass(charRanges ...frontend.CharRange) *frontend.Node {
	return &frontend.Node{
		Kind: frontend.KindCharClass,
		CharClass: frontend.CharClass{
			Negate: true,
			Ranges: charRanges,
		},
	}
}

// CharRange is constructing character ranges for character classes.
// The character range must be in the range of valid Unicode characters from 0x00 to [unicode.MaxRune].
// Low must always be smaller or equal to high.
// Set low and high to the same character to have a single character.
func CharRange(low rune, high rune) frontend.CharRange {
	return frontend.CharRange{
		Low:  low,
		High: high,
	}
}

// UnicodeCategory constructs a list of character ranges matching the given Unicode character categories from the
// [unicode] package.
func UnicodeCategory(tables ...*unicode.RangeTable) []frontend.CharRange {
	var result []frontend.CharRange
	for _, table := range tables {
		for _, r16 := range table.R16 {
			result = append(result, CharRange(
				rune(r16.Lo),
				rune(r16.Hi),
			))
		}
		for _, r32 := range table.R32 {
			result = append(result, CharRange(
				rune(r32.Lo), //nolint:gosec // Integer overflow conversion is expected behavior
				rune(r32.Hi), //nolint:gosec // Integer overflow conversion is expected behavior
			))
		}
	}
	return result
}

// Concat constructs a regular expression matching all its children in sequence.
func Concat(children ...*frontend.Node) *frontend.Node {
	return &frontend.Node{
		Kind: frontend.KindConcat,
		Concat: frontend.Concat{
			Children: children,
		},
	}
}

// Literal constructs a regular expression matching its text as a literal.
func Literal(text string) *frontend.Node {
	return &frontend.Node{
		Kind: frontend.KindLiteral,
		Literal: frontend.Literal{
			Text: text,
		},
	}
}

// OneOrMore constructs a regular expression matching one or more instances of its child.
func OneOrMore(child *frontend.Node) *frontend.Node {
	return &frontend.Node{
		Kind: frontend.KindOneOrMore,
		OneOrMore: frontend.OneOrMore{
			Child: child,
		},
	}
}

// Optional constructs a regular expression matching zero or one instances of its child.
func Optional(child *frontend.Node) *frontend.Node {
	return &frontend.Node{
		Kind: frontend.KindOptional,
		Optional: frontend.Optional{
			Child: child,
		},
	}
}

// Or constructs a regular expression matching one of its children.
func Or(children ...*frontend.Node) *frontend.Node {
	return &frontend.Node{
		Kind: frontend.KindOr,
		Or: frontend.Or{
			Children: children,
		},
	}
}

// Repetition constructs a regular expression matching its child for a specific number of times.
// Minimum must always be smaller or equal to Maximum.
// Set Minimum and Maximum to the same value to have an exact number of repetitions.
func Repetition(child *frontend.Node, minimum int, maximum int) *frontend.Node {
	return &frontend.Node{
		Kind: frontend.KindRepetition,
		Repetition: frontend.Repetition{
			Minimum: minimum,
			Maximum: maximum,
			Child:   child,
		},
	}
}

// ZeroOrMore constructs a regular expression matching zero or more instances of its child.
func ZeroOrMore(child *frontend.Node) *frontend.Node {
	return &frontend.Node{
		Kind: frontend.KindZeroOrMore,
		ZeroOrMore: frontend.ZeroOrMore{
			Child: child,
		},
	}
}
