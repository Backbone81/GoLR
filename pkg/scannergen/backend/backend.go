package backend

import intbackend "golr/internal/scannergen/backend"

type (
	// DFA is a deterministic finite automata.
	DFA = intbackend.DFA

	// State is a single DFA state.
	State = intbackend.State

	// Transition is a single transition on a character range to the next state.
	Transition = intbackend.Transition
)
