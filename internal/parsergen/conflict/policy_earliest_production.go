package conflict

// EarliestProductionPolicy resolves a conflict between two reductions in favor of the production which was declared
// first in the grammar. It is the rule of last resort, because it always decides between reductions, which makes it the
// policy which guarantees that a compound policy leaves no reduce/reduce conflict unresolved.
//
// A shift among the candidates is left untouched, so this policy alone does not resolve a shift/reduce conflict.
type EarliestProductionPolicy struct{}

// EarliestProductionPolicy implements Policy.
var _ Policy = (*EarliestProductionPolicy)(nil)

// NewEarliestProductionPolicy returns the policy which resolves a reduce/reduce conflict in favor of the production
// which was declared first.
func NewEarliestProductionPolicy() *EarliestProductionPolicy {
	return &EarliestProductionPolicy{}
}

// Resolve removes every reduction but the one on the production with the lowest production index.
func (p *EarliestProductionPolicy) Resolve(terminalIdx int, candidates ContributionSet) ContributionSet {
	var result ContributionSet
	earliestFound := false
	for _, candidate := range candidates.All() {
		if candidate.IsShiftAction() {
			result.Add(candidate)
			continue
		}
		if earliestFound {
			// The candidates are ordered by their production index, so every reduction after the first one we see is on
			// a production which was declared later.
			continue
		}
		result.Add(candidate)
		earliestFound = true
	}
	return result
}

// ContributeSplitStability defers the narrowing to Resolve and only decides whether that narrowing is split-stable.
//
// Removing the later reductions is anchored by the earliest reduction, so those reductions being potential does not
// threaten split stability: they lose the reduce/reduce comparison whether or not an isocore makes them. The reduction
// which survives is the earliest one, and it winning is only split-stable when it is an always contribution: if it is
// potential, an isocore which does not make it reduces on a later production instead, or does not reduce at all, so the
// dominant contribution changes.
func (p *EarliestProductionPolicy) ContributeSplitStability(terminalIdx int, splitStability *SplitStability) {
	splitStability.remaining = p.Resolve(terminalIdx, splitStability.remaining)
	for _, contribution := range splitStability.remaining.All() {
		if contribution.IsReduceAction() && !splitStability.isAlways(contribution) {
			splitStability.markUnstable()
		}
	}
}
