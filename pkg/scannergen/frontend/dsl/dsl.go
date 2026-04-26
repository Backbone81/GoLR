package dsl

import (
	intdsl "golr/internal/scannergen/frontend/dsl"
)

// Any constructs a regular expression matching any character.
// The character must be in the range of valid Unicode characters from 0x00 to [unicode.MaxRune].
var Any = intdsl.Any

// CharClass constructs a regular expression matching a class of characters specified by character ranges.
var CharClass = intdsl.CharClass

// NegCharClass constructs a regular expression matching a class of characters outside the specified character ranges.
// The characters must still be in the range of valid Unicode characters from 0x00 to [unicode.MaxRune].
var NegCharClass = intdsl.NegCharClass

// CharRange is constructing character ranges for character classes.
// The character range must be in the range of valid Unicode characters from 0x00 to [unicode.MaxRune].
// Low must always be smaller or equal to high.
// Set low and high to the same character to have a single character.
var CharRange = intdsl.CharRange

// UnicodeCategory constructs a list of character ranges matching the given Unicode character categories from the
// [unicode] package.
var UnicodeCategory = intdsl.UnicodeCategory

// Concat constructs a regular expression matching all its children in sequence.
// The children need to implement the [Node] interface.
var Concat = intdsl.Concat

// Literal constructs a regular expression matching its text as a literal.
var Literal = intdsl.Literal

// OneOrMore constructs a regular expression matching one or more instances of its child.
// The child needs to implement the [Node] interface.
var OneOrMore = intdsl.OneOrMore

// Optional constructs a regular expression matching zero or one instances of its child.
// The child needs to implement the [Node] interface.
var Optional = intdsl.Optional

// Or constructs a regular expression matching one of its children.
// The children need to implement the [Node] interface.
var Or = intdsl.Or

// Repetition constructs a regular expression matching its child for a specific number of times.
// The child needs to implement the [Node] interface.
// Minimum must always be smaller or equal to Maximum.
// Set Minimum and Maximum to the same value to have an exact number of repetitions.
var Repetition = intdsl.Repetition

// ZeroOrMore constructs a regular expression matching zero or more instances of its child.
// The child needs to implement the [Node] interface.
var ZeroOrMore = intdsl.ZeroOrMore
