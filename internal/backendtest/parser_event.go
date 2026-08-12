package backendtest

import (
	"fmt"
)

// Shift reports a terminal being shifted onto the parse stack, which is one leaf of the parse tree.
//
// The event carries no byte offsets, deliberately. The scanner trace of the same case already states the extent of
// every token, so offsets here would re-test the scanner rather than the parser. Leaving them out is also what keeps
// the trace derivable from a finished parse tree, since a byte offset per tree node is something no generated parser
// carries today and nothing but this trace would want.
type Shift struct {
	TerminalName string
}

// String returns the canonical trace line for the event, without the terminating newline.
func (s Shift) String() string {
	return "SHIFT " + s.TerminalName
}

// Reduce reports a production being reduced, identified by its left hand side and by the number of symbols taken off
// the parse stack.
type Reduce struct {
	NonterminalName     string
	RightHandSideLength int
}

// String returns the canonical trace line for the event, without the terminating newline.
func (r Reduce) String() string {
	return fmt.Sprintf("REDUCE %s %d", r.NonterminalName, r.RightHandSideLength)
}

// ParserError reports the byte offset of the terminal on which the parser detected an error. The offset is the finding,
// not an incidental detail, because a default reduction delays the detection of an error and every language has to
// delay it by exactly as much.
type ParserError struct {
	Offset int
}

// String returns the canonical trace line for the event, without the terminating newline.
func (p ParserError) String() string {
	return fmt.Sprintf("ERROR %d", p.Offset)
}

// Resync reports that error recovery found a state which shifts the error symbol and that the parser resumed there. It
// is the leaf that shift left in the parse tree, so it appears where that leaf appears in the walk and stands for the
// part of the input which was dropped.
//
// There is no event for the states the recovery popped, nor for the tokens it discarded. Neither survives in the parse
// tree, which is all a generated parser hands back, so neither can be observed without a trace hook on every generated
// parser. What the recovery threw away is visible instead as the shifts and reductions which are missing from the walk.
type Resync struct{}

// String returns the canonical trace line for the event, without the terminating newline.
func (r Resync) String() string {
	return "RESYNC"
}

// Accept reports that the parser accepted the input.
type Accept struct{}

// String returns the canonical trace line for the event, without the terminating newline.
func (a Accept) String() string {
	return "ACCEPT"
}
