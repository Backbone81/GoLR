package nfa

import (
	"context"
	"golr/internal/scannergen/frontend"
	"runtime/trace"
)

// State is a single NFA state.
type State struct {
	// RuleIdx is holding the index for the rule this state is part of.
	RuleIdx int

	// Accept reports if this state is an accepting state for the rule given with RuleIdx.
	Accept bool

	// Transitions are the transitions to other NFA states.
	Transitions []Transition
}

// Transition is a single transition on a character range to the next state.
type Transition struct {
	// Empty is true in situations where this is an empty transition.
	Empty bool

	// CharRange describes the characters on which to use this transition.
	CharRange frontend.CharRange

	// NextStateIdx is the target state to transition to.
	NextStateIdx int
}

// FromRegex constructs an NFA from a regular expression.
func FromRegex(regexNode *frontend.Node, ruleIdx int) []State {
	defer trace.StartRegion(context.TODO(), "golr/internal/scannergen/nfa/FromRegex()").End()

	return NewThompsonsConstruction().Build(regexNode, ruleIdx)
}

// RulesToNFA is a helper function creating one combined NFA from all rules provided.
func RulesToNFA(rules []frontend.Rule) []State {
	nfas := make([][]State, 0, len(rules))
	for ruleIdx, rule := range rules {
		ruleNFA := FromRegex(&rule.Regex, ruleIdx)
		nfas = append(nfas, ruleNFA)
	}
	return Merge(nfas...)
}
