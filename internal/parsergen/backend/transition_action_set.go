package backend

import (
	"github.com/backbone81/golr/internal/utils"
)

// TransitionActionSet is an ordered set of transition actions.
type TransitionActionSet struct {
	utils.OrderedSet[TransitionAction]
}

// NewTransitionActionSet creates a new ordered transition action set.
func NewTransitionActionSet(values ...TransitionAction) TransitionActionSet {
	return TransitionActionSet{
		OrderedSet: utils.NewOrderedSet[TransitionAction](values...),
	}
}

// Clone returns a copy of the ordered set which shares no storage with the original. A plain copy of a transition
// action set keeps referencing the actions of the original, so removing an action from the copy would remove it from
// the original as well. Clone is what you want when the original must stay untouched.
func (s *TransitionActionSet) Clone() TransitionActionSet {
	return TransitionActionSet{
		OrderedSet: s.OrderedSet.Clone(),
	}
}
