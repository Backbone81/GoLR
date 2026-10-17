package model

// Span is a range of bytes in the text.
type Span struct {
	Offset int
	Length int
}

// End returns the offset one past the last byte of the span.
func (s Span) End() int {
	return s.Offset + s.Length
}
