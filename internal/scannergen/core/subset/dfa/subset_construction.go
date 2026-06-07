package dfa

import (
	"context"
	"runtime/trace"
	"slices"

	"github.com/backbone81/golr/internal/scannergen/backend"
	thompsonsnfa "github.com/backbone81/golr/internal/scannergen/core/subset/nfa"
	"github.com/backbone81/golr/internal/utils"
)

// SubsetConstruction is responsible for building the DFA from an NFA.
// It is an implementation of the Subset Construction. The main idea here is to create sets of NFA states and
// transitioning to the next set of NFA states on the available character ranges. At the end, each unique set
// of NFA states corresponds to a DFA state of the final DFA.
type SubsetConstruction struct {
	// nfaStates holds the NFA states which are transformed to DFA states by the subset construction.
	nfaStates []thompsonsnfa.State

	// dfaStates holds the DFA states which are the result of the subset construction.
	dfaStates []backend.State

	// nfaStateIdxsByDfaStateIdx maps a DFA state index to a set of NFA state indexes
	nfaStateIdxsByDfaStateIdx []utils.OrderedSet[int]

	// Comparing sets of NFA states for equality is quite expensive. Therefore, we are hashing the indexes
	// of a sorted list of NFA states and store those states under their hash. This is significant faster than
	// comparing every state in a set. As we need to accommodate for hash collisions, we have a slice of DFA state
	// indexes as the entry for a hash. But we assume that in most situations the slice will only contain a single
	// entry.
	dfaStateIdxsByNfaStateIdxsHash map[uint64][]int

	// This is a helper variable which is used during closure calculation to hold the unprocessed states. It does not
	// hold information on the long run. We have this variable here to reduce the amount of memory allocations only.
	unprocessedNfaStateIdxs utils.DynamicRingBuffer[int]
}

// NewSubsetConstruction creates a new builder instance.
func NewSubsetConstruction(inputNFA []thompsonsnfa.State) *SubsetConstruction {
	return &SubsetConstruction{
		nfaStates:                      inputNFA,
		dfaStateIdxsByNfaStateIdxsHash: make(map[uint64][]int),
		unprocessedNfaStateIdxs:        utils.NewDynamicRingBuffer[int](),
	}
}

// Build creates a new DFA from the given NFA.
func (b *SubsetConstruction) Build() []backend.State {
	defer trace.StartRegion(
		context.TODO(),
		"github.com/backbone81/golr/internal/scannergen/dfa/SubsetConstruction.Build()",
	).End()

	startNfaStateIdxs := b.EmptyClosure(utils.NewOrderedSet(0))
	startDfaStateIdx := b.addState(startNfaStateIdxs)

	unprocessedDfaStateIdxs := utils.NewDynamicRingBuffer[int]()
	unprocessedDfaStateIdxs.Add(startDfaStateIdx)
	for !unprocessedDfaStateIdxs.IsEmpty() {
		currDfaStateIdx := unprocessedDfaStateIdxs.Remove()

		charRanges := b.GetByteRanges(b.nfaStateIdxsByDfaStateIdx[currDfaStateIdx])
		for _, charRange := range charRanges {
			nfaStateIdxs := b.EmptyClosure(b.transitionOnByteRange(b.nfaStateIdxsByDfaStateIdx[currDfaStateIdx], charRange))
			nextDfaStateIdx, found := b.getState(nfaStateIdxs)
			if !found {
				nextDfaStateIdx = b.addState(nfaStateIdxs)
				unprocessedDfaStateIdxs.Add(nextDfaStateIdx)
			}
			b.dfaStates[currDfaStateIdx].Transitions = append(b.dfaStates[currDfaStateIdx].Transitions, backend.Transition{
				ByteRange: charRange,
				StateIdx:  nextDfaStateIdx,
			})
		}
	}
	b.sortTransitions()

	return b.dfaStates
}

// sortTransitions sorts the transitions on all DFA states in an increasing order.
func (b *SubsetConstruction) sortTransitions() {
	for _, dfaState := range b.dfaStates {
		slices.SortStableFunc(dfaState.Transitions, func(a, b backend.Transition) int {
			switch {
			case a.ByteRange.Low < b.ByteRange.Low:
				return -1
			case a.ByteRange.Low > b.ByteRange.Low:
				return 1
			default:
				return 0
			}
		})
	}
}

// addState adds the DFA state to the list of known states.
func (b *SubsetConstruction) addState(nfaStateIdxs utils.OrderedSet[int]) int {
	b.dfaStates = append(b.dfaStates, backend.State{})
	dfaStateIdx := len(b.dfaStates) - 1

	highestPrioNFAState, isAccepting := b.getHighestPrioNFAState(nfaStateIdxs)
	b.dfaStates[dfaStateIdx].RuleIdx = b.nfaStates[highestPrioNFAState].RuleIdx
	b.dfaStates[dfaStateIdx].Accept = isAccepting

	b.nfaStateIdxsByDfaStateIdx = append(b.nfaStateIdxsByDfaStateIdx, nfaStateIdxs)

	hash := nfaStateIdxs.Hash()
	b.dfaStateIdxsByNfaStateIdxsHash[hash] = append(b.dfaStateIdxsByNfaStateIdxsHash[hash], dfaStateIdx)

	return dfaStateIdx
}

// getState returns the DFA state index corresponding to the set of NFA state indexes provided. The boolean return value
// indicates if the DFA state index was found or not.
func (b *SubsetConstruction) getState(nfaStateIdxs utils.OrderedSet[int]) (int, bool) {
	hash := nfaStateIdxs.Hash()
	bucket := b.dfaStateIdxsByNfaStateIdxsHash[hash]
	if len(bucket) == 1 {
		return bucket[0], true
	} else {
		for _, dfaStateIdx := range bucket {
			if b.nfaStateIdxsByDfaStateIdx[dfaStateIdx].Equal(&nfaStateIdxs) {
				return dfaStateIdx, true
			}
		}
	}
	return 0, false
}

// getHighestPrioNFAState returns the highest prio NFA state. When no accepting state is present in the NFA state list,
// the priority of all states is considered. If there is at least one accepting state, only the priority of accepting
// states is considered. This is necessary to make sure that a state which is accepting but also intermediate has
// priority on accepting. Otherwise, wrong tokens are recognized. The boolean return value signal if it was an accepting
// state.
func (b *SubsetConstruction) getHighestPrioNFAState(nfaStateIdxs utils.OrderedSet[int]) (int, bool) {
	highestPrioStandardStateIdx := -1
	highestPrioAcceptingStateIdx := -1
	for _, nfaStateIdx := range nfaStateIdxs.All() {
		if b.nfaStates[nfaStateIdx].Accept {
			if highestPrioAcceptingStateIdx == -1 ||
				b.nfaStates[nfaStateIdx].RuleIdx < b.nfaStates[highestPrioAcceptingStateIdx].RuleIdx {
				highestPrioAcceptingStateIdx = nfaStateIdx
			}
		} else {
			if highestPrioStandardStateIdx == -1 ||
				b.nfaStates[nfaStateIdx].RuleIdx < b.nfaStates[highestPrioStandardStateIdx].RuleIdx {
				highestPrioStandardStateIdx = nfaStateIdx
			}
		}
	}
	if highestPrioAcceptingStateIdx != -1 {
		return highestPrioAcceptingStateIdx, true
	}
	return highestPrioStandardStateIdx, false
}

// transitionOnByteRange moves from the given states on the byte range to the next states.
func (b *SubsetConstruction) transitionOnByteRange(
	nfaStateIdxs utils.OrderedSet[int],
	byteRange backend.ByteRange,
) utils.OrderedSet[int] {
	var result utils.OrderedSet[int]
	for _, nfaStateIdx := range nfaStateIdxs.All() {
		for _, transition := range b.nfaStates[nfaStateIdx].Transitions {
			if transition.Empty {
				continue
			}
			if byteRange.Low < transition.ByteRange.Low {
				continue
			}
			if byteRange.High > transition.ByteRange.High {
				continue
			}
			result.Add(transition.NextStateIdx)
		}
	}
	return result
}

// EmptyClosure extends the given list of states by those which are reachable through empty transitions from the given
// states.
func (b *SubsetConstruction) EmptyClosure(nfaStateIdxs utils.OrderedSet[int]) utils.OrderedSet[int] {
	for _, nfaStateIdx := range nfaStateIdxs.All() {
		b.unprocessedNfaStateIdxs.Add(nfaStateIdx)
	}
	for !b.unprocessedNfaStateIdxs.IsEmpty() {
		currStateIdx := b.unprocessedNfaStateIdxs.Remove()

		for _, transition := range b.nfaStates[currStateIdx].Transitions {
			if !transition.Empty {
				// We only care for empty transitions
				continue
			}

			if nfaStateIdxs.Add(transition.NextStateIdx) {
				b.unprocessedNfaStateIdxs.Add(transition.NextStateIdx)
			}
		}
	}
	return nfaStateIdxs
}

// GetByteRanges returns the character ranges which are responsible for non-empty transitions out of the given
// states.
func (b *SubsetConstruction) GetByteRanges(nfaStateIdxs utils.OrderedSet[int]) []backend.ByteRange {
	var result []backend.ByteRange
	for _, nfaStateIdx := range nfaStateIdxs.All() {
		for _, transition := range b.nfaStates[nfaStateIdx].Transitions {
			if transition.Empty {
				continue
			}
			result = backend.SplitByteRanges(result, transition.ByteRange.Low)
			result = backend.SplitByteRanges(result, transition.ByteRange.High+1)

			newByteRanges := []backend.ByteRange{transition.ByteRange}
			for _, characterRange := range result {
				newByteRanges = backend.SplitByteRanges(newByteRanges, characterRange.Low)
				newByteRanges = backend.SplitByteRanges(newByteRanges, characterRange.High+1)
				newByteRanges = backend.RemoveByteRanges(newByteRanges, characterRange)
			}

			result = append(result, newByteRanges...)
		}
	}
	return result
}
