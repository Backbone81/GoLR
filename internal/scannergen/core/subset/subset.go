package thompsons

import (
	"context"
	"runtime/trace"

	"github.com/backbone81/golr/internal/scannergen/backend"
	"github.com/backbone81/golr/internal/scannergen/core/subset/dfa"
	"github.com/backbone81/golr/internal/scannergen/core/subset/nfa"
	"github.com/backbone81/golr/internal/scannergen/frontend"
)

// RulesToDFA creates a deterministic finite automaton from a set of rules.
func RulesToDFA(rules []frontend.Rule) backend.DFA {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Cores: Thompsons: RulesToDFA").End()

	startNFA := nfa.RulesToNFA(rules)
	intermediateDFA := dfa.NewSubsetConstruction(startNFA).Build()
	minimalDFA := dfa.NewHopcroftsAlgorithm().Build(intermediateDFA)
	return backend.DFA{
		Rules:  rules,
		States: minimalDFA,
	}
}
