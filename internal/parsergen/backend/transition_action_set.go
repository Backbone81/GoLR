package backend

import (
	"golr/internal/utils"
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
