package conflict

import (
	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/utils"
)

// ContributionSet is the set of actions a state can take on a single terminal. This is the value of the contributions
// function of definition 2.17 of IELR(1). A set with more than one contribution describes a conflicted terminal as
// specified in definition 2.18.
type ContributionSet = utils.OrderedSet[Contribution]

// NewContributionSet creates a new set with the given contributions.
func NewContributionSet(contributions ...Contribution) ContributionSet {
	return utils.NewOrderedSet(contributions...)
}

// ContributionsByTerminalIdx returns the actions the state can take, keyed by the terminal index the action happens on.
// This is the contributions function of definition 2.17 of IELR(1).
//
// The contributions are derived from the actions of the state, so this works on any parser table, no matter which
// algorithm computed it. A terminal whose contribution set holds more than one contribution is a conflicted terminal.
func ContributionsByTerminalIdx(state backend.State) map[int]ContributionSet {
	result := make(map[int]ContributionSet)
	for _, transition := range state.TransitionActions.All() {
		if transition.SymbolRef().IsNonterminal() {
			// A goto is not an action on a terminal, so it can never take part in a conflict.
			continue
		}
		contributions := result[transition.SymbolRef().Idx()]
		contributions.Add(NewShiftContribution())
		result[transition.SymbolRef().Idx()] = contributions
	}
	for _, reduction := range state.ReduceActions.All() {
		for terminalIdx := range reduction.LookaheadSet.All() {
			contributions := result[terminalIdx]
			contributions.Add(NewReduceContribution(reduction.ProductionIdx))
			result[terminalIdx] = contributions
		}
	}
	return result
}
