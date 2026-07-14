package conflict

import (
	"context"
	"errors"
	"maps"
	"runtime/trace"
	"slices"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

// Conflict describes a single conflicted terminal of a state, together with what the policy decided about it. This is
// what a parser generator reports to the user as a shift/reduce or a reduce/reduce conflict.
type Conflict struct {
	// StateIdx is the state index of the conflicted state.
	StateIdx int

	// TerminalIdx is the terminal index of the conflicted terminal.
	TerminalIdx int

	// Contributions are the actions which competed for the conflicted terminal, before the policy decided between them.
	// This is the value of the contributions function of definition 2.17 of IELR(1).
	Contributions ContributionSet

	// Decision is what the policy decided about the conflict.
	Decision Decision
}

// Resolve applies the policy to every conflict of the parser tables, removes the losing actions from them, and returns
// every conflict which was found. This is phase 5 of IELR(1), which section 3.7 of the paper describes, and it is what
// makes parser tables with conflicts usable for a parser.
//
// The parser tables are modified in place. Resolving the conflicts is the last thing which happens to the parser tables
// before a backend serializes them, and the tables with their conflicts still in them are of no use to anyone
// afterwards, so there is nothing to be gained from copying every state of what can be a very large set of tables.
// Everything the caller needs to know about the conflicts is in the conflicts which are returned, including the actions
// which competed before the policy removed any of them.
//
// Every conflict is returned, whether it was resolved or not, because a parser generator reports the conflicts of a
// grammar to the user even when it decided them on its own. When a single contribution wins, the losing actions are
// removed, and when the terminal is rejected, every action is removed.
//
// The error reports the conflicts which the policies did not decide, joined into a single error, one
// UnresolvedConflictError per conflict, see there. Those conflicts keep the actions of the contributions they were left
// with, so the parser tables still hold them, which leaves the state with more than one action for the conflicted
// terminal. A parser cannot be generated from such parser tables, because it would not know which of the actions to
// take, so the caller has to give up on the grammar instead of handing the tables to a backend.
func Resolve(parser *backend.Parser, policy Policy) ([]Conflict, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Conflict: Resolve").End()

	var conflicts []Conflict
	var errs []error
	for stateIdx := range parser.States {
		// The conflicts of the state are collected before any action is removed from it, because removing the actions
		// which lost is what makes the state stop being conflicted.
		stateConflicts := getConflicts(parser.States[stateIdx], stateIdx)
		resolveState(&parser.States[stateIdx], stateConflicts, policy)

		for _, stateConflict := range stateConflicts {
			if stateConflict.Decision.Kind != DecisionUnresolved {
				continue
			}
			errs = append(errs, UnresolvedConflictError{
				Conflict:     stateConflict,
				TerminalName: parser.Grammar.Terminals[stateConflict.TerminalIdx].String(),
			})
		}
		conflicts = append(conflicts, stateConflicts...)
	}
	return conflicts, errors.Join(errs...)
}

// getConflicts returns every conflict of the state, which are the terminals the state has more than one action for. The
// decision of the conflicts is not filled in yet, that is what resolveState does.
func getConflicts(state backend.State, stateIdx int) []Conflict {
	contributionsByTerminalIdx := ContributionsByTerminalIdx(state)

	var result []Conflict
	// The keys of a map come in no particular order, but the conflicts end up in a report for the user, so we want them
	// to be stable across runs.
	for _, terminalIdx := range slices.Sorted(maps.Keys(contributionsByTerminalIdx)) {
		contributions := contributionsByTerminalIdx[terminalIdx]
		if contributions.Length() <= 1 {
			// The terminal has a single action only, so there is nothing which competes for it.
			continue
		}
		result = append(result, Conflict{
			StateIdx:      stateIdx,
			TerminalIdx:   terminalIdx,
			Contributions: contributions,
		})
	}
	return result
}

// resolveState asks the policy about every conflict of the state and removes the losing actions from the state. The
// decision of every conflict is filled in along the way.
func resolveState(state *backend.State, conflicts []Conflict, policy Policy) {
	for i := range conflicts {
		conflicts[i].Decision = DominantContribution(policy, conflicts[i].TerminalIdx, conflicts[i].Contributions)

		survivors := conflicts[i].Decision.Survivors()
		for _, contribution := range conflicts[i].Contributions.All() {
			if survivors.Contains(contribution) {
				// This contribution survived the conflict, so its action stays. That is the single contribution which
				// won, or one of several an unresolved conflict was left with.
				continue
			}
			// Every other contribution lost the conflict. When the terminal is rejected, that is every contribution
			// there is, which leaves the state without an action for the terminal and makes the parser report an error
			// when it sees it.
			removeContribution(state, contribution, conflicts[i].TerminalIdx)
		}
	}
}

// removeContribution removes the action which the contribution describes from the state, so that the state can no
// longer take it on the conflicted terminal.
//
// A shift is a transition on the conflicted terminal, which is removed as a whole. A reduction is one terminal of the
// lookahead set of a reduce action, so the reduce action is replaced by one which does not reduce on the conflicted
// terminal anymore. A reduce action which is left without any terminal to reduce on is removed entirely.
func removeContribution(state *backend.State, contribution Contribution, terminalIdx int) {
	if contribution.IsShiftAction() {
		removeShiftAction(state, terminalIdx)
		return
	}

	reduceAction, ok := getReduceAction(state, contribution.ProductionIdx())
	if !ok {
		// The reduce action is gone already, which happens when the reduction lost the conflict on more than one
		// terminal and the last of them left it without a terminal to reduce on.
		return
	}
	state.ReduceActions.Remove(reduceAction)

	reduceAction.LookaheadSet.Remove(terminalIdx)
	if reduceAction.LookaheadSet.IsEmpty() {
		// The reduction lost on every terminal it could reduce on, so there is nothing left to add back.
		return
	}
	state.ReduceActions.Add(reduceAction)
}

// removeShiftAction removes the transition action which shifts the terminal from the state.
func removeShiftAction(state *backend.State, terminalIdx int) {
	symbolRef := frontend.NewTerminalRef(terminalIdx)
	for _, transitionAction := range state.TransitionActions.All() {
		if transitionAction.SymbolRef() != symbolRef {
			continue
		}
		state.TransitionActions.Remove(transitionAction)
		return
	}
	utils.DebugAssert(func() error {
		// The shift contribution was derived from a transition of this state, and a state cannot have two transitions
		// on the same terminal, so the transition we are looking for is there and is removed exactly once.
		return errors.New("state does not have a transition on the conflicted terminal")
	})
}

// getReduceAction returns the reduce action of the state which reduces the production. The second return value reports
// if the state still has such a reduce action.
func getReduceAction(state *backend.State, productionIdx int) (backend.ReduceAction, bool) {
	for _, reduceAction := range state.ReduceActions.All() {
		if reduceAction.ProductionIdx == productionIdx {
			return reduceAction, true
		}
	}
	return backend.ReduceAction{}, false
}
