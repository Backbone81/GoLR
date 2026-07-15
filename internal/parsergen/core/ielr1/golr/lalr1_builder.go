package golr

import (
	"context"
	"errors"
	"runtime/trace"
	"slices"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

// LALR1Builder is implementing the algorithm for building LALR(1) parser tables.
//
// At first LR(0) parser tables are created. Afterward the algorithm as described by DeRemer and Pennello in
// "Efficient Computation of LALR(1) Look-Ahead Sets" at https://doi.org/10.1145/69622.357187 is applied to compute
// reduction lookahead sets from goto follow sets.
//
// There is no dedicated LR(0) builder as the LR(0) construction can already provide vital information needed for
// LALR(1). This reduces the overhead of calculating information needed for LALR(1) which is already available during
// LR(0) construction.
//
// Building LALR(1) parser tables is phase 0 of the IELR(1) algorithm.
type LALR1Builder struct {
	// grammar is the augmented context free grammar for which LALR(1) parser tables should be created.
	grammar frontend.Grammar

	// productionIdxsByNonterminalIdx maps a nonterminal index to a slice of production indexes. This makes it easier to
	// find all productions which have the given nonterminal on the left hand side of the production.
	productionIdxsByNonterminalIdx map[int][]int

	// states is the list of states for the parser. During LR(0) construction each state is filled with its kernel items,
	// its transition actions and the productions it reduces, so the states describe the automaton on their own. The
	// reduction lookahead builder consumes them and their reduce actions receive their lookahead sets afterward.
	states []backend.State

	// stateIdxsByKernelItemHash maps a kernel item hash to a list of state indexes. That way we can check if a set of
	// kernel items was already seen and what states could possibly match those kernel items.
	stateIdxsByKernelItemHash map[uint64][]int

	// lookaheads computes the reduction lookahead sets from the LR(0) automaton. It also exposes the goto records, goto
	// follows and always follows the later IELR(1) phases need. Building LALR(1) parser tables is phase 0 of IELR(1),
	// and this is its second step (section 3.2.2).
	lookaheads ReductionLookaheadBuilder
}

// NewLALR1Builder returns a new builder for LALR(1) parser tables. The grammar provided MUST be an augmented grammar.
func NewLALR1Builder(grammar frontend.Grammar) LALR1Builder {
	return LALR1Builder{
		grammar:                        grammar,
		productionIdxsByNonterminalIdx: make(map[int][]int, 128),

		stateIdxsByKernelItemHash: make(map[uint64][]int, 128),
	}
}

// Build constructs LALR(1) parser tables. You can retrieve the generated parser with a call to Parser afterward.
func (b *LALR1Builder) Build() {
	defer trace.StartRegion(context.TODO(), "Build LALR(1) parser tables").End()

	b.initProductionIdxsByNonterminalIdx()
	b.buildLR0States()

	// The LR(0) states now describe the automaton on their own, so the reduction lookahead builder can derive the goto
	// follows from them and compute the reduction lookahead sets.
	b.lookaheads = NewReductionLookaheadBuilder(b.grammar, b.states)
	b.lookaheads.Build()
	b.applyReductionLookaheads()
}

// applyReductionLookaheads writes the reduction lookahead sets computed by the reduction lookahead builder back into the
// reduce actions of the states. The LR(0) construction added the reduce actions with an empty lookahead set, so we
// replace them with the ones carrying their lookahead set. A reduce action is keyed by its production and its lookahead
// set, so we cannot amend a lookahead set in place and rebuild the reduce actions of each state instead. Clearing keeps
// the backing storage the LR(0) construction already allocated, so refilling it with the same productions reuses that
// storage instead of allocating a new set per state.
func (b *LALR1Builder) applyReductionLookaheads() {
	for stateIdx := range b.states {
		b.states[stateIdx].ReduceActions.Clear()
	}
	for _, reduceAction := range b.lookaheads.ReduceActions() {
		b.states[reduceAction.StateIdx].ReduceActions.Add(
			backend.NewReduceAction(reduceAction.LookaheadSet, reduceAction.Core.ProductionIdx()),
		)
	}
}

// initProductionIdxsByNonterminalIdx initializes the helper variable productionIdxsByNonterminalIdx.
func (b *LALR1Builder) initProductionIdxsByNonterminalIdx() {
	for idx, production := range b.grammar.Productions {
		b.productionIdxsByNonterminalIdx[production.NonterminalIdx] = append(
			b.productionIdxsByNonterminalIdx[production.NonterminalIdx],
			idx,
		)
	}
}

// buildLR0States constructs the LR(0) states and records information needed for LALR(1) construction later on. This
// method is doing a fixed-point computation to find all possible states by calculating the closure for the cores of
// each state and advancing the position of each core by one.
func (b *LALR1Builder) buildLR0States() {
	defer trace.StartRegion(context.TODO(), "Build LR(0) parser tables").End()

	// We keep those variables outside the loop to re-use their allocated memory in subsequent loops.
	nextKernelItemsBySymbolRef := make(map[frontend.SymbolRef]*backend.CoreSet, 32)
	emptyProductionIdxs := make([]int, 0, 32)
	sortedSymbolRefs := make([]frontend.SymbolRef, 0, 32)

	stateIdxWorkList := utils.NewDynamicRingBuffer[int]()
	b.initStartState(&stateIdxWorkList)
	for !stateIdxWorkList.IsEmpty() {
		stateIdx := stateIdxWorkList.Remove()

		b.buildNextKernelItems(stateIdx, nextKernelItemsBySymbolRef, &emptyProductionIdxs)
		b.recordReduceActions(stateIdx, emptyProductionIdxs)
		sortedSymbolRefs = b.getSortedSymbolRefs(nextKernelItemsBySymbolRef, sortedSymbolRefs)
		for _, symbolRef := range sortedSymbolRefs {
			nextKernelItems := nextKernelItemsBySymbolRef[symbolRef]
			nextKernelItemsHash := nextKernelItems.Hash()

			destinationStateIdx, found := b.getStateIdxByKernelItems(nextKernelItems, nextKernelItemsHash)
			if !found {
				destinationStateIdx = b.addNewState(&stateIdxWorkList, nextKernelItems, nextKernelItemsHash)
			}
			b.recordTransition(stateIdx, symbolRef, destinationStateIdx)
		}

		// clean up our temporary variables, so they are ready to go in the next loop
		clear(nextKernelItemsBySymbolRef)
		emptyProductionIdxs = emptyProductionIdxs[:0]
	}
}

// initStartState constructs the start state of the parser and sets up the work list with that state.
func (b *LALR1Builder) initStartState(stateIdxWorkList *utils.DynamicRingBuffer[int]) {
	var startCores backend.CoreSet
	for _, productionIdx := range b.productionIdxsByNonterminalIdx[b.grammar.StartNonterminalIdx] {
		startCores.Add(backend.NewCore(productionIdx, 0))
	}
	utils.DebugAssert(func() error {
		if startCores.Length() != 1 {
			return errors.New("augmented grammars are expected to have exactly one kernel item for the start state")
		}
		return nil
	})

	b.addNewState(stateIdxWorkList, &startCores, startCores.Hash())
}

// addNewState adds a new state for the kernel items and adds the new state to the worklist for further processing in
// one of the next loops.
func (b *LALR1Builder) addNewState(
	stateIdxWorkList *utils.DynamicRingBuffer[int],
	kernelItems *backend.CoreSet,
	kernelItemsHash uint64,
) int {
	newState := backend.State{
		KernelItems: *kernelItems,
	}
	b.states = append(b.states, newState)

	stateIdx := len(b.states) - 1

	// Record the new state with its kernel item hash, so we have an easier time when we want to know if the state
	// already exists.
	b.stateIdxsByKernelItemHash[kernelItemsHash] = append(b.stateIdxsByKernelItemHash[kernelItemsHash], stateIdx)
	stateIdxWorkList.Add(stateIdx)
	return stateIdx
}

// buildNextKernelItems constructs the kernel items which we can transition to from the current state. The next kernel
// items are provided by symbol index. At the same time a list of productions with an empty right hand side is built,
// as that information is needed for the correct reduce actions and can be acquired here for free.
func (b *LALR1Builder) buildNextKernelItems(
	stateIdx int,
	nextKernelItemsBySymbolRef map[frontend.SymbolRef]*backend.CoreSet,
	emptyProductionIdxs *[]int,
) {
	for _, core := range b.states[stateIdx].KernelItems.All() {
		b.advanceCore(core, nextKernelItemsBySymbolRef, emptyProductionIdxs)
	}
}

// advanceCore is moving the given core one position forward. It is recursively processing cores for nonterminals which
// the core is moving over.
// The cores are collected into a core set for the symbol index which the position moved over. That way we get the
// transitions between cores and can calculate the transition actions.
// As productions which are empty on the right hand side will not show up in any core, we output these through the
// emptyProductionIdxs parameter. This can then be used for calculating reduce actions.
func (b *LALR1Builder) advanceCore(
	core backend.Core,
	nextKernelItemsBySymbolRef map[frontend.SymbolRef]*backend.CoreSet,
	emptyProductionIdxs *[]int,
) {
	production := b.grammar.Productions[core.ProductionIdx()]
	if core.Position() == len(production.SymbolRefs) {
		// We are already at the end of the production.
		if len(production.SymbolRefs) == 0 && !slices.Contains(*emptyProductionIdxs, core.ProductionIdx()) {
			// We found an empty production and need to record it for the main LR(0) loop. The same empty production can
			// be reached through several closure items which step over the same nullable nonterminal. As empty
			// productions never become part of a core, they bypass the core deduplication below, so we guard against
			// recording them more than once here. Without this guard we would emit duplicate reduce actions for the
			// empty production. The number of distinct empty productions per state is tiny, so the linear scan is
			// cheap and keeps this allocation-neutral.
			*emptyProductionIdxs = append(*emptyProductionIdxs, core.ProductionIdx())
		}
		return
	}

	symbolRef := production.SymbolRefs[core.Position()]
	advancedCore := backend.NewCore(core.ProductionIdx(), core.Position()+1)
	nextKernelItems, exists := nextKernelItemsBySymbolRef[symbolRef]
	if !exists {
		nextKernelItems = &backend.CoreSet{}
		nextKernelItemsBySymbolRef[symbolRef] = nextKernelItems
	}
	if !nextKernelItems.Add(advancedCore) {
		// We already have that core in our list, we can return early.
		return
	}

	if symbolRef.IsNonterminal() {
		// The symbol is a nonterminal, and we need to consider all productions which have that nonterminal on the left
		// hand side of the production.
		productionIdxs := b.productionIdxsByNonterminalIdx[symbolRef.Idx()]
		for _, productionIdx := range productionIdxs {
			b.advanceCore(backend.NewCore(productionIdx, 0), nextKernelItemsBySymbolRef, emptyProductionIdxs)
		}
	}
}

// recordReduceActions adds the reduce actions to the parser state. It looks for cores in the current state with a
// position at the end of the production. As the kernel items do not contain cores which have the position at the start,
// we need to supply the indexes for empty productions separately. The reduce actions are added with an empty lookahead
// set, which the reduction lookahead builder fills in later.
func (b *LALR1Builder) recordReduceActions(stateIdx int, emptyProductionIdxs []int) {
	for _, core := range b.states[stateIdx].KernelItems.All() {
		production := b.grammar.Productions[core.ProductionIdx()]
		if core.Position() == len(production.SymbolRefs) {
			b.states[stateIdx].ReduceActions.Add(backend.NewReduceAction(backend.LookaheadSet{}, core.ProductionIdx()))
		}
	}
	for _, productionIdx := range emptyProductionIdxs {
		b.states[stateIdx].ReduceActions.Add(backend.NewReduceAction(backend.LookaheadSet{}, productionIdx))
	}
}

// getSortedSymbolRefs returns a sorted list of symbols to index into the nextKernelItemsBySymbolRef. This is important
// as we want to have a stable order in which states are created, and we would not have that stability with the map
// alone, as the map does not guarantee any specific order for its keys.
//
// The buffer is reused across states to avoid an allocation per state. It is reset before use, so the caller can pass
// the slice returned by the previous call back in.
func (b *LALR1Builder) getSortedSymbolRefs(
	nextKernelItemsBySymbolRef map[frontend.SymbolRef]*backend.CoreSet,
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
func (b *LALR1Builder) getStateIdxByKernelItems(kernelItems *backend.CoreSet, kernelItemsHash uint64) (int, bool) {
	if stateIdxs, found := b.stateIdxsByKernelItemHash[kernelItemsHash]; found {
		for _, stateIdx := range stateIdxs {
			if kernelItems.Equal(&b.states[stateIdx].KernelItems) {
				return stateIdx, true
			}
		}
	}
	return 0, false
}

// recordTransition adds the transition to the state. The transition action makes the state describe its share of the
// automaton on its own, which is what the reduction lookahead builder later reads to re-derive the goto records and the
// backward transitions.
func (b *LALR1Builder) recordTransition(fromStateIdx int, symbolRef frontend.SymbolRef, toStateIdx int) {
	b.states[fromStateIdx].TransitionActions.Add(backend.NewTransitionAction(symbolRef, toStateIdx))
}

// Parser returns the LALR(1) parser table.
//
// The parser table is only valid when Build was called before.
func (b *LALR1Builder) Parser() backend.Parser {
	return backend.Parser{
		Grammar: b.grammar,
		States:  b.states,
	}
}
