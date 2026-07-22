package golr

import "github.com/backbone81/golr/internal/utils"

// ContributionRow describes how any isocore split from a state can make a single contribution of an inadequacy. This
// is one entry of the inadequacy contribution matrix of point 2 of definition 3.28 of IELR(1).
type ContributionRow struct {
	// Defined reports if the contribution row holds a meaningful set of kernel items. When it is false, the row is
	// undefined in terms of point 2a of definition 3.28, which means the contribution is an always contribution: every
	// isocore which can be split from the annotated state is guaranteed to make the contribution, no matter what its
	// kernel item lookahead sets are.
	Defined bool

	// KernelItems holds the kernel item indexes of the annotated state on which the contribution depends. An isocore
	// split from the annotated state makes the contribution if the conflicted terminal appears in the lookahead set of
	// any of these kernel items in that isocore. KernelItems is only meaningful when Defined is true.
	KernelItems utils.Bitset
}

// IsAlways reports if the contribution is an always contribution, which every isocore split from the annotated state
// is guaranteed to make. This is point 2a of definition 3.28 of IELR(1).
func (r ContributionRow) IsAlways() bool {
	return !r.Defined
}

// IsNever reports if the contribution is a never contribution, which no isocore split from the annotated state can
// make. This is point 2(b)ii of definition 3.28 of IELR(1).
func (r ContributionRow) IsNever() bool {
	return r.Defined && r.KernelItems.IsEmpty()
}

// IsPotential reports if the contribution is a potential contribution, which an isocore split from the annotated state
// makes depending on the lookahead sets of its kernel items. This is point 2(b)i of definition 3.28 of IELR(1).
func (r ContributionRow) IsPotential() bool {
	return r.Defined && !r.KernelItems.IsEmpty()
}

// Equal reports if both contribution rows are the same.
func (r ContributionRow) Equal(other ContributionRow) bool {
	if r.Defined != other.Defined {
		return false
	}
	if !r.Defined {
		return true
	}
	return r.KernelItems.Equal(other.KernelItems)
}
