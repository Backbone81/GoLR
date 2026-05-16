package frontend

import (
	"github.com/backbone81/golr/internal/scannergen/frontend"
)

// Any is a regular expression matching any character.
// The character must be in the range of valid Unicode characters from 0x00 to [unicode.MaxRune].
type Any = frontend.Any

// CharClass is a regular expression matching a class of characters specified by character ranges.
type CharClass = frontend.CharClass

// CharRange is specifying character ranges for character classes.
// The character range must be in the range of valid Unicode characters from 0x00 to [unicode.MaxRune].
type CharRange = frontend.CharRange

// Concat is a regular expression matching all its children in sequence.
// The children need to implement the [Node] interface.
type Concat = frontend.Concat

// Literal is a regular expression matching its text as a literal.
type Literal = frontend.Literal

// Node is the interface all regular expressions need to implement.
type Node = frontend.Node

// OneOrMore is a regular expression matching one or more instances of its child.
// The child needs to implement the [Node] interface.
type OneOrMore = frontend.OneOrMore

// Optional is a regular expression matching zero or one instances of its child.
// The child needs to implement the [Node] interface.
type Optional = frontend.Optional

// Or is a regular expression matching one of its children.
// The children need to implement the [Node] interface.
type Or = frontend.Or

// Repetition is a regular expression matching its child for a specific number of times.
// The child needs to implement the [Node] interface.
type Repetition = frontend.Repetition

// ZeroOrMore is a regular expression matching zero or more instances of its child.
// The child needs to implement the [Node] interface.
type ZeroOrMore = frontend.ZeroOrMore
