package ielr1go

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

	// nullableByNonterminalIdx provides information about a nonterminal index being nullable or not. This is needed
	// for calculating if the rest of some item can be empty or not.
	nullableByNonterminalIdx map[int]bool

	// states is the list of states for the parser.
	states []backend.State

	// stateIdxsByKernelItemHash maps a kernel item hash to a list of state indexes. That way we can check if a set of
	// kernel items was already seen and what states could possibly match those kernel items.
	stateIdxsByKernelItemHash map[uint64][]int

	// reduceActions is a list with all reduce actions for the parser.
	reduceActions []ReduceActionRecord

	// terminalTransitions is the flat list of all terminal transitions for the parser, appended state by state during
	// LR(0) construction.
	terminalTransitions []TransitionRecord

	// terminalTransitionsByState indexes into terminalTransitions to give the terminal transitions of a single state
	// when indexed by state index. This relies on the terminal transitions of a state being stored contiguously: while
	// a state is processed nothing else appends to terminalTransitions, so its transitions form one uninterrupted run
	// which the view captures as an offset and a length.
	terminalTransitionsByState []SliceView

	// gotoRecords provides details about each nonterminal transition. This is derived from definition 3.4 of IELR(1).
	gotoRecords []GotoRecord

	// gotoFollows holds the goto follow set for each goto, indexed by goto index. This is "goto_follows" from IELR(1)
	// definition 3.4.
	gotoFollows []backend.LookaheadSet

	// alwaysFollows holds the follow set from definition 3.20 of IELR(1), indexed by goto index.
	alwaysFollows []backend.LookaheadSet

	// gotoIdxsByStateIdx provides a list of goto indexes when indexed by state index. This is helpful when calculating
	// internal dependencies, as we need access to all gotos within the same state.
	gotoIdxsByStateIdx map[int][]int

	// backwardTransitionsByStateIdx provides information about which transitions lead into the state index.
	backwardTransitionsByStateIdx map[int]BackwardTransitionInfo

	// gotoFollowsSuccessorRelation is the digraph describing the successor dependencies as GFs(g, g') from IELR(1)
	// definition 3.5.
	gotoFollowsSuccessorRelation []Edge

	// gotoFollowsInternalRelation is the digraph describing the internal dependencies as GFi(g, g') from IELR(1)
	// definition 3.8. This relation is needed in later stages of IELR(1), therefore save it here separately.
	gotoFollowsInternalRelation []Edge

	// gotoFollowsPredecessorRelation is the digraph describing the predecessor dependencies as GFp(g, g') from IELR(1)
	// definition 3.9.
	gotoFollowsPredecessorRelation []Edge

	// successorDependencyCandidates is a list which holds the goto indexes of gotos which happen on nullable
	// nonterminals. Those gotos are the destinations for the goto follows successor relations. The relations can then
	// be built from this list.
	successorDependencyCandidates []int

	// internalDependencyCandidates provides a list of goto indexes which are part of an internal dependency. This list
	// is constructed during LR(0) state construction and used afterward to build the goto follows internal relations.
	internalDependencyCandidates []InternalDependencyCandidate

	// predecessorDependencyCandidates provides a list of goto indexes which are part of a predecessor dependency. This
	// list is constructed during LR(0) state construction and used afterward to build the goto follows predecessor
	// relations.
	predecessorDependencyCandidates []PredecessorDependencyCandidate
}

// NewLALR1Builder returns a new builder for LALR(1) parser tables. The grammar provided MUST be an augmented grammar.
func NewLALR1Builder(grammar frontend.Grammar) LALR1Builder {
	return LALR1Builder{
		grammar:                        grammar,
		productionIdxsByNonterminalIdx: make(map[int][]int, 128),
		nullableByNonterminalIdx:       make(map[int]bool, 128),

		stateIdxsByKernelItemHash: make(map[uint64][]int, 128),

		gotoIdxsByStateIdx: make(map[int][]int, 256),

		backwardTransitionsByStateIdx: make(map[int]BackwardTransitionInfo),
	}
}

// Build constructs LALR(1) parser tables. You can retrieve the generated parser with a call to Parser afterward.
func (b *LALR1Builder) Build() {
	defer trace.StartRegion(context.TODO(), "Build LALR(1) parser tables").End()

	b.initProductionIdxsByNonterminalIdx()
	b.initNullableByNonterminalIdx()
	b.buildLR0States()
	b.addReductionLookaheadSets()
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

// initNullableByNonterminalIdx initializes the helper variable nullableByNonterminalIdx. It is doing a fixed-point
// computation to find all the nullable nonterminals by inspecting the productions and checking for directly empty
// right hand sides of the productions or by indirectly empty right hand sides.
func (b *LALR1Builder) initNullableByNonterminalIdx() {
	changed := true
	for changed {
		changed = false
		for nonterminalIdx, productionIdxs := range b.productionIdxsByNonterminalIdx {
			if b.nullableByNonterminalIdx[nonterminalIdx] {
				// We already know that this nonterminal is nullable, so we do not need to check all productions
				// for that nonterminal again.
				continue
			}
			if slices.ContainsFunc(productionIdxs, b.isProductionNullable) {
				// As the right hand side of the production can be empty, we know that the nonterminal on the
				// left hand side of the production is nullable.
				b.nullableByNonterminalIdx[nonterminalIdx] = true
				changed = true
			}
		}
	}
}

// isProductionNullable reports if the right hand side of the production is empty or the right hand side consists
// only of nonterminals which are nullable themselves.
func (b *LALR1Builder) isProductionNullable(productionIdx int) bool {
	return b.isCoreTailEmpty(backend.NewCore(productionIdx, 0))
}

// isCoreTailEmpty reports if the position within the production is at the end of the production or the symbols for the
// following the current position are all nullable.
func (b *LALR1Builder) isCoreTailEmpty(core backend.Core) bool {
	production := b.grammar.Productions[core.ProductionIdx()]

	if core.Position() == len(production.SymbolRefs) {
		// The item is already at the end of the production. The tail is therefore empty.
		return true
	}

	for _, symbolRef := range production.SymbolRefs[core.Position():] {
		if symbolRef.IsTerminal() {
			// The symbol is a terminal which means the tail can not be empty.
			return false
		}
		if !b.nullableByNonterminalIdx[symbolRef.Idx()] {
			// The symbol is a nonterminal which is not nullable which means the tail can not be empty.
			return false
		}
	}
	// All remaining symbols were nonterminals and each nonterminal was nullable. Therefore, the core tail is empty.
	return true
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

		// The terminal transitions of this state are about to be appended contiguously to terminalTransitions, so we
		// remember where they start and record their length once the state is fully processed. See the doc comment on
		// terminalTransitionsByState for the contiguity invariant this relies on.
		// TODO: Consider replacing this offset/length bookkeeping with a simpler per-state transition map.
		b.terminalTransitionsByState[stateIdx].Offset = len(b.terminalTransitions)

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
			b.recordTransition(stateIdx, symbolRef, destinationStateIdx, nextKernelItems)
		}
		b.terminalTransitionsByState[stateIdx].Length = len(b.terminalTransitions) - b.terminalTransitionsByState[stateIdx].Offset

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

	// Resize these as well, so that they can be accessed with the state index.
	b.terminalTransitionsByState = append(b.terminalTransitionsByState, SliceView{})

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
// we need to supply the indexes for empty productions separately.
func (b *LALR1Builder) recordReduceActions(stateIdx int, emptyProductionIdxs []int) {
	for _, core := range b.states[stateIdx].KernelItems.All() {
		production := b.grammar.Productions[core.ProductionIdx()]
		if core.Position() == len(production.SymbolRefs) {
			b.reduceActions = append(b.reduceActions, ReduceActionRecord{
				StateIdx: stateIdx,
				Core:     core,
			})
		}
	}
	for _, productionIdx := range emptyProductionIdxs {
		b.reduceActions = append(b.reduceActions, ReduceActionRecord{
			StateIdx: stateIdx,
			Core:     backend.NewCore(productionIdx, 0),
		})
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

// recordTransition adds the transition to the state.
func (b *LALR1Builder) recordTransition(
	fromStateIdx int,
	symbolRef frontend.SymbolRef,
	toStateIdx int,
	nextKernelItems *backend.CoreSet,
) {
	if symbolRef.IsNonterminal() {
		b.recordNonterminalTransition(fromStateIdx, symbolRef.Idx(), toStateIdx, nextKernelItems)
	} else {
		b.recordTerminalTransition(fromStateIdx, symbolRef.Idx(), toStateIdx)
	}
}

// recordNonterminalTransition adds the nonterminal transition to the state.
func (b *LALR1Builder) recordNonterminalTransition(
	fromStateIdx int,
	nonterminalIdx int,
	toStateIdx int,
	nextKernelItems *backend.CoreSet,
) {
	// record forward transition
	b.gotoRecords = append(b.gotoRecords, GotoRecord{
		FromStateIdx:   fromStateIdx,
		ToStateIdx:     toStateIdx,
		NonterminalIdx: nonterminalIdx,
	})
	gotoIdx := len(b.gotoRecords) - 1
	b.gotoIdxsByStateIdx[fromStateIdx] = append(b.gotoIdxsByStateIdx[fromStateIdx], gotoIdx)

	// record backward transition
	transitions, exist := b.backwardTransitionsByStateIdx[toStateIdx]
	if !exist {
		transitions = NewBackwardTransitionInfo()
	}
	transitions.NonterminalTransitions[nonterminalIdx] = append(
		transitions.NonterminalTransitions[nonterminalIdx],
		fromStateIdx,
	)
	b.backwardTransitionsByStateIdx[toStateIdx] = transitions

	// record information needed to calculate goto follows later
	b.recordSuccessorDependencyCandidate(nonterminalIdx, gotoIdx)
	b.recordIncludesDependencyCandidate(nextKernelItems, gotoIdx)
}

// recordSuccessorDependencyCandidate checks if the goto is part of a goto follows successor relation as specified by
// definition 3.5 of IELR(1) and records it as candidate for later use.
func (b *LALR1Builder) recordSuccessorDependencyCandidate(nonterminalIdx int, gotoIdx int) {
	// Check if this goto is part of a successor dependency for the goto follows.
	if b.nullableByNonterminalIdx[nonterminalIdx] {
		// We need this information when constructing the goto follows successor relation.
		b.successorDependencyCandidates = append(b.successorDependencyCandidates, gotoIdx)
	}
}

// recordIncludesDependencyCandidate checks if the kernel items are part of a goto follows includes relation as
// specified by definition 3.7 of IELR(1). As goto follows includes relations can be broken down into goto follows
// internal relations as specified by definition 3.8 of IELR(1) and goto follows predecessor relations as specified by
// definition 3.9 of IELR(1), this method checks for both cases and records the goto as a candidate for later use.
func (b *LALR1Builder) recordIncludesDependencyCandidate(nextKernelItems *backend.CoreSet, gotoIdx int) {
	for _, kernelItem := range nextKernelItems.All() {
		if !b.isCoreTailEmpty(kernelItem) {
			// We are not interested in items which are not empty for the rest of the production.
			continue
		}

		if kernelItem.Position() == 1 {
			// We found an item which is at the start of the production and empty after the current position. This
			// is a candidate for an internal dependency which we need to record for later. Note that the item here
			// was already advanced by one symbol, so we need to check for 1 instead of 0.
			production := b.grammar.Productions[kernelItem.ProductionIdx()]
			b.internalDependencyCandidates = append(b.internalDependencyCandidates, InternalDependencyCandidate{
				GotoIdx:        gotoIdx,
				NonterminalIdx: production.NonterminalIdx,
			})
		} else {
			// We found an item which is not at the start of the production and empty after the current position.
			// This is a candidate for a predecessor dependency which we need to record for later.
			b.predecessorDependencyCandidates = append(b.predecessorDependencyCandidates, PredecessorDependencyCandidate{
				GotoIdx: gotoIdx,
				// Note that we need to move back the kernel by one position, because the kernel we have here is already
				// moved forward by one, but we need the core as it was for the state we are transitioning from.
				Core: backend.NewCore(kernelItem.ProductionIdx(), kernelItem.Position()-1),
			})
		}
	}
}

// recordTerminalTransition adds the terminal transition to the state.
func (b *LALR1Builder) recordTerminalTransition(fromStateIdx int, terminalIdx int, toStateIdx int) {
	// Record the forward transition.
	b.terminalTransitions = append(b.terminalTransitions, TransitionRecord{
		FromStateIdx: fromStateIdx,
		SymbolIdx:    terminalIdx,
		ToStateIdx:   toStateIdx,
	})

	// Record backward transition.
	transitions, exist := b.backwardTransitionsByStateIdx[toStateIdx]
	if !exist {
		transitions = NewBackwardTransitionInfo()
	}
	transitions.TerminalTransitions[terminalIdx] = append(transitions.TerminalTransitions[terminalIdx], fromStateIdx)
	b.backwardTransitionsByStateIdx[toStateIdx] = transitions
}

// addReductionLookaheadSets extends the reduce actions with reduction lookahead sets derived from goto follows.
func (b *LALR1Builder) addReductionLookaheadSets() {
	defer trace.StartRegion(context.TODO(), "Add reduction lookahead sets").End()

	b.buildGotoFollowsSuccessorRelations()
	b.buildGotoFollowsInternalRelations()
	b.buildGotoFollowsPredecessorRelations()

	// We follow implementation 2 from IELR(1) section 3.3.5, which computes goto follows from always follows
	// (definition 3.24) and never computes successor follows (definition 3.6). Only the successor relation itself is
	// needed, as an input to the always follows.
	//
	// TODO: Check if we can improve performance by not calculating all always and goto follows up front, but instead
	// lazily calculate those which we need for reduce actions. This could result in a significant amount of follow
	// sets not being calculated as they are not involved in any reduce action.
	b.calculateAlwaysFollows()
	b.calculateGotoFollows()

	// Calculate lookahead sets for reduce actions. We do this by tracing the core of the reduction back to the gotos
	// which initially generated the core. The goto follows of those gotos are then responsible for the reduction
	// lookahead set.
	for i := range b.reduceActions {
		for _, gotoIdx := range b.getGeneratedGotoIdxs(b.reduceActions[i].StateIdx, b.reduceActions[i].Core) {
			b.reduceActions[i].LookaheadSet.Merge(&b.gotoFollows[gotoIdx])
		}
	}
}

// buildGotoFollowsSuccessorRelations is building up the digraph for the goto follows successor relation as specified
// in IELR(1) definition 3.5. We are taking all the gotos we found to happen on nullable nonterminals during LR(0)
// state construction, and we are creating edges to those gotos from the gotos which are pointing to the same state.
func (b *LALR1Builder) buildGotoFollowsSuccessorRelations() {
	// Index all gotos by their target state, so we can find the gotos entering a state without scanning every goto for
	// each candidate. This turns the relation construction from quadratic into linear in the number of gotos plus the
	// number of produced edges.
	gotoIdxsByToStateIdx := make([][]int, len(b.states))
	for gotoIdx := range b.gotoRecords {
		toStateIdx := b.gotoRecords[gotoIdx].ToStateIdx
		gotoIdxsByToStateIdx[toStateIdx] = append(gotoIdxsByToStateIdx[toStateIdx], gotoIdx)
	}

	for _, nullableGotoIdx := range b.successorDependencyCandidates {
		// A goto g contributes an edge to the nullable goto g' exactly when to_state[g] = from_state[g'], so we only
		// need the gotos which end in the state the nullable goto starts from.
		fromStateIdx := b.gotoRecords[nullableGotoIdx].FromStateIdx
		for _, gotoIdx := range gotoIdxsByToStateIdx[fromStateIdx] {
			b.gotoFollowsSuccessorRelation = append(b.gotoFollowsSuccessorRelation, Edge{
				FromIdx: gotoIdx,
				ToIdx:   nullableGotoIdx,
			})
		}
	}
}

// buildGotoFollowsInternalRelations builds up the digraph for the goto follows internal relation as specified in
// definition 3.8 of IELR(1).
func (b *LALR1Builder) buildGotoFollowsInternalRelations() {
	for _, candidate := range b.internalDependencyCandidates {
		// We are looking for gotos within the same state which are done on the nonterminal which is on the left hand
		// side of the item of the candidate.
		stateIdx := b.gotoRecords[candidate.GotoIdx].FromStateIdx
		gotoIdxs := b.gotoIdxsByStateIdx[stateIdx]
		for _, gotoIdx := range gotoIdxs {
			if b.gotoRecords[gotoIdx].NonterminalIdx != candidate.NonterminalIdx {
				// This can not be an internal dependency, as the goto is happening on a different symbol than the
				// candidate.
				continue
			}
			// The internal relations are needed separately in later IELR(1) phases, therefore we note them down
			// separately.
			b.gotoFollowsInternalRelation = append(b.gotoFollowsInternalRelation, Edge{
				FromIdx: candidate.GotoIdx,
				ToIdx:   gotoIdx,
			})
		}
	}
}

// buildGotoFollowsPredecessorRelations builds up the digraph for the goto follows predecessor relation as specified in
// definition 3.9 of IELR(1). This is done by moving backwards through the states to find the goto which generated the
// core of our candidate goto.
func (b *LALR1Builder) buildGotoFollowsPredecessorRelations() {
	for _, candidate := range b.predecessorDependencyCandidates {
		gotoIdxs := b.getGeneratedGotoIdxs(b.gotoRecords[candidate.GotoIdx].FromStateIdx, candidate.Core)
		for _, gotoIdx := range gotoIdxs {
			b.gotoFollowsPredecessorRelation = append(b.gotoFollowsPredecessorRelation, Edge{
				FromIdx: candidate.GotoIdx,
				ToIdx:   gotoIdx,
			})
		}
	}
}

// calculateAlwaysFollows fills the always follows as specified in definition 3.20 of IELR(1).
func (b *LALR1Builder) calculateAlwaysFollows() {
	b.alwaysFollows = make([]backend.LookaheadSet, len(b.gotoRecords))
	// Initialize the always follows with the terminal transitions of the target state.
	for i := range b.gotoRecords {
		stateIdx := b.gotoRecords[i].ToStateIdx
		for _, transition := range SliceFromView(b.terminalTransitions, b.terminalTransitionsByState[stateIdx]) {
			b.alwaysFollows[i].Add(transition.SymbolIdx)
		}
	}

	gotoFollowsAlwaysRelation := make([]Edge, len(b.gotoFollowsSuccessorRelation)+len(b.gotoFollowsInternalRelation))
	copy(gotoFollowsAlwaysRelation[:len(b.gotoFollowsSuccessorRelation)], b.gotoFollowsSuccessorRelation)
	copy(gotoFollowsAlwaysRelation[len(b.gotoFollowsSuccessorRelation):], b.gotoFollowsInternalRelation)
	propagation := NewDigraphAlgorithm(b.alwaysFollows, gotoFollowsAlwaysRelation)
	propagation.Execute()
}

// calculateGotoFollows fills the goto follows as specified in definition 3.24 of IELR(1) by propagating the follow sets
// along the goto follows includes relations.
func (b *LALR1Builder) calculateGotoFollows() {
	b.gotoFollows = make([]backend.LookaheadSet, len(b.gotoRecords))
	// Initialize the goto follows with the always follows of the same goto.
	for i := range b.gotoRecords {
		b.gotoFollows[i].Merge(&b.alwaysFollows[i])
	}
	gotoFollowsIncludesRelation := make([]Edge, len(b.gotoFollowsInternalRelation)+len(b.gotoFollowsPredecessorRelation))
	copy(gotoFollowsIncludesRelation[:len(b.gotoFollowsInternalRelation)], b.gotoFollowsInternalRelation)
	copy(gotoFollowsIncludesRelation[len(b.gotoFollowsInternalRelation):], b.gotoFollowsPredecessorRelation)
	propagation := NewDigraphAlgorithm(b.gotoFollows, gotoFollowsIncludesRelation)
	propagation.Execute()
}

// getGeneratedGotoIdxs returns a list of goto indexes which generated the given core. This is done by tracing the core
// backwards through the states until we are at the start of the production and the left hand side appears in a goto.
func (b *LALR1Builder) getGeneratedGotoIdxs(stateIdx int, core backend.Core) []int {
	predecessorStateIdxs := []int{
		stateIdx,
	}

	// We need to move back through the states until we are at the start of the item.
	for position := core.Position(); position > 0; position-- {
		predecessorStateIdxs = b.followCoreBackward(backend.NewCore(core.ProductionIdx(), position), predecessorStateIdxs)
	}

	// Now let's look for the goto which has the left hand side of the production as a nonterminal transition.
	production := b.grammar.Productions[core.ProductionIdx()]
	var result []int
	for _, predecessorStateIdx := range predecessorStateIdxs {
		gotoIdxs := b.gotoIdxsByStateIdx[predecessorStateIdx]
		for _, gotoIdx := range gotoIdxs {
			if b.gotoRecords[gotoIdx].NonterminalIdx != production.NonterminalIdx {
				// We are looking for gotos on nonterminals which equal the left hand side of our production. This
				// is not one.
				continue
			}
			result = append(result, gotoIdx)
		}
	}
	return result
}

// followCoreBackward is moving the core one step back through the states and returns the list of state indexes the core
// was coming from.
func (b *LALR1Builder) followCoreBackward(core backend.Core, stateIdxs []int) []int {
	production := b.grammar.Productions[core.ProductionIdx()]
	symbolRef := production.SymbolRefs[core.Position()-1]
	var predecessorStateIdxs []int
	for _, stateIdx := range stateIdxs {
		// TODO: It is not necessary to check for specific symbols. As we only have kernel item cores in each state,
		// every predecessor must have the core in their state and every transition must be part of the core.
		if symbolRef.IsNonterminal() {
			predecessorStateIdxs = append(
				predecessorStateIdxs,
				b.backwardTransitionsByStateIdx[stateIdx].NonterminalTransitions[symbolRef.Idx()]...,
			)
		} else {
			predecessorStateIdxs = append(
				predecessorStateIdxs,
				b.backwardTransitionsByStateIdx[stateIdx].TerminalTransitions[symbolRef.Idx()]...,
			)
		}
	}
	return predecessorStateIdxs
}

// Parser returns the LALR(1) parser table.
//
// The parser table is only valid when Build was called before.
func (b *LALR1Builder) Parser() backend.Parser {
	var result backend.Parser

	result.Grammar = b.grammar
	result.States = b.states

	for _, reduceAction := range b.reduceActions {
		result.States[reduceAction.StateIdx].ReduceActions.Add(
			backend.NewReduceAction(reduceAction.LookaheadSet, reduceAction.Core.ProductionIdx()),
		)
	}

	for _, transitionAction := range b.terminalTransitions {
		result.States[transitionAction.FromStateIdx].TransitionActions.Add(
			backend.NewTransitionAction(
				frontend.NewTerminalRef(transitionAction.SymbolIdx),
				transitionAction.ToStateIdx,
			),
		)
	}
	for stateIdx, gotoIdxs := range b.gotoIdxsByStateIdx {
		for _, gotoIdx := range gotoIdxs {
			result.States[stateIdx].TransitionActions.Add(
				backend.NewTransitionAction(
					frontend.NewNonterminalRef(b.gotoRecords[gotoIdx].NonterminalIdx),
					b.gotoRecords[gotoIdx].ToStateIdx,
				),
			)
		}
	}
	return result
}
