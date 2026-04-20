package backend

import "golr/internal/scannergen/frontend"

// Transition is a single transition on a character range to the next state.
type Transition struct {
	// CharRange describes the characters on which to use this transition
	CharRange frontend.CharRange `json:"charRange" yaml:"charRange"`

	// StateIdx is the target state to transition to.
	StateIdx int `json:"stateIdx" yaml:"stateIdx"`
}
