package thompsons

import (
	"context"
	"golr/internal/scannergen/backend"
	"golr/internal/scannergen/core/subset/dfa"
	"golr/internal/scannergen/core/subset/nfa"
	"golr/internal/scannergen/frontend"
	"runtime/trace"
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
