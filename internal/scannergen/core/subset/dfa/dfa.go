package dfa

import (
	"context"
	"runtime/trace"

	"github.com/backbone81/golr/internal/scannergen/backend"
	thompsonsnfa "github.com/backbone81/golr/internal/scannergen/core/subset/nfa"
)

// FromNFA constructs a DFA from the NFA given as parameter.
func FromNFA(inputNFA []thompsonsnfa.State) []backend.State {
	defer trace.StartRegion(
		context.TODO(),
		"github.com/backbone81/golr/internal/scannergen/core/subset/dfa/FromNFA()",
	).End()

	intermediateDFA := NewSubsetConstruction(inputNFA).Build()
	minimalDFA := NewHopcroftsAlgorithm().Build(intermediateDFA)
	return minimalDFA
}
