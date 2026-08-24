package table

import (
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// NoTerminal is the terminal index used where a grammar does not have the terminal in question at all. Terminal indexes
// are never negative, so this can not be confused with one.
const NoTerminal = -1

// errorTerminalIdx returns the terminal index of the error symbol of the given grammar, or NoTerminal when the grammar
// does not have the error symbol.
//
// The error symbol sits at no fixed terminal index, so it is resolved once for a whole grammar and compared by index
// afterwards. Note that its presence in the grammar says nothing about whether the grammar uses it: the Bison backed
// cores seed the symbol into every grammar they hand back. What tells the two apart is whether any state shifts it, see
// CompressedParser.ErrorShiftStateIdx.
func errorTerminalIdx(grammar frontend.Grammar) int {
	errorRef, hasErrorTerminal := frontend.ErrorTerminalRef(grammar)
	if !hasErrorTerminal {
		return NoTerminal
	}
	return errorRef.Idx()
}
