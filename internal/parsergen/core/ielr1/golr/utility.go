package golr

import "github.com/backbone81/golr/internal/parsergen/backend"

// SliceView is a helper struct which describes an offset into a slice and a length. It is used to work with growing
// slices where we want to keep a reference to some subsection even after the underlying array has grown.
type SliceView struct {
	Offset int
	Length int
}

// SliceFromView returns a slice as a subset of a bigger slice. The extent of the subset is specified by the view.
func SliceFromView[T any](slice []T, view SliceView) []T {
	return slice[view.Offset : view.Offset+view.Length]
}

type ReduceActionRecord struct {
	StateIdx     int
	Core         backend.Core
	LookaheadSet backend.LookaheadSet
}

type TransitionRecord struct {
	FromStateIdx int
	SymbolIdx    int
	ToStateIdx   int
}

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
