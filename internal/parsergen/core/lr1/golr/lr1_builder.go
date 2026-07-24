package golr

import (
	"context"
	"runtime/trace"
	"slices"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

// LR1Builder is implementing the algorithm for building canonical LR(1) parser tables.
//
// States are constructed from LR(1) items, which are cores extended with a lookahead set. Two states are only shared
// when their kernel items agree on both the cores and the lookahead sets.
//
// The builder does not resolve conflicts. A state with overlapping reduce actions, or with a reduce action whose
// lookahead set overlaps a terminal transition, is reported as it is.
type LR1Builder struct {
	// grammar is the augmented context free grammar for which canonical LR(1) parser tables should be created.
	grammar frontend.Grammar

	// maxStates is the number of states after which the construction gives up with backend.ErrStateLimitExceeded.
	maxStates int

	// productionIdxsByNonterminalIdx maps a nonterminal index to a slice of production indexes. This makes it easier to
	// find all productions which have the given nonterminal on the left hand side of the production.
	productionIdxsByNonterminalIdx [][]int

	// nullableByNonterminalIdx reports if the nonterminal can derive the empty string.
	nullableByNonterminalIdx []bool

	// firstByNonterminalIdx holds the terminals which can start a string derived from the nonterminal.
	firstByNonterminalIdx []backend.LookaheadSet

	// states is the list of states for the parser.
	states []backend.State

	// kernelItemsByStateIdx holds the kernel items of each state, indexed by state index. In contrast to the kernel
	// items of backend.State these carry the lookahead sets, which canonical LR(1) needs to tell states apart.
	kernelItemsByStateIdx []ItemSet

	// stateIdxsByKernelItemHash maps a kernel item hash to a list of state indexes. That way we can check if a set of
	// kernel items was already seen and what states could possibly match those kernel items.
	stateIdxsByKernelItemHash map[uint64][]int
}

// NewLR1Builder returns a new builder for canonical LR(1) parser tables. The grammar provided MUST be an augmented
// grammar. The builder gives up with backend.ErrStateLimitExceeded once the table grows beyond maxStates states.
func NewLR1Builder(grammar frontend.Grammar) LR1Builder {
	return LR1Builder{
		grammar:                   grammar,
		maxStates:                 backend.MaxAddressableStates(grammar),
		stateIdxsByKernelItemHash: make(map[uint64][]int, 128),
	}
}

// Build constructs canonical LR(1) parser tables. You can retrieve the generated parser with a call to Parser
// afterward.
func (b *LR1Builder) Build() error {
	defer trace.StartRegion(context.TODO(), "Build canonical LR(1) parser tables").End()

	b.initProductionIdxsByNonterminalIdx()
	b.initNullableAndFirstByNonterminalIdx()
	return b.buildStates()
}

// initProductionIdxsByNonterminalIdx initializes the helper variable productionIdxsByNonterminalIdx.
func (b *LR1Builder) initProductionIdxsByNonterminalIdx() {
	b.productionIdxsByNonterminalIdx = make([][]int, len(b.grammar.Nonterminals))
	for productionIdx, production := range b.grammar.Productions {
		b.productionIdxsByNonterminalIdx[production.NonterminalIdx] = append(
			b.productionIdxsByNonterminalIdx[production.NonterminalIdx],
			productionIdx,
		)
	}
}

// initNullableAndFirstByNonterminalIdx computes the nullable nonterminals and their first sets in a single fixed-point
// computation. A nonterminal is nullable when one of its productions has a right hand side which can vanish entirely.
// The first set of a nonterminal holds every terminal which can start a string derived from that nonterminal, which is
// collected by walking the right hand side of each of its productions until a symbol is reached which cannot vanish.
func (b *LR1Builder) initNullableAndFirstByNonterminalIdx() {
	b.nullableByNonterminalIdx = make([]bool, len(b.grammar.Nonterminals))
	b.firstByNonterminalIdx = make([]backend.LookaheadSet, len(b.grammar.Nonterminals))

	changed := true
	for changed {
		changed = false
		for _, production := range b.grammar.Productions {
			firstSet := &b.firstByNonterminalIdx[production.NonterminalIdx]

			firstSetChanged, nullable := b.firstOfSequence(production.SymbolRefs, firstSet)
			if firstSetChanged {
				changed = true
			}
			if nullable && !b.nullableByNonterminalIdx[production.NonterminalIdx] {
				// The whole right hand side of the production can vanish, which makes the nonterminal on the left hand
				// side nullable.
				b.nullableByNonterminalIdx[production.NonterminalIdx] = true
				changed = true
			}
		}
	}
}

// firstOfSequence merges the terminals which can start the sequence of symbols into the lookahead set. The first return
// value reports if that grew the lookahead set. The second return value reports if the whole sequence can vanish, in
// which case whatever follows the sequence can start it as well.
//
// While the nullable and first sets are still being computed, this works on the intermediate results. The enclosing
// fixed-point computation repeats until those results stop growing.
func (b *LR1Builder) firstOfSequence(
	symbolRefs []frontend.SymbolRef,
	lookaheadSet *backend.LookaheadSet,
) (bool, bool) {
	changed := false
	for _, symbolRef := range symbolRefs {
		if symbolRef.IsTerminal() {
			changed = lookaheadSet.Add(symbolRef.Idx()) || changed
			return changed, false
		}

		changed = lookaheadSet.Merge(&b.firstByNonterminalIdx[symbolRef.Idx()]) || changed
		if !b.nullableByNonterminalIdx[symbolRef.Idx()] {
			return changed, false
		}
	}
	return changed, true
}

// buildStates constructs the LR(1) states. This is a fixed-point computation which starts at the state for the start
// production and follows the transitions of every state until no new state shows up.
func (b *LR1Builder) buildStates() error {
	defer trace.StartRegion(context.TODO(), "Build LR(1) states").End()

	// We keep those variables outside the loop to re-use their allocated memory in subsequent loops.
	nextKernelItemsBySymbolRef := make(map[frontend.SymbolRef]*ItemSet, 32)
	sortedSymbolRefs := make([]frontend.SymbolRef, 0, 32)

	stateIdxWorkList := utils.NewDynamicRingBuffer[int]()
	b.initStartState(&stateIdxWorkList)
	for !stateIdxWorkList.IsEmpty() {
		stateIdx := stateIdxWorkList.Remove()

		closure := b.closure(&b.kernelItemsByStateIdx[stateIdx])
		b.recordReduceActions(stateIdx, &closure)
		b.buildNextKernelItems(&closure, nextKernelItemsBySymbolRef)

		sortedSymbolRefs = b.getSortedSymbolRefs(nextKernelItemsBySymbolRef, sortedSymbolRefs)
		for _, symbolRef := range sortedSymbolRefs {
			nextKernelItems := nextKernelItemsBySymbolRef[symbolRef]
			nextKernelItemsHash := nextKernelItems.Hash()

			destinationStateIdx, found := b.getStateIdxByKernelItems(nextKernelItems, nextKernelItemsHash)
			if !found {
				destinationStateIdx = b.addNewState(&stateIdxWorkList, nextKernelItems, nextKernelItemsHash)
			}
			b.states[stateIdx].TransitionActions.Add(backend.NewTransitionAction(symbolRef, destinationStateIdx))
		}

		// A single state can push us over the limit with more than one new state. We accept that overshoot and check
		// only once per state instead of on every new state, which keeps the check at the level where the states are
		// grown as a whole.
		if err := backend.CheckStateLimit("canonical LR(1)", len(b.states), b.maxStates); err != nil {
			return err
		}

		// clean up our temporary variables, so they are ready to go in the next loop
		clear(nextKernelItemsBySymbolRef)
	}
	return nil
}

// initStartState constructs the start state of the parser and sets up the work list with that state.
//
// The augmented grammar spells out the end of input as the last symbol of the start production, so the item for the
// start production needs no lookahead terminal of its own. Nothing can follow the end of input.
func (b *LR1Builder) initStartState(stateIdxWorkList *utils.DynamicRingBuffer[int]) {
	var startKernelItems ItemSet
	for _, productionIdx := range b.productionIdxsByNonterminalIdx[b.grammar.StartNonterminalIdx] {
		startKernelItems.Add(backend.NewCore(productionIdx, 0), &backend.LookaheadSet{})
	}

	b.addNewState(stateIdxWorkList, &startKernelItems, startKernelItems.Hash())
}

// addNewState adds a new state for the kernel items and adds the new state to the worklist for further processing in
// one of the next loops.
func (b *LR1Builder) addNewState(
	stateIdxWorkList *utils.DynamicRingBuffer[int],
	kernelItems *ItemSet,
	kernelItemsHash uint64,
) int {
	b.states = append(b.states, backend.State{
		KernelItems: kernelItems.CoreSet(),
	})
	b.kernelItemsByStateIdx = append(b.kernelItemsByStateIdx, *kernelItems)

	stateIdx := len(b.states) - 1

	// Record the new state with its kernel item hash, so we have an easier time when we want to know if the state
	// already exists.
	b.stateIdxsByKernelItemHash[kernelItemsHash] = append(b.stateIdxsByKernelItemHash[kernelItemsHash], stateIdx)
	stateIdxWorkList.Add(stateIdx)
	return stateIdx
}

// closure extends the kernel items with every item which the parser could be working on at the same time. Whenever an
// item expects a nonterminal at its current position, an item at the start of each production of that nonterminal is
// added. The lookahead set of such an added item consists of the terminals which can start the rest of the item behind
// the nonterminal, plus the lookahead set of the item itself when that rest can vanish.
func (b *LR1Builder) closure(kernelItems *ItemSet) ItemSet {
	var result ItemSet
	coreWorkList := utils.NewDynamicRingBuffer[backend.Core]()
	for _, item := range kernelItems.All() {
		result.Add(item.Core, &item.LookaheadSet)
		coreWorkList.Add(item.Core)
	}

	for !coreWorkList.IsEmpty() {
		core := coreWorkList.Remove()

		production := b.grammar.Productions[core.ProductionIdx()]
		if core.Position() == len(production.SymbolRefs) {
			// The item is at the end of the production and expects no further symbol.
			continue
		}
		symbolRef := production.SymbolRefs[core.Position()]
		if symbolRef.IsTerminal() {
			// The item expects a terminal, which contributes no items to the closure.
			continue
		}

		var generatedLookaheadSet backend.LookaheadSet
		if _, nullable := b.firstOfSequence(production.SymbolRefs[core.Position()+1:], &generatedLookaheadSet); nullable {
			// Everything behind the nonterminal can vanish, so whatever may follow the item may also follow the
			// nonterminal. NOTE: The pointer must not outlive the following calls to Add, which can move the items
			// around in memory.
			generatedLookaheadSet.Merge(result.LookaheadSetForCore(core))
		}

		for _, productionIdx := range b.productionIdxsByNonterminalIdx[symbolRef.Idx()] {
			generatedCore := backend.NewCore(productionIdx, 0)
			if result.Add(generatedCore, &generatedLookaheadSet) {
				// The item is new, or it grew a lookahead terminal which it needs to pass on to the items it generates
				// itself. Either way it has to be visited (again).
				coreWorkList.Add(generatedCore)
			}
		}
	}
	return result
}

// recordReduceActions adds a reduce action for every item of the closure which is at the end of its production. The
// lookahead set of the item is the lookahead set of the reduce action.
func (b *LR1Builder) recordReduceActions(stateIdx int, closure *ItemSet) {
	for _, item := range closure.All() {
		production := b.grammar.Productions[item.Core.ProductionIdx()]
		if item.Core.Position() != len(production.SymbolRefs) {
			continue
		}
		b.states[stateIdx].ReduceActions.Add(backend.NewReduceAction(item.LookaheadSet, item.Core.ProductionIdx()))
	}
}

// buildNextKernelItems constructs the kernel items which we can transition to from the current state, keyed by the
// symbol which is transitioned over. The lookahead sets are carried over unchanged, as moving the position within a
// production does not change what may follow that production.
func (b *LR1Builder) buildNextKernelItems(
	closure *ItemSet,
	nextKernelItemsBySymbolRef map[frontend.SymbolRef]*ItemSet,
) {
	for _, item := range closure.All() {
		production := b.grammar.Productions[item.Core.ProductionIdx()]
		if item.Core.Position() == len(production.SymbolRefs) {
			continue
		}

		symbolRef := production.SymbolRefs[item.Core.Position()]
		nextKernelItems, exists := nextKernelItemsBySymbolRef[symbolRef]
		if !exists {
			nextKernelItems = &ItemSet{}
			nextKernelItemsBySymbolRef[symbolRef] = nextKernelItems
		}
		nextKernelItems.Add(backend.NewCore(item.Core.ProductionIdx(), item.Core.Position()+1), &item.LookaheadSet)
	}
}

// getSortedSymbolRefs returns a sorted list of symbols to index into the nextKernelItemsBySymbolRef. This is important
// as we want to have a stable order in which states are created, and we would not have that stability with the map
// alone, as the map does not guarantee any specific order for its keys.
//
// The buffer is reused across states to avoid an allocation per state. It is reset before use, so the caller can pass
// the slice returned by the previous call back in.
func (b *LR1Builder) getSortedSymbolRefs(
	nextKernelItemsBySymbolRef map[frontend.SymbolRef]*ItemSet,
	buffer []frontend.SymbolRef,
) []frontend.SymbolRef {
	symbolRefs := buffer[:0]
	for symbolRef := range nextKernelItemsBySymbolRef {
		symbolRefs = append(symbolRefs, symbolRef)
	}
	slices.Sort(symbolRefs)
	return symbolRefs
}

// getStateIdxByKernelItems looks up our list of states and returns the state index which has the given kernel items.
// The second return value will report if the state was found or not.
func (b *LR1Builder) getStateIdxByKernelItems(kernelItems *ItemSet, kernelItemsHash uint64) (int, bool) {
	for _, stateIdx := range b.stateIdxsByKernelItemHash[kernelItemsHash] {
		if kernelItems.Equal(&b.kernelItemsByStateIdx[stateIdx]) {
			return stateIdx, true
		}
	}
	return 0, false
}

// Parser returns the canonical LR(1) parser table.
//
// The parser table is only valid when Build was called before and did not return an error.
func (b *LR1Builder) Parser() backend.Parser {
	return backend.Parser{
		Grammar: b.grammar,
		States:  b.states,
	}
}
