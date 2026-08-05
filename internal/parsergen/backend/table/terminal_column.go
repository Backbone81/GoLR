package table

import (
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// NoTerminalColumn is the column of the action table which no terminal of the grammar occupies. It is always empty, so
// a lookup in it finds no entry and falls through to the default action of the state, which is what a parser does with
// a token it does not know.
//
// It exists because a generated parser has to translate the token its scanner delivers into the column which holds the
// decisions for it, and the two are numbered independently: the scanner numbers its tokens, the parser generator
// numbers the terminals of the grammar, and nothing makes the two agree. Only the GoLR grammar format declares both in
// one place, and even there the error symbol is added to the grammar after the scanner section has been read. So the
// translation is a lookup, and a lookup needs an answer for a token which is no terminal of this grammar at all - the
// scanner may have tokens the grammar never mentions.
const NoTerminalColumn = 0

// TerminalColumn returns the column of the action table which holds the decisions for the given terminal. This is the
// terminal index moved up by one, which is what leaves NoTerminalColumn free.
func TerminalColumn(terminalIdx int) int {
	return terminalIdx + 1
}

// terminalColumnCount returns the number of columns the action table of the given grammar has, which is one per
// terminal plus NoTerminalColumn.
func terminalColumnCount(grammar frontend.Grammar) int {
	return len(grammar.Terminals) + 1
}
