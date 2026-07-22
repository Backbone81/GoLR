package conflict

import "slices"

// SplitStability is the shared bookkeeping the policies of a compound policy fill in to decide whether the dominant
// contribution of a conflict is split-stable. This is the general case of definition 3.35 of IELR(1).
//
// Phase 2 of IELR(1) discards an annotation whose dominant contribution is split-stable, because splitting the
// annotated state cannot change which contribution dominates the conflict, see the golr core's
// Annotation.IsSplitStable.
// The contributions of the conflict are partitioned into the ones every isocore of the state makes (the always
// contributions) and the ones only some isocores make (the potential contributions); a never contribution, which no
// isocore makes, is left out entirely. The dominant contribution is split-stable when the policy decides the conflict
// the same way no matter which of the potential contributions an isocore happens to make.
//
// Deciding that by evaluating the policy on every subset of the potential contributions would be exponential. Instead
// the policies fill in this bookkeeping in a single pass: each policy narrows remaining exactly as its Resolve narrows
// its candidates, and clears stable whenever a narrowing it made depended on a potential contribution being present. A
// narrowing anchored by always contributions holds for every isocore, so it does not threaten split stability; a
// narrowing which hinges on a potential contribution reverses in the isocores which do not make that contribution, so
// the dominant contribution is not split-stable.
//
// Because a compound policy applies its policies to the same bookkeeping in order, a later policy sees what the earlier
// policies left in remaining. That shared view is what lets the policies decide together what none of them could decide
// on its own: the split stability of a compound policy does not follow from the split stability of its policies in
// isolation, because one policy removing a contribution changes what the next policy resolves.
type SplitStability struct {
	// remaining is the set of contributions still in play. It starts as the always and potential contributions together,
	// which is the maximal isocore that makes every potential contribution (the set Gamma' of definition 3.35), and each
	// policy narrows it just like Resolve narrows its candidates.
	remaining ContributionSet

	// always are the contributions every isocore makes. A contribution which is in remaining but not in always is a
	// potential contribution, which only some isocores make.
	always ContributionSet

	// stable stays true while every narrowing so far was anchored by always contributions. Once a policy makes a
	// narrowing which depends on a potential contribution, the dominant contribution can change when the state is split,
	// so stable is cleared and never set again.
	stable bool
}

// NewSplitStability creates the bookkeeping for a conflict whose contributions are the always contributions together
// with the potential contributions. This is the set Gamma' of definition 3.35, the contributions of the maximal
// isocore, with the never contributions already left out. The policies of the resolving policy fill it in with
// ContributeSplitStability before IsSplitStable reads the verdict.
func NewSplitStability(always ContributionSet, potentials ContributionSet) SplitStability {
	remaining := always.Clone()
	remaining.Merge(&potentials)
	return SplitStability{
		remaining: remaining,
		always:    always.Clone(),
		stable:    true,
	}
}

// IsSplitStable reports whether the bookkeeping the policies filled in describes a split-stable dominant contribution.
//
// Two conditions must hold. First, every narrowing the policies made must have been anchored by always contributions,
// which stable tracks: a narrowing which depended on a potential contribution reverses in the isocores which do not
// make that contribution. Second, no potential contribution may survive into remaining: a potential contribution which
// is still part of the decision changes that decision in the isocores which do not make it, whether it is the single
// dominant contribution or one of several the conflict was left undecided between.
func (s *SplitStability) IsSplitStable() bool {
	if !s.stable {
		return false
	}
	for _, contribution := range s.remaining.All() {
		if !s.always.Contains(contribution) {
			return false
		}
	}
	return true
}

// markUnstable records that a narrowing depended on a potential contribution, so the dominant contribution is not
// split-stable. It is called by the policies while they fill in the bookkeeping.
func (s *SplitStability) markUnstable() {
	s.stable = false
}

// isAlways reports whether the contribution is made by every isocore of the state. A contribution which is not always
// is a potential contribution. It is a convenience for the policies while they fill in the bookkeeping.
func (s *SplitStability) isAlways(contribution Contribution) bool {
	return s.always.Contains(contribution)
}

// anyAlways reports whether any of the contributions is an always contribution. A narrowing which several contributions
// cause holds for every isocore as long as at least one of them is an always contribution, because that one is present
// in every isocore, so this is how a policy decides whether such a narrowing is split-stable.
func (s *SplitStability) anyAlways(contributions []Contribution) bool {
	return slices.ContainsFunc(contributions, s.isAlways)
}
