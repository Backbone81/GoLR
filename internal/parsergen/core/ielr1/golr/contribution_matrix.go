package golr

// ContributionMatrix describes how any isocore split from a state can make each of the contributions of an
// inadequacy. It is indexed by the contribution index within the contributions of the inadequacy. This is the
// inadequacy contribution matrix of point 2 of definition 3.28 of IELR(1).
type ContributionMatrix []ContributionRow

// Equal reports if both contribution matrices are the same.
func (m ContributionMatrix) Equal(other ContributionMatrix) bool {
	if len(m) != len(other) {
		return false
	}
	for i := range m {
		if !m[i].Equal(other[i]) {
			return false
		}
	}
	return true
}
