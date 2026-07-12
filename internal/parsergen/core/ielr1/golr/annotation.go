package golr

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
// observation 3.33 and observation 3.34 of IELR(1).
//
// Observation 3.33 is the special case of definition 3.35 which holds when all contributions are always contributions
// or never contributions. In that case every isocore which can be split from the annotated state makes exactly the
// same contributions, so splitting the state cannot change which contribution dominates the conflict. We do not
// implement the general case of definition 3.35, because it depends on the dominant contribution function, which is
// the conflict resolution of phase 5 and does not exist yet. Missing out on useless annotations does not sacrifice
// correctness, it only leaves phase 3 with more annotations to consider than strictly necessary.
func (a *Annotation) IsSplitStable() bool {
	for _, contributionRow := range a.ContributionMatrix {
		if contributionRow.IsPotential() {
			return false
		}
	}
	return true
}
