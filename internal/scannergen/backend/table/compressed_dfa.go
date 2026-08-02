package table

import (
	"github.com/backbone81/golr/internal/scannergen/backend"
	"github.com/backbone81/golr/internal/utils"
)

// NoRule is the entry used for a state which does not accept any rule. Rule indices are never negative, so this can not
// be confused with accepting the rule with index 0.
const NoRule = -1

// CompressedDFA is a DFA in the form a table driven scanner reads it at runtime. It holds the transitions of the DFA as
// lookup tables instead of as control flow, which is what allows a backend for a new language to be a small driver
// instead of another encoding of the automaton.
//
// The transitions are compressed in two steps. The byte equivalence classes turn the 256 possible input bytes into the
// far smaller number of columns the automaton actually distinguishes, and the row displacement then packs the resulting
// rows into a single array. Both steps are lossless, so a lookup returns exactly what walking the DFA states returns.
type CompressedDFA struct {
	// ByteClasses maps an input byte to the column of the transition table which holds the transition for it.
	ByteClasses ByteClasses

	// Transitions holds the transitions of all states packed into a single array, indexed by state and byte class.
	// An entry is the index of the state to transition to, and a column without an entry means the state has no
	// transition on the bytes of that class.
	Transitions utils.RowDisplacement

	// AcceptRuleIdxByStateIdx holds, for every state, the index of the rule the state accepts, or NoRule when the
	// state does not accept. This is the bookkeeping the maximal munch needs: whenever the scanner enters a state
	// which accepts, it remembers the rule and the position as the longest match found so far.
	AcceptRuleIdxByStateIdx []int
}

// NewCompressedDFA compresses the given DFA into the tables a table driven scanner uses.
func NewCompressedDFA(dfa backend.DFA) CompressedDFA {
	transitionTable := NewTransitionTable(dfa)

	acceptRuleIdxByStateIdx := make([]int, len(dfa.States))
	for stateIdx := range dfa.States {
		acceptRuleIdxByStateIdx[stateIdx] = NoRule
		if dfa.States[stateIdx].Accept {
			acceptRuleIdxByStateIdx[stateIdx] = dfa.States[stateIdx].RuleIdx
		}
	}

	return CompressedDFA{
		ByteClasses:             transitionTable.ByteClasses,
		Transitions:             utils.NewRowDisplacement(transitionTable.Rows, NoTransition),
		AcceptRuleIdxByStateIdx: acceptRuleIdxByStateIdx,
	}
}

// StateCount returns the number of states of the DFA.
func (c *CompressedDFA) StateCount() int {
	return len(c.AcceptRuleIdxByStateIdx)
}

// Transition returns the state the given state transitions to on the given byte, or NoTransition if it has no
// transition on that byte. This is the access code a generated table driver performs.
func (c *CompressedDFA) Transition(stateIdx int, byteValue byte) int {
	return c.Transitions.Lookup(stateIdx, c.ByteClasses.ClassByByte[byteValue])
}
