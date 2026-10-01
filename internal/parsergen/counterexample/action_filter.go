package counterexample

import (
	"slices"

	"github.com/backbone81/golr/internal/parsergen/conflict"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// stateTerminal is a terminal of a state, the key of the action filter.
type stateTerminal struct {
	stateIdx    int
	terminalIdx int
}

// IsActionAllowed reports if the search may take the action on the terminal in the state: a shift, or a reduction of
// the production of a reduce item of the state. The action has to exist, which for a reduction means that the lookahead
// set of its item holds the terminal. Where several actions compete for the terminal, the declarations of the grammar
// have to leave the action competing: it has to be among the actions a rule of last resort decided between or the
// policy left unresolved, or among the survivors of the decision when the declarations decided on their own. This is
// how the search inspects precedence while it runs, see section 6, "Exploiting precedence".
func (t *LookupTables) IsActionAllowed(stateIdx int, terminalIdx int, contribution conflict.Contribution) bool {
	key := stateTerminal{stateIdx: stateIdx, terminalIdx: terminalIdx}
	allowed, found := t.allowedContributionsByStateTerminal[key]
	if !found {
		allowed = t.allowedContributions(stateIdx, terminalIdx)
		t.allowedContributionsByStateTerminal[key] = allowed
	}
	return allowed.Contains(contribution)
}

// allowedContributions returns the actions the search may take on the terminal in the state, see IsActionAllowed.
func (t *LookupTables) allowedContributions(stateIdx int, terminalIdx int) conflict.ContributionSet {
	contributions := t.contributions(stateIdx, terminalIdx)
	if contributions.Length() < 2 {
		return contributions
	}
	decision, undeclared := conflict.DominantContribution(t.policy, terminalIdx, contributions)
	if !undeclared.IsEmpty() {
		return undeclared
	}
	return decision.Survivors()
}

// contributions returns the actions of the state on the terminal: the shift when the state has a transition on it, and
// the reduction of every reduce item whose lookahead set holds it.
func (t *LookupTables) contributions(stateIdx int, terminalIdx int) conflict.ContributionSet {
	var result conflict.ContributionSet
	if _, found := t.transitionStateIdx(stateIdx, frontend.NewTerminalRef(terminalIdx)); found {
		result.Add(conflict.NewShiftContribution())
	}

	for _, itemIdx := range t.reduceItems(stateIdx) {
		if t.Lookahead(itemIdx).Contains(terminalIdx) {
			result.Add(conflict.NewReduceContribution(t.coreByItemIdx[itemIdx].ProductionIdx()))
		}
	}
	return result
}

// reduceItems returns the items of the state with the dot at the end. The returned slice must not be modified.
func (t *LookupTables) reduceItems(stateIdx int) []int {
	itemIdxs := t.itemIdxsByNextSymbol[t.itemIdxOffsetByStateIdx[stateIdx]:t.itemIdxOffsetByStateIdx[stateIdx+1]]
	from, _ := slices.BinarySearchFunc(itemIdxs, noNextSymbolKey, t.compareNextSymbolKey)
	return itemIdxs[from:]
}
