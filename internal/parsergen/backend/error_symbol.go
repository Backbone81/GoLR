package backend

import (
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// ErrorShiftStateIdx returns the state the given state transitions to when the error symbol is shifted, and reports if
// the state has a transition on the error symbol at all. A state which has one is a resynchronization point of an error
// recovery production: the place a parser resumes at after a syntax error, once it popped its stack down to that state.
// The error reference is the one frontend.ErrorTerminalRef resolves for the grammar of the parser.
//
// A transition action packs the symbol reference into the bits above the target state, so the transition action set is
// ordered by symbol first. Searching for the smallest transition action carrying the error symbol therefore lands on
// the transition on that symbol if the state has one, without knowing the target state up front.
func ErrorShiftStateIdx(state *State, errorRef frontend.SymbolRef) (int, bool) {
	idx := state.TransitionActions.LowerBound(NewTransitionAction(errorRef, 0))
	if idx == state.TransitionActions.Length() {
		return 0, false
	}
	transitionAction := state.TransitionActions.GetByIndex(idx)
	if transitionAction.SymbolRef() != errorRef {
		return 0, false
	}
	return transitionAction.StateIdx(), true
}
