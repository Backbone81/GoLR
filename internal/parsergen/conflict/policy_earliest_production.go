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
