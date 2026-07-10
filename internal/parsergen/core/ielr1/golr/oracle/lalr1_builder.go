package oracle

import (
	"fmt"
	"maps"
	"slices"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// LALR1FromLR1 converts a canonical LR(1) parser table into the LALR(1) parser table it corresponds to, by merging all
// states which agree on their kernel items and taking the union of the lookahead sets of their reduce actions.
//
// The result is an LALR(1) parser table like any other, which is what makes it usable as an oracle: it can be compared
// with Diff against the parser table of an LALR(1) implementation, and it can be handed to any backend to look at when
// such a comparison fails.
func LALR1FromLR1(parser backend.Parser) (backend.Parser, error) {
	merger := NewLALR1Builder(parser)
	if err := merger.Build(); err != nil {
		return backend.Parser{}, err
	}
	return merger.Parser(), nil
}

// LALR1Builder merges the isocores of a canonical LR(1) parser table into an LALR(1) parser table.
//
// Merging isocores collapses states, so the states of the LALR(1) parser table need indexes of their own. The builder
// hands them out in a first pass over the LR(1) states, and only fills the states in a second pass, so that every
// transition can name its destination by an index which is already known.
type LALR1Builder struct {
	// lr1Parser is the canonical LR(1) parser table to merge.
	lr1Parser backend.Parser

	// lalr1Parser is the LALR(1) parser table under construction.
	lalr1Parser backend.Parser

	// lalr1StateIdxByLR1StateIdx maps a state of the LR(1) parser table to the state of the LALR(1) parser table it
	// merges into. States which agree on their kernel items are isocores and share an entry.
	lalr1StateIdxByLR1StateIdx []int

	// lr1StateIdxByLALR1StateIdx maps a state of the LALR(1) parser table to the first state of the isocore it was
	// created for. That first state is the representative of its isocore.
	lr1StateIdxByLALR1StateIdx []int

	// lookaheadSetsByLALR1StateIdx holds the lookahead set of every production a state of the LALR(1) parser table
	// reduces. The reduce actions are only created once these lookahead sets are complete, because a ReduceActionSet is
	// ordered by production index and lookahead set, so growing a lookahead set which is already part of the set would
	// corrupt that order.
	lookaheadSetsByLALR1StateIdx []map[int]backend.LookaheadSet
}

// NewLALR1Builder returns a merger for the given canonical LR(1) parser table.
func NewLALR1Builder(lr1Parser backend.Parser) LALR1Builder {
	return LALR1Builder{
		lr1Parser: lr1Parser,
	}
}

// Build builds the LALR(1) parser table from the canonical LR(1) parser table.
func (m *LALR1Builder) Build() error {
	if err := m.initStateIdxs(); err != nil {
		return err
	}
	m.initStates()
	if err := m.mergeStates(); err != nil {
		return err
	}
	m.addReduceActions()
	return nil
}

// Parser returns the LALR(1) parser table. It is only complete after Build returned without an error.
func (m *LALR1Builder) Parser() backend.Parser {
	return m.lalr1Parser
}

// initStateIdxs gives every state of the LR(1) parser table the index of the state of the LALR(1) parser table it
// merges into, and remembers the representative of every isocore.
//
// Walking the LR(1) states in index order hands out the LALR(1) indexes in the order in which the isocores are first
// seen, which keeps the start state of the LR(1) parser table the start state of the LALR(1) parser table.
func (m *LALR1Builder) initStateIdxs() error {
	m.lalr1StateIdxByLR1StateIdx = make([]int, len(m.lr1Parser.States))
	m.lr1StateIdxByLALR1StateIdx = make([]int, 0, len(m.lr1Parser.States))
	lalr1StateIdxByKernelKey := make(map[string]int, len(m.lr1Parser.States))

	for lr1StateIdx := range m.lr1Parser.States {
		lr1State := &m.lr1Parser.States[lr1StateIdx]
		if lr1State.DefaultReduceProductionIdx != nil {
			// A default reduce action reduces on any lookahead, which means the lookahead set the merge would have to
			// take the union of is gone. Merging such a state cannot give the right answer, so we refuse to guess.
			return fmt.Errorf(
				"state %d has a default reduce action for production %d, which discards the lookahead set for merging",
				lr1StateIdx, *lr1State.DefaultReduceProductionIdx,
			)
		}

		// The core set is ordered, so the raw bytes of its values are a stable and collision-free identity.
		kernelKey := string(lr1State.KernelItems.Bytes())
		lalr1StateIdx, exists := lalr1StateIdxByKernelKey[kernelKey]
		if !exists {
			lalr1StateIdx = len(m.lr1StateIdxByLALR1StateIdx)
			lalr1StateIdxByKernelKey[kernelKey] = lalr1StateIdx
			m.lr1StateIdxByLALR1StateIdx = append(m.lr1StateIdxByLALR1StateIdx, lr1StateIdx)
		}
		m.lalr1StateIdxByLR1StateIdx[lr1StateIdx] = lalr1StateIdx
	}
	return nil
}

// initStates creates the states of the LALR(1) parser table, with nothing but their kernel items filled in.
func (m *LALR1Builder) initStates() {
	m.lalr1Parser = backend.Parser{
		Grammar: m.lr1Parser.Grammar,
		States:  make([]backend.State, len(m.lr1StateIdxByLALR1StateIdx)),
	}
	m.lookaheadSetsByLALR1StateIdx = make([]map[int]backend.LookaheadSet, len(m.lalr1Parser.States))

	for lalr1StateIdx, lr1StateIdx := range m.lr1StateIdxByLALR1StateIdx {
		// Every state of an isocore agrees on the kernel items by construction, so the representative speaks for all
		// of them.
		m.lalr1Parser.States[lalr1StateIdx].KernelItems = m.lr1Parser.States[lr1StateIdx].KernelItems
		m.lookaheadSetsByLALR1StateIdx[lalr1StateIdx] = make(map[int]backend.LookaheadSet)
	}
}

// mergeStates merges every state of the LR(1) parser table into the state of the LALR(1) parser table it belongs to.
func (m *LALR1Builder) mergeStates() error {
	for lr1StateIdx := range m.lr1Parser.States {
		if err := m.mergeTransitions(lr1StateIdx); err != nil {
			return err
		}
		m.mergeReduceActions(lr1StateIdx)
	}
	return nil
}

// mergeTransitions adds the transitions of the LR(1) state to the LALR(1) state it merges into, with every destination
// translated into the index of the state it merges into.
//
// Merging isocores never changes where a transition leads: the cores of a destination state follow from the cores of
// its source state alone. Every state merged into this one must therefore agree on the destination of each of its
// transitions, and a disagreement means the automaton is not what we take it to be.
func (m *LALR1Builder) mergeTransitions(lr1StateIdx int) error {
	lr1State := &m.lr1Parser.States[lr1StateIdx]
	lalr1State := &m.lalr1Parser.States[m.lalr1StateIdxByLR1StateIdx[lr1StateIdx]]

	for _, transitionAction := range lr1State.TransitionActions.All() {
		destinationIdx := m.lalr1StateIdxByLR1StateIdx[transitionAction.StateIdx()]

		existingIdx, exists := getDestinationStateIdx(lalr1State, transitionAction.SymbolRef())
		if exists && existingIdx != destinationIdx {
			return fmt.Errorf(
				"state %d transitions on %s to state %d, but a state with the same kernel items transitions to state %d",
				lr1StateIdx, transitionAction.SymbolRef(), destinationIdx, existingIdx,
			)
		}
		lalr1State.TransitionActions.Add(backend.NewTransitionAction(transitionAction.SymbolRef(), destinationIdx))
	}
	return nil
}

// mergeReduceActions collects the lookahead sets of the reduce actions of the LR(1) state into the lookahead sets of
// the LALR(1) state it merges into, keyed by production index. Taking the union of the lookahead sets of the same
// production is what turns the reduction lookaheads of the merged canonical LR(1) states into the reduction lookaheads
// of an LALR(1) state.
func (m *LALR1Builder) mergeReduceActions(lr1StateIdx int) {
	lr1State := &m.lr1Parser.States[lr1StateIdx]
	lookaheadSets := m.lookaheadSetsByLALR1StateIdx[m.lalr1StateIdxByLR1StateIdx[lr1StateIdx]]

	for _, reduceAction := range lr1State.ReduceActions.All() {
		// The zero value of a lookahead set holds no storage of its own, so the first merge copies instead of writing
		// into the lookahead set of the LR(1) parser table we were given.
		lookaheadSet := lookaheadSets[reduceAction.ProductionIdx]
		lookaheadSet.Merge(&reduceAction.LookaheadSet)
		lookaheadSets[reduceAction.ProductionIdx] = lookaheadSet
	}
}

// addReduceActions adds one reduce action per production to every state of the LALR(1) parser table, now that the
// lookahead set of each production is complete.
func (m *LALR1Builder) addReduceActions() {
	for lalr1StateIdx := range m.lalr1Parser.States {
		lalr1State := &m.lalr1Parser.States[lalr1StateIdx]
		lookaheadSets := m.lookaheadSetsByLALR1StateIdx[lalr1StateIdx]

		for _, productionIdx := range slices.Sorted(maps.Keys(lookaheadSets)) {
			lalr1State.ReduceActions.Add(backend.NewReduceAction(lookaheadSets[productionIdx], productionIdx))
		}
	}
}

// getDestinationStateIdx returns the destination state index of the transition of the state on the given symbol, and
// reports whether the state has such a transition.
func getDestinationStateIdx(state *backend.State, symbolRef frontend.SymbolRef) (int, bool) {
	for _, transitionAction := range state.TransitionActions.All() {
		if transitionAction.SymbolRef() == symbolRef {
			return transitionAction.StateIdx(), true
		}
	}
	return 0, false
}
