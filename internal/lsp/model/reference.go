package model

import "cmp"

// Reference is a use of a symbol by name or by string: in a production body, a precedence line, @precedence(...) or
// @start.
type Reference struct {
	Name string
	Span Span
}

// compareOffset orders references by their offset.
func (r Reference) compareOffset(other Reference) int {
	return cmp.Compare(r.Span.Offset, other.Span.Offset)
}
