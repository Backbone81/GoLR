package dfa

import (
	"context"
	"runtime/trace"

	"golr/internal/scannergen/backend"
	thompsonsnfa "golr/internal/scannergen/core/subset/nfa"
)

// FromNFA constructs a DFA from the NFA given as parameter.
func FromNFA(inputNFA []thompsonsnfa.State) []backend.State {
	defer trace.StartRegion(context.TODO(), "golr/internal/scannergen/core/subset/dfa/FromNFA()").End()

	intermediateDFA := NewSubsetConstruction(inputNFA).Build()
	minimalDFA := NewHopcroftsAlgorithm().Build(intermediateDFA)
	return minimalDFA
}
