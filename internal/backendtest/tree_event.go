package backendtest

import (
	"fmt"
	"strings"
)

// A tree trace is the parse tree a parse produced, one line per node, in pre-order. It is the third trace of a case,
// next to the scanner trace and the parser trace, and the only one which states what a parse built rather than what it
// did. A parse which was given up produces no tree and therefore an empty trace.
//
// Every line carries the position of the node, the span of the input it covers as "offset+length", and the node
// itself, indented by two spaces per level of depth. The position and the span columns are therefore in the same place
// on every line, however deep a node sits, and a tree trace lines up with the scanner and the parser trace of the same
// input.
//
// The span is what makes this trace worth committing: it pins down which part of the input every node covers, which is
// what a user of the parse tree reports an error with. The rules the spans follow are part of the format:
//
//   - A terminal covers its lexeme.
//   - A nonterminal covers its children, from the start of the first child which covers something up to the end of the
//     last one which does. Children which cover nothing do not widen the span, so no whitespace around an ε-child is
//     pulled in.
//   - A node which covers nothing - an ε-production, or an error node which dropped nothing - has zero length at the
//     position the parser was looking at when it was created.
//   - An error node covers everything the recovery round which pushed it dropped, which is the tokens it discarded and
//     the nodes it popped.
//
// The invariant over the whole format is that the text of a node is the source its span covers, and that for a
// terminal this is the lexeme the scanner produced.

// treeLine formats one line of a tree trace: the position of the node, its span, and the node itself indented by its
// depth. Indenting the payload rather than the whole line is what keeps the position and the span columns aligned.
func treeLine(line int, column int, byteOffset int, byteLength int, depth int, payload string) string {
	span := fmt.Sprintf("%d+%d", byteOffset, byteLength)
	return traceLine(line, column, span, strings.Repeat("  ", depth)+payload)
}

// TreeTerminal is a leaf the parser shifted a token onto. Its text is the source its span covers, which for a terminal
// is the lexeme the scanner produced.
type TreeTerminal struct {
	Line         int
	Column       int
	ByteOffset   int
	ByteLength   int
	Depth        int
	TerminalName string
	Text         string
}

// String returns the canonical trace line for the node, without the terminating newline.
func (t TreeTerminal) String() string {
	return treeLine(t.Line, t.Column, t.ByteOffset, t.ByteLength, t.Depth,
		t.TerminalName+` "`+EscapeLexeme(t.Text)+`"`)
}

// TreeNonterminal is an internal node the parser reduced a production to. It names the production the way a reduce
// event of a parser trace does, so the two traces of a case can be read against each other.
type TreeNonterminal struct {
	Line          int
	Column        int
	ByteOffset    int
	ByteLength    int
	Depth         int
	LeftHandSide  string
	RightHandSide []string
}

// String returns the canonical trace line for the node, without the terminating newline.
func (t TreeNonterminal) String() string {
	return treeLine(t.Line, t.Column, t.ByteOffset, t.ByteLength, t.Depth,
		productionText(t.LeftHandSide, t.RightHandSide))
}

// TreeErrorNode is the leaf error recovery pushed for the part of the input it dropped. It stands for no token of its
// own, so its span is all it carries. It is named like the error symbol of a parser trace, with the leading dollar sign
// no rule of a scanner can have, so it can not be mistaken for a terminal.
type TreeErrorNode struct {
	Line       int
	Column     int
	ByteOffset int
	ByteLength int
	Depth      int
}

// String returns the canonical trace line for the node, without the terminating newline.
func (t TreeErrorNode) String() string {
	return treeLine(t.Line, t.Column, t.ByteOffset, t.ByteLength, t.Depth, "$error")
}
