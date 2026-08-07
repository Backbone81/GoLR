package backendtest

import (
	"fmt"
	"strings"
)

// Trace is the sequence of events a scanner or a parser produces for one input. Every event writes one line of the
// trace.
//
// The canonical text is UTF-8 with LF line endings and a trailing newline. All offsets in it are byte offsets and
// never rune or character offsets, because bytes are the only unit every target language agrees on without conversion.
// Demanding them is what forces each driver to deal with the string model of its own language correctly.
//
// A trace is only ever written here and never read back. Comparing two backends is a diff of the text they printed, so
// nothing needs the text in a structured form again.
//
// The events of a scanner and the events of a parser go into separate traces and are never mixed into one. Both kinds
// have an ERROR event which reports nothing but an offset, so a combined trace could not tell a scanner error from a
// parser error. Keeping them apart also lets a grammar without a parser have a trace of its own.
//
// Neither state indices nor production indices appear in any event. A trace is meant to be a property of the grammar
// and the input alone, so that it survives a change in state numbering, and so that it does not break when the core is
// switched between implementations which number productions differently. The nonterminal name together with the length
// of the right hand side identifies the production well enough for the grammars in the corpus, and the corpus is
// expected to keep it that way.
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
