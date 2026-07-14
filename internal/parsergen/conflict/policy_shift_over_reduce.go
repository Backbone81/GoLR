package conflict

// ShiftOverReducePolicy resolves a conflict between a shift and one or more reductions in favor of the shift. This is
// the rule which makes the dangling else of an if-then-else grammar bind to the innermost if, and it is what a parser
// generator without precedence declarations falls back to.
//
// A compound policy which does not hold this policy leaves a shift/reduce conflict to the policies behind it.
type ShiftOverReducePolicy struct{}

// ShiftOverReducePolicy implements Policy.
var _ Policy = (*ShiftOverReducePolicy)(nil)

// NewShiftOverReducePolicy returns the policy which resolves a shift/reduce conflict in favor of the shift.
func NewShiftOverReducePolicy() *ShiftOverReducePolicy {
	return &ShiftOverReducePolicy{}
}

// Resolve removes every reduction from the candidates when a shift is among them.
func (p *ShiftOverReducePolicy) Resolve(terminalIdx int, candidates ContributionSet) ContributionSet {
	shift := NewShiftContribution()
	if !candidates.Contains(shift) {
		// There is no shift which could win, so this is a conflict between reductions only, which this policy has
		// nothing to say about.
		return candidates
	}
	return NewContributionSet(shift)
}
