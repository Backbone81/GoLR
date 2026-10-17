package model

import "cmp"

// Occurrence is a definition, alias or reference.
type Occurrence struct {
	Name string
	Role Role
	Span Span
}

// compareEnd compares the end of the occurrence with the offset.
func (o Occurrence) compareEnd(offset int) int {
	return cmp.Compare(o.Span.End(), offset)
}
