package model

import "cmp"

// CompletionContext is the completion kind from an offset on, up to the offset of the next context.
type CompletionContext struct {
	Offset int
	Kind   CompletionKind
}

// compareOffset compares the Offset of the context with the Offset.
func (c CompletionContext) compareOffset(offset int) int {
	return cmp.Compare(c.Offset, offset)
}
