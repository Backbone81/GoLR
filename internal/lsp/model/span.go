package model

import "cmp"

// Span is a range of bytes in the text.
type Span struct {
	Offset int
	Length int
}

// Compare orders spans by their offset.
func (s Span) Compare(other Span) int {
	return cmp.Compare(s.Offset, other.Offset)
}

// End returns the offset one past the last byte of the span.
func (s Span) End() int {
	return s.Offset + s.Length
}
