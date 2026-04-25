package dfa

import (
	"cmp"
	"context"
	"golr/internal/scannergen/backend"
	"golr/internal/utils"
	"runtime/trace"
	"slices"
)

// HopcroftsAlgorithm is responsible for creating a minimal DFA from a DFA created by the subset construction.
// It is an implementation of Hopcroft's Algorithm as described in the paper "AN n log n ALGORITHM FOR MINIMIZING
// STATES IN A FINITE AUTOMATON" by John Hopcroft (https://doi.org/10.1016/B978-0-12-417750-5.50022-1). The main idea
// here is to have partitions of DFA states which behave identical. Identical behavior means all DFA states in a
// partition have transitions on the same character ranges to DFA states which are all part of the same partition.
// The algorithm starts out with one partition of all accepting states and one partition with all states which are
// not accepting states. By iterating over all partitions and splitting off new partitions with states which do
// not behave identical, we reach at some point equilibrium and every partition at the end corresponds to a DFA
// state of the new minimal DFA.
type HopcroftsAlgorithm struct {
	// inputDFA is the DFA which needs to be minimized.
	inputDFA []backend.State

	// partitionForStateIdx keeps track of what state belongs to which partition. This is important to keep track of
	// to quickly decide if a target state of some transition belongs to the same partition as the reference target
	// state.
	partitionForStateIdx map[int]*Partition
}

// NewHopcroftsAlgorithm creates a new builder instance.
func NewHopcroftsAlgorithm() *HopcroftsAlgorithm {
	return &HopcroftsAlgorithm{
		partitionForStateIdx: make(map[int]*Partition, 1024),
	}
}

// Partition is a set of DFA states which should behave identical.
type Partition struct {
	// StateIdxs holds all DFA states which make up the partition.
	StateIdxs utils.OrderedSet[int]

	// FinalStateIdx is a helper attribute which holds the final DFA state during DFA construction.
	FinalStateIdx int
}

// Build creates a new DFA from the given DFA.
func (b *HopcroftsAlgorithm) Build(inputDFA []backend.State) []backend.State {
	defer trace.StartRegion(context.TODO(), "golr/internal/scannergen/dfa/HopcroftsAlgorithm.Build()").End()

	b.inputDFA = inputDFA
	clear(b.partitionForStateIdx)

	partitions := b.buildInitialPartitions()
	partitions = b.splitAllPartitionsOnBehavior(partitions)

	// We sort the partitions by their first states. This results in a more intuitive ordering of states as if it was
	// a breath first search.
	slices.SortFunc(partitions, func(a, b *Partition) int {
		return cmp.Compare(a.StateIdxs.GetByIndex(0), b.StateIdxs.GetByIndex(0))
	})

	return b.buildDFAFromPartitions(partitions)
}

// buildInitialPartitions creates the initial partitions which we want to run our refinement loop on. Note that
// we not only create partitions of the accepting states and all other states, we also split the accepting partition
// on the name of the DFA states, as we need to retain dedicated accepting states for all our tokens. If we were to
// combine accepting states of different tokens, the scanner would later return the wrong tokens.
func (b *HopcroftsAlgorithm) buildInitialPartitions() []*Partition {
	var acceptingPartition Partition
	var otherPartition Partition
	for idx, state := range b.inputDFA {
		if state.Accept {
			acceptingPartition.StateIdxs.Add(idx)
			b.partitionForStateIdx[idx] = &acceptingPartition
		} else {
			otherPartition.StateIdxs.Add(idx)
			b.partitionForStateIdx[idx] = &otherPartition
		}
	}

	acceptingPartitions := b.splitAllPartitionsByName([]*Partition{&acceptingPartition})

	var result []*Partition
	if !otherPartition.StateIdxs.IsEmpty() {
		// There are situations where the partition might be empty, because the DFA consists of accepting states only.
		// We only want to add that partition if we have states in that partition to prevent errors when we use the
		// first state as a reference state during refinement.
		result = append(result, &otherPartition)
	}
	return append(result, acceptingPartitions...)
}

// buildDFAFromPartitions creates a new DFA from the given partitions. Each partition gets its own new DFA state
// and transitions are copied over from the reference state of the partition.
func (b *HopcroftsAlgorithm) buildDFAFromPartitions(partitions []*Partition) []backend.State {
	// create one state for each partition, copy over the name of the reference state
	states := make([]backend.State, len(partitions))
	for partitionIdx, partition := range partitions {
		partitions[partitionIdx].FinalStateIdx = partitionIdx
		state := b.inputDFA[partition.StateIdxs.GetByIndex(0)]
		states[partitionIdx].RuleIdx = state.RuleIdx
		states[partitionIdx].Accept = state.Accept
	}

	// copy over all transitions from the reference state to the new state
	for partitionIdx, partition := range partitions {
		inputState := b.inputDFA[partition.StateIdxs.GetByIndex(0)]
		outputState := &states[partitionIdx]
		for _, transition := range inputState.Transitions {
			outputState.Transitions = append(outputState.Transitions, backend.Transition{
				CharRange: transition.CharRange,
				StateIdx:  b.partitionForStateIdx[transition.StateIdx].FinalStateIdx,
			})
		}
	}

	return states
}

// splitAllPartitionsOnBehavior continues splitting the given partitions on their behavior until no further
// splitting is possible.
func (b *HopcroftsAlgorithm) splitAllPartitionsOnBehavior(partitions []*Partition) []*Partition {
	partitionSplit := true
	for partitionSplit {
		partitionSplit = false
		for i := 0; i < len(partitions); i++ {
			newPartition := b.splitPartitionOnBehavior(partitions[i])
			if newPartition == nil {
				continue
			}
			partitions = append(partitions, newPartition)
			partitionSplit = true
		}
	}
	return partitions
}

// splitPartitionOnBehavior is splitting the given partition if we notice that there are states which do not behave in a similar
// way. We always take the first state of the partition provided as parameter as the reference state and check all other
// states against that. States which do not behave the same as the reference state are moved to a new partition which
// is returned as result. The partition provided as parameter is modified to have the states removed which were moved
// to the new partition. If all states of the partition behave identical to the reference state, nil is returned
// as a result to signal that splitting the partition was not necessary.
func (b *HopcroftsAlgorithm) splitPartitionOnBehavior(partition *Partition) *Partition {
	var newPartition Partition
	referenceState := partition.StateIdxs.GetByIndex(0)

	for i, stateIdx := range partition.StateIdxs.All() {
		if i == 0 {
			// the first state is the reference state and always stays with the original partition
			continue
		}
		if b.statesAreEquivalent(&b.inputDFA[stateIdx], &b.inputDFA[referenceState]) {
			continue
		}
		if newPartition.StateIdxs.Contains(stateIdx) {
			continue
		}
		newPartition.StateIdxs.Add(stateIdx)
		b.partitionForStateIdx[stateIdx] = &newPartition
	}

	if newPartition.StateIdxs.IsEmpty() {
		return nil
	}

	// Now we need to remove the states we moved over to the new partition from the old partition.
	for _, stateIdx := range newPartition.StateIdxs.All() {
		partition.StateIdxs.Remove(stateIdx)
	}
	return &newPartition
}

// statesAreEquivalent checks if two states are equivalent in regard to their behavior. It checks that the transitions
// match and that the target states of the transitions match the same partition.
func (b *HopcroftsAlgorithm) statesAreEquivalent(state *backend.State, referenceState *backend.State) bool {
	if len(state.Transitions) != len(referenceState.Transitions) {
		// a state with a different number of transitions is always different in behavior, it needs to be split
		return false
	}
	for _, transition := range state.Transitions {
		referenceTransition := referenceState.GetTransition(transition.CharRange)
		if referenceTransition == nil {
			// the transition does not exist on the reference state, it needs to be split
			return false
		}
		if b.partitionForStateIdx[transition.StateIdx] != b.partitionForStateIdx[referenceTransition.StateIdx] {
			// the transition targets a different partition, it needs to be split
			return false
		}
	}
	return true
}

// splitAllPartitionsByName continues splitting the given partitions on the name of the DFA states until no further
// splitting is possible.
func (b *HopcroftsAlgorithm) splitAllPartitionsByName(partitions []*Partition) []*Partition {
	partitionSplit := true
	for partitionSplit {
		partitionSplit = false
		for i := 0; i < len(partitions); i++ {
			newPartition := b.splitPartitionByRuleIdx(partitions[i])
			if newPartition == nil {
				continue
			}
			partitions = append(partitions, newPartition)
			partitionSplit = true
		}
	}
	return partitions
}

// splitPartitionByRuleIdx is splitting the given partition if we notice that there are states which have a different rule.
// We always take the first state of the partition provided as parameter as the reference state and check all other
// states against that. States which have a different rule as the reference state are moved to a new partition which
// is returned as result. The partition provided as parameter is modified to have the states removed which were moved
// to the new partition. If all states of the partition have the same rule as the reference state, nil is returned
// as a result to signal that splitting the partition was not necessary.
func (b *HopcroftsAlgorithm) splitPartitionByRuleIdx(partition *Partition) *Partition {
	var newPartition Partition
	referenceStateIdx := partition.StateIdxs.GetByIndex(0)
	for _, stateIdx := range partition.StateIdxs.All() {
		if b.inputDFA[referenceStateIdx].RuleIdx == b.inputDFA[stateIdx].RuleIdx {
			continue
		}
		if newPartition.StateIdxs.Contains(stateIdx) {
			continue
		}
		newPartition.StateIdxs.Add(stateIdx)
		b.partitionForStateIdx[stateIdx] = &newPartition
	}

	if newPartition.StateIdxs.IsEmpty() {
		return nil
	}

	// Now we need to remove the states we moved over to the new partition from the old partition.
	for _, stateIdx := range newPartition.StateIdxs.All() {
		partition.StateIdxs.Remove(stateIdx)
	}
	return &newPartition
}
