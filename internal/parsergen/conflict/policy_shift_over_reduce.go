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

// ContributeSplitStability defers the narrowing to Resolve and only decides whether that narrowing is split-stable.
//
// Removing the reductions is anchored by the shift, so a reduction being potential does not threaten split stability:
// every isocore which makes the shift resolves in its favor, whether or not the isocore also makes the reduction. The
// shift itself is an always contribution in every conflict, because splitting a state keeps its transitions, so the
// narrowing holds for every isocore. Should a shift ever be a potential contribution, the isocores which do not make it
// keep the reductions, so the narrowing would no longer hold and the bookkeeping is marked unstable.
func (p *ShiftOverReducePolicy) ContributeSplitStability(terminalIdx int, splitStability *SplitStability) {
	shift := NewShiftContribution()
	if splitStability.remaining.Contains(shift) && !splitStability.isAlways(shift) {
		splitStability.markUnstable()
	}
	splitStability.remaining = p.Resolve(terminalIdx, splitStability.remaining)
}
