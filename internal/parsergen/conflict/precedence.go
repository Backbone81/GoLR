package conflict

import (
	"slices"

	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// IsNoPrecedence reports if the precedence level says that no precedence was declared at all, which is what a level of
// zero means. A terminal without a precedence declaration keeps the zero value of the field, and no frontend hands out
// level zero to a terminal it declares a precedence for.
//
// Every other level is a declared precedence, where a higher level binds tighter than a lower level. Only the order of
// two levels is meaningful, never the distance between them, because the frontends are free to number the levels
// however they like as long as a tighter binding terminal ends up with a higher level. The GoLR frontend counts the
// levels down from math.MaxInt, because the precedence declared first binds tightest there, while the Bison frontend
// counts them up from one, because Bison lets the precedence declared last bind tightest.
func IsNoPrecedence(precedence int) bool {
	return precedence == 0
}

// ProductionPrecedence returns the precedence level of the production, which the production inherits from a terminal.
// That terminal is the one the production declares explicitly, and the rightmost terminal on the right hand side of the
// production otherwise. The level is the one of IsNoPrecedence when the production has no terminal to inherit a
// precedence from, or when that terminal has no precedence declared.
//
// A production has no associativity to go with the level. A conflict between two contributions which bind equally tight
// is decided by the associativity of the conflicted terminal alone.
func ProductionPrecedence(grammar frontend.Grammar, productionIdx int) int {
	production := grammar.Productions[productionIdx]
	if production.PrecedenceTerminalIdx != nil {
		return grammar.Terminals[*production.PrecedenceTerminalIdx].Precedence
	}

	for _, symbolRef := range slices.Backward(production.SymbolRefs) {
		if symbolRef.IsTerminal() {
			return grammar.Terminals[symbolRef.Idx()].Precedence
		}
	}
	return 0
}
