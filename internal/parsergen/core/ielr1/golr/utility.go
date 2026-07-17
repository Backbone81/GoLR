package golr

import "github.com/backbone81/golr/internal/parsergen/backend"

// ReduceActionRecord is a reduce action of the automaton as the reduction lookahead builder works on it: the state the
// reduction happens in, the item which reduces there, and the reduction lookahead set the builder computes for it.
type ReduceActionRecord struct {
	// StateIdx is the state index of the state the reduction happens in.
	StateIdx int

	// Core is the item which reduces, with its position at the end of the production. It is what the builder traces
	// backward through the states to the gotos which generated it, whose goto follow sets make up the lookahead set.
	Core backend.Core

	// LookaheadSet is the reduction lookahead set. It is empty until the builder fills it in.
	LookaheadSet backend.LookaheadSet
}

// Edge is a single edge of a digraph relation between gotos, as consumed by the digraph algorithm. The direction
// follows the goto follows relations of IELR(1): the set of FromIdx depends on the set of ToIdx, so propagation merges
// the set of ToIdx into the set of FromIdx.
type Edge struct {
	FromIdx int
	ToIdx   int
}

type GotoRecord struct {
	// FromStateIdx is the state index a goto is moving from. This is "from_state" from IELR(1) definition 3.4.
	FromStateIdx int

	// ToStateIdx is the state index a goto is moving to. This is "to_state" from IELR(1) definition 3.4.
	ToStateIdx int

	// NonterminalIdx is the nonterminal index which is triggering this goto.
	NonterminalIdx int
}

type InternalDependencyCandidate struct {
	// GotoIdx is the goto index this candidate is for.
	GotoIdx int

	// NonterminalIdx is the nonterminal index which was on the left hand side of the production which created this
	// candidate.
	NonterminalIdx int
}

type BackwardTransitionInfo struct {
	TerminalTransitions    map[int][]int
	NonterminalTransitions map[int][]int
}

func NewBackwardTransitionInfo() BackwardTransitionInfo {
	return BackwardTransitionInfo{
		TerminalTransitions:    make(map[int][]int),
		NonterminalTransitions: make(map[int][]int),
	}
}

type PredecessorDependencyCandidate struct {
	// GotoIdx is the goto index this candidate is for.
	GotoIdx int

	// This is the item core which created predecessor dependency candidate. This item needs to be followed back to the
	// state which created it for the dependency to be created.
	Core backend.Core
}
