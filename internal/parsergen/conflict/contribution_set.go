package conflict

import (
	"github.com/backbone81/golr/internal/utils"
)

// ContributionSet is the set of actions a state can take on a single terminal. This is the value of the contributions
// function of definition 2.17 of IELR(1). A set with more than one contribution describes a conflicted terminal as
// specified in definition 2.18.
//
// The contributions of a state are derived from its actions, which Scanner does for the conflicted terminals,
// the only ones anybody asks about.
type ContributionSet = utils.OrderedSet[Contribution]

// NewContributionSet creates a new set with the given contributions.
func NewContributionSet(contributions ...Contribution) ContributionSet {
	return utils.NewOrderedSet(contributions...)
}
