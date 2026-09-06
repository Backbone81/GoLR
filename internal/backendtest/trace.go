package backendtest

import (
	"fmt"
	"strings"
)

// Trace is the sequence of events a scanner or a parser produces for one input, one line per event. The canonical text
// is UTF-8 with LF line endings and a trailing newline.
//
// A scanner trace locates everything by byte offset. A parser trace locates every action by line and column, so an
// unexpected parse in a large source can be found by position.
//
// Scanner and parser events go into separate traces, never one. No event carries a state or production index, so a
// trace is a property of the grammar and the input: a reduction is named "lhs => rhs", not by number.
type Trace []fmt.Stringer

// String returns the canonical text of the trace. An empty trace is the empty string, so that it does not consist of a
// single empty line.
func (t Trace) String() string {
	if len(t) == 0 {
		return ""
	}

	lines := make([]string, 0, len(t))
	for _, event := range t {
		lines = append(lines, event.String())
	}
	return strings.Join(lines, "\n") + "\n"
}
