package golr

import (
	"errors"

	"github.com/backbone81/golr/internal/parsergen/conflict"
	"github.com/backbone81/golr/internal/utils"
)

// Annotation describes whether and how any isocore which phase 3 might split from a state can make the contributions
// of the inadequacy it refers to. This is definition 3.28 of IELR(1).
type Annotation struct {
	// Inadequacy is the inadequacy this annotation was computed from. Annotations travel backwards along the lanes of
	// the conflicted state, so an annotation is regularly attached to a state which is not the conflicted state of
	// this inadequacy.
	Inadequacy *Inadequacy

	// ContributionMatrix describes how any isocore split from the annotated state contributes to the contributions of
	// the inadequacy.
	ContributionMatrix ContributionMatrix
}

// Equal reports if both annotations describe the same contributions for the same inadequacy. This is what point 2 of
// definition 3.29 of IELR(1) needs to detect that a state already carries an identical annotation, which terminates
// the reverse iteration along a lane.
func (a *Annotation) Equal(other *Annotation) bool {
	return a.Inadequacy == other.Inadequacy && a.ContributionMatrix.Equal(other.ContributionMatrix)
}

// IsSplitStable reports if the annotation specifies a split-stable dominant contribution, which makes the annotation
// useless. A useless annotation can be discarded, and the reverse iteration along the lane can be terminated. This is
// definition 3.35 together with observation 3.34 of IELR(1).
//
// The contributions of the inadequacy fall into three kinds, see definition 3.28: an always contribution is made by
// every isocore which phase 3 could split from the annotated state, a never contribution by none of them, and a
// potential contribution only by the isocores whose kernel item lookahead sets happen to contain the conflicted
// terminal. The dominant contribution is split-stable when the conflict resolution policy decides the conflict the same
// way no matter which of the potential contributions an isocore makes, because then splitting the state cannot change
// which contribution dominates.
//
// The policy decides that in a single pass over its rules with the SplitStability bookkeeping: the never contributions
// are left out, the always and potential contributions become the maximal isocore, and each policy of the compound
// records whether its narrowing depended on a potential contribution. Observation 3.33, where there are no potential
// contributions and the result is split-stable regardless of the policy, falls out of this as the case where no
// narrowing can depend on a potential contribution.
func (a *Annotation) IsSplitStable(policy conflict.Policy) bool {
	var always, potentials conflict.ContributionSet
	for contributionIdx, contributionRow := range a.ContributionMatrix {
		contribution := a.Inadequacy.Contributions.GetByIndex(contributionIdx)
		switch {
		case contributionRow.IsAlways():
			always.Add(contribution)
		case contributionRow.IsPotential():
			potentials.Add(contribution)
		default:
			// A never contribution is made by no isocore, so it is left out of the bookkeeping entirely.
		}
	}
	utils.DebugAssert(func() error {
		// The policies which fill in the bookkeeping rely on the shift being an always contribution, which point 1 of
		// definition 3.30 guarantees by making its contribution matrix row undefined. A potential shift would make the
		// split stability of the precedence policy wrong, so guard against it rather than discard annotations silently.
		for _, contribution := range potentials.All() {
			if contribution.IsShiftAction() {
				return errors.New("a shift must be an always contribution per definition 3.30 point 1, not a potential one")
			}
		}
		return nil
	})
	splitStability := conflict.NewSplitStability(always, potentials)
	policy.ContributeSplitStability(a.Inadequacy.TerminalIdx, &splitStability)
	return splitStability.IsSplitStable()
}
