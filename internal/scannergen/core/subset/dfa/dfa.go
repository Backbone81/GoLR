package dfa

import (
	"context"
	"golr/internal/scannergen/backend"
	thompsonsnfa "golr/internal/scannergen/core/subset/nfa"
	"runtime/trace"
)

// FromNFA constructs a DFA from the NFA given as parameter.
func FromNFA(inputNFA []thompsonsnfa.State) []backend.State {
	defer trace.StartRegion(context.TODO(), "golr/internal/scannergen/dfa/FromNFA()").End()

	intermediateDFA := NewSubsetConstruction(inputNFA).Build()
	minimalDFA := NewHopcroftsAlgorithm().Build(intermediateDFA)
	return minimalDFA
}
