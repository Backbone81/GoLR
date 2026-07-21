package conflict

import (
	"context"
	"errors"
	"runtime/trace"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/utils"
)

// RemoveUnreachableStates removes every state which the parser can no longer reach from the start state and renumbers
// the states which are left, so that the transitions of the parser tables keep pointing at the states they pointed at
// before. The conflicts come back with their state indexes brought up to date, so that a conflict still names the state
// it occurred in. It takes what Resolve returns and returns the same again, which makes it a filter to put behind
// Resolve in the call chain of a parser core.
//
// This is the unreachable state removal of section 3.8.2 of IELR(1), which the paper describes as its optional phase 6,
// and Definition 3.48 is what it removes: a state is unreachable when no sequence of grammar symbols takes the parser
// from the start state to it. Resolving the conflicts is what strands the states in the first place. A conflict which a
// shift loses removes the transition on the conflicted terminal, and when that transition was the only way into its
// target state, the target and everything which is only reachable through it become dead weight in the parser tables.
// Removing them costs nothing in correctness, because a parser can never take a transition into them, and it keeps the
// tables from growing with states no input can ever reach.
//
// The parser which is handed in is spent: its states are compacted in place instead of into a copy, because this runs
// after Resolve, on parser tables which are on their way to a backend and are of no use to anyone in their old shape
// afterwards, which is the same contract Resolve itself follows. Only the parser which is returned describes the parser
// tables correctly.
//
// The states keep their relative order while they are compacted, so a parser table which has nothing to remove keeps
// every state index it had and no conflict moves. Removing a state can strand further states, but this needs no
// repeated passes to settle: a state survives exactly when the traversal from the start state arrives at it, and a
// traversal which never enters a stranded state never leaves it either, so the cascade is already accounted for the
// first time around. The whole pass is linear in the size of the parser tables, visiting every state and every
// transition of a state once.
func RemoveUnreachableStates(parser backend.Parser, conflicts []Conflict) (backend.Parser, []Conflict) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Conflict: RemoveUnreachableStates").End()

	if len(parser.States) == 0 {
		// A parser table without any state has no start state to reach anything from.
		return parser, conflicts
	}

	reachable := reachableStates(parser.States)

	// The states keep the order they are in, so only the states which are removed shift the states behind them forward.
	// A state which is removed has no new state index, which -1 stands for.
	newStateIdxByOldStateIdx := make([]int, len(parser.States))
	newStateIdx := 0
	for oldStateIdx := range parser.States {
		if !reachable.Contains(oldStateIdx) {
			newStateIdxByOldStateIdx[oldStateIdx] = -1
			continue
		}
		newStateIdxByOldStateIdx[oldStateIdx] = newStateIdx
		newStateIdx++
	}
	if newStateIdx == len(parser.States) {
		// Every state is reachable, so there is nothing to compact, no transition which points somewhere else than it
		// did before, and no conflict which names a different state than it did before.
		return parser, conflicts
	}

	for oldStateIdx := range parser.States {
		if !reachable.Contains(oldStateIdx) {
			continue
		}
		state := parser.States[oldStateIdx]
		state.TransitionActions = renumberTransitionActions(state.TransitionActions, newStateIdxByOldStateIdx)
		// The state moves to its new index, which is never behind its old one, so this never overwrites a state which
		// has not been moved yet.
		parser.States[newStateIdxByOldStateIdx[oldStateIdx]] = state
	}
	parser.States = parser.States[:newStateIdx]
	return parser, remapConflicts(conflicts, newStateIdxByOldStateIdx)
}

// remapConflicts brings the state indexes of the conflicts up to date after the states were renumbered, so that a
// conflict still names the state it occurred in. It takes the new state index of every state, indexed by the old state
// index, with -1 for the states which were removed.
//
// A conflict of a state which was removed is dropped. Such a conflict is one the policy decided, because an unresolved
// conflict makes the caller give up on the grammar before any state is removed, and it happened in a state no input can
// reach anymore, so there is nothing left for the grammar author to act on. There is also no state index left to report
// it under, and reporting it under the index of whichever state moved into its place would name the wrong state.
func remapConflicts(conflicts []Conflict, newStateIdxByOldStateIdx []int) []Conflict {
	var result []Conflict
	for _, c := range conflicts {
		newStateIdx := newStateIdxByOldStateIdx[c.StateIdx]
		if newStateIdx < 0 {
			continue
		}
		c.StateIdx = newStateIdx
		result = append(result, c)
	}
	return result
}

// reachableStates returns the state indexes of the states which the parser can reach from the start state, which is
// state 0. Reachability follows the transitions of the states on terminals and nonterminals alike, because both of them
// are symbols the parser can move over, which is what Definition 3.48 of IELR(1) asks for.
func reachableStates(states []backend.State) utils.Bitset {
	var result utils.Bitset
	result.Add(0)

	var pending utils.Stack[int]
	pending.Push(0)
	for !pending.IsEmpty() {
		stateIdx := pending.Top()
		pending.Pop()

		for _, transitionAction := range states[stateIdx].TransitionActions.All() {
			if !result.Add(transitionAction.StateIdx()) {
				// The state was reached before, so its transitions have been followed already.
				continue
			}
			pending.Push(transitionAction.StateIdx())
		}
	}
	return result
}

// renumberTransitionActions returns the transition actions with every target state index replaced by the new state
// index of that state.
//
// A state which survives never transitions into a state which does not, because the target of a transition of a
// reachable state is reachable itself, so every target has a new state index to be renumbered to.
func renumberTransitionActions(
	transitionActions backend.TransitionActionSet,
	newStateIdxByOldStateIdx []int,
) backend.TransitionActionSet {
	renumbered := make([]backend.TransitionAction, 0, transitionActions.Length())
	for _, transitionAction := range transitionActions.All() {
		newStateIdx := newStateIdxByOldStateIdx[transitionAction.StateIdx()]
		utils.DebugAssert(func() error {
			if newStateIdx < 0 {
				return errors.New("reachable state has a transition into an unreachable state")
			}
			return nil
		})
		renumbered = append(renumbered, backend.NewTransitionAction(transitionAction.SymbolRef(), newStateIdx))
	}
	// A state has at most one transition per symbol, and a transition action orders by its symbol before its target
	// state, so renumbering the target states leaves the transition actions in the order they already were in.
	return backend.NewTransitionActionSet(renumbered...)
}
