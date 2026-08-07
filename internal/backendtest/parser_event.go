package backendtest

import (
	"fmt"
)

// Shift reports a terminal being shifted onto the parse stack.
type Shift struct {
	TerminalName string
	Start        int
	End          int
}

// String returns the canonical trace line for the event, without the terminating newline.
func (s Shift) String() string {
	return fmt.Sprintf("SHIFT %s %d %d", s.TerminalName, s.Start, s.End)
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

// Pop reports how many states error recovery took off the parse stack while looking for a state which shifts the error
// token.
type Pop struct {
	Count int
}

// String returns the canonical trace line for the event, without the terminating newline.
func (p Pop) String() string {
	return fmt.Sprintf("POP %d", p.Count)
}

// Discard reports a terminal which error recovery threw away while looking for one the parser can continue on.
type Discard struct {
	TerminalName string
	Start        int
	End          int
}

// String returns the canonical trace line for the event, without the terminating newline.
func (d Discard) String() string {
	return fmt.Sprintf("DISCARD %s %d %d", d.TerminalName, d.Start, d.End)
}

// Resync reports that error recovery found a terminal it can continue on and that the parser resumed.
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
