package backend

import (
	"errors"
	"fmt"

	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// ErrStateLimitExceeded is returned when a parser table grows beyond the number of states a transition action can
// address.
//
// This is a limit of the table encoding, not of any one core: TransitionAction packs the target state into the lower
// half of its bits, so no core can hand out a state index a transition cannot point at. Every core which grows a state
// slice therefore checks against MaxAddressableStates and reports this error instead of building a table whose
// transitions cannot be represented.
var ErrStateLimitExceeded = errors.New("the number of states exceeds the state limit")

// MaxAddressableStates returns the number of states a construction can build without ever handing a state index to
// NewTransitionAction which a transition action cannot address.
//
// The state limit is only checked once per state a construction takes off its work list, which means the last state can
// push the construction over the limit by as many states as it has transitions. Every symbol of the grammar can
// contribute at most one transition, so leaving room for one state per symbol keeps that overshoot addressable.
func MaxAddressableStates(grammar frontend.Grammar) int {
	symbolCount := len(grammar.Terminals) + len(grammar.Nonterminals)
	return TransitionActionMaxState + 1 - symbolCount
}

// CheckStateLimit reports ErrStateLimitExceeded when a table of stateCount states has grown beyond maxStates, and nil
// otherwise. Pass the maximum from MaxAddressableStates.
//
// The core name goes into the message because it is what tells the user what to do about it: a grammar whose canonical
// LR(1) tables do not fit can still have LALR(1) or IELR(1) tables which do.
func CheckStateLimit(coreName string, stateCount int, maxStates int) error {
	if stateCount <= maxStates {
		return nil
	}
	return fmt.Errorf("%s: %w of %d states", coreName, ErrStateLimitExceeded, maxStates)
}
