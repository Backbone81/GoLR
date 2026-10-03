package counterexample

import (
	"cmp"
	"context"
	"fmt"
	"math"
	"runtime/trace"
	"slices"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

// LookupTables answers the queries of the counterexample search which parser tables do not answer directly. This is
// section 6, "Data structures", of the paper: the items of every state including the closure, the transitions and
// production steps between them in both directions, and the lookahead set of every item. The tables are built once and
// shared by the searches of all conflicts. The lookahead-sensitive graph of section 4 is not built here, the search
// discovers its vertices and edges.
//
// An item of a state, which is a vertex (s, itm) of the lookahead-sensitive graph without its lookahead set, is named
// by a single integer, its item index. The closure items are numbered too, although backend.State holds only the kernel
// items: they belong to the same state and have transitions, reductions and lookahead sets of their own. Kernel and
// closure items share one numbering, so the tables indexed by item and the edges between items need not tell them
// apart.
//
// The items of a state are adjacent: its kernel items sorted by core, followed by its closure items sorted by
// nonterminal and production index, so the closure items of a nonterminal are adjacent too. See kernelItems and
// closureItems for the ranges.
//
// The action filter and the reach of a terminal are computed on demand and kept, so the tables are not safe for
// concurrent use.
type LookupTables struct {
	grammar frontend.Grammar
	states  []backend.State

	// policy holds the declarations of the grammar. The action filter only needs what they leave competing, which is
	// the same under every policy, as the rules of last resort only decide what the declarations left.
	policy conflict.Policy

	firstSets                      frontend.FirstSets
	productionIdxsByNonterminalIdx [][]int

	// itemIdxOffsetByStateIdx holds the first item of every state, followed by the number of items, so the items of
	// state s are [itemIdxOffsetByStateIdx[s], itemIdxOffsetByStateIdx[s+1]).
	itemIdxOffsetByStateIdx []int

	// closureItemIdxOffsetByStateIdx holds the first closure item of every state, which ends its kernel items.
	closureItemIdxOffsetByStateIdx []int

	stateIdxByItemIdx []int
	coreByItemIdx     []backend.Core

	// transitionByItemIdx holds the item which the transition on the symbol after the dot leads to, or noItemIdx.
	transitionByItemIdx []int

	// reverseTransitions holds the items whose transition leads to item i at
	// [reverseTransitionOffsetByItemIdx[i], reverseTransitionOffsetByItemIdx[i+1]).
	reverseTransitionOffsetByItemIdx []int
	reverseTransitions               []int

	// itemIdxsByNextSymbol holds the items of every state ordered by the symbol after the dot, with the reduce items
	// last. Every state keeps the offsets of its items, so itemIdxOffsetByStateIdx slices it too.
	itemIdxsByNextSymbol []int

	// The lookahead sets, see initLookaheads. Lookahead set j is shared by the items
	// [itemIdxOffsetByLookaheadIdx[j], itemIdxOffsetByLookaheadIdx[j+1]).
	lookaheadIdxByItemIdx       []int
	itemIdxOffsetByLookaheadIdx []int
	lookaheads                  []backend.LookaheadSet

	// allowedContributionsByStateTerminal holds the action filter, see IsActionAllowed.
	allowedContributionsByStateTerminal map[stateTerminal]conflict.ContributionSet

	// emptyDerivationStepByNonterminalIdx holds the derivation steps, see EmptyDerivation.
	emptyDerivationStepByNonterminalIdx []derivationStep

	// terminalReachByTerminalIdx holds the reach of every terminal used so far, see terminalReach.
	terminalReachByTerminalIdx map[int]*terminalReach
}

// noItemIdx stands for a missing item, like the target of an item without a transition.
const noItemIdx = -1

// noNextSymbolKey orders the items without a symbol after the dot behind all others, see nextSymbolKey.
const noNextSymbolKey = math.MaxInt

// NewLookupTables builds the lookup tables of the parser tables a core returned, with their conflicts resolved or left
// unresolved.
func NewLookupTables(parser backend.Parser) LookupTables {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Counterexample: NewLookupTables").End()

	result := LookupTables{
		grammar:                             parser.Grammar,
		states:                              parser.States,
		policy:                              conflict.PrecedencePolicy(parser.Grammar),
		firstSets:                           frontend.NewFirstSets(parser.Grammar),
		allowedContributionsByStateTerminal: make(map[stateTerminal]conflict.ContributionSet),
		terminalReachByTerminalIdx:          make(map[int]*terminalReach),
	}
	result.initProductionIdxsByNonterminalIdx()
	result.initItems()
	result.initTransitions()
	result.initItemIdxsByNextSymbol()
	result.initLookaheads()
	result.initEmptyDerivationSteps()
	return result
}

// ItemCount returns the number of items of all states.
func (t *LookupTables) ItemCount() int {
	return len(t.coreByItemIdx)
}

// StateIdx returns the state the item belongs to.
func (t *LookupTables) StateIdx(itemIdx int) int {
	return t.stateIdxByItemIdx[itemIdx]
}

// Core returns the production and the position of the dot of the item.
func (t *LookupTables) Core(itemIdx int) backend.Core {
	return t.coreByItemIdx[itemIdx]
}

// ItemIdx returns the item of the state with the given core, if the state has one.
func (t *LookupTables) ItemIdx(stateIdx int, core backend.Core) (int, bool) {
	if itemIdx, found := t.kernelItemIdx(stateIdx, core); found {
		return itemIdx, true
	}
	// Closure items have the dot at the start.
	if core.Position() != 0 {
		return noItemIdx, false
	}
	closureFrom, closureTo := t.closureItems(stateIdx)
	offset, found := slices.BinarySearchFunc(t.coreByItemIdx[closureFrom:closureTo], core, t.compareClosureCores)
	return closureFrom + offset, found
}

// NextSymbol returns the symbol after the dot of the item. It reports false for a reduce item.
func (t *LookupTables) NextSymbol(itemIdx int) (frontend.SymbolRef, bool) {
	return t.symbolAfterDot(t.coreByItemIdx[itemIdx])
}

// Transition returns the item which the transition on the symbol after the dot leads to, see figure 4(a). It reports
// false for a reduce item, and for an item whose transition a declaration of the grammar removed.
func (t *LookupTables) Transition(itemIdx int) (int, bool) {
	targetItemIdx := t.transitionByItemIdx[itemIdx]
	return targetItemIdx, targetItemIdx != noItemIdx
}

// ReverseTransitions returns the items whose transition leads to the item. It is empty for an item with the dot at the
// start of the production. The returned slice must not be modified.
func (t *LookupTables) ReverseTransitions(itemIdx int) []int {
	return t.reverseTransitions[t.reverseTransitionOffsetByItemIdx[itemIdx]:t.reverseTransitionOffsetByItemIdx[itemIdx+1]]
}

// ProductionSteps returns the range [from, to) of the items a production step from the item leads to, which are the
// closure items of the nonterminal after the dot in the same state, see figure 4(b). The range is empty when no
// nonterminal follows the dot.
func (t *LookupTables) ProductionSteps(itemIdx int) (int, int) {
	symbolRef, found := t.NextSymbol(itemIdx)
	if !found || symbolRef.IsTerminal() {
		return 0, 0
	}
	return t.closureItemRange(t.stateIdxByItemIdx[itemIdx], symbolRef.Idx())
}

// ReverseProductionSteps returns the items of the same state which a production step to the item comes from, which are
// those with the nonterminal of the item after the dot. It is empty for a kernel item. The returned slice must not be
// modified.
func (t *LookupTables) ReverseProductionSteps(itemIdx int) []int {
	stateIdx := t.stateIdxByItemIdx[itemIdx]
	if t.isKernelItem(stateIdx, itemIdx) {
		return nil
	}
	return t.ItemsWithNextSymbol(stateIdx, frontend.NewNonterminalRef(t.productionNonterminalIdx(itemIdx)))
}

// ItemsWithNextSymbol returns the items of the state with the symbol after the dot. The returned slice must not be
// modified.
func (t *LookupTables) ItemsWithNextSymbol(stateIdx int, symbolRef frontend.SymbolRef) []int {
	itemIdxs := t.itemIdxsByNextSymbol[t.itemIdxOffsetByStateIdx[stateIdx]:t.itemIdxOffsetByStateIdx[stateIdx+1]]
	key := int(symbolRef)
	from, _ := slices.BinarySearchFunc(itemIdxs, key, t.compareNextSymbolKey)
	to, _ := slices.BinarySearchFunc(itemIdxs, key+1, t.compareNextSymbolKey)
	return itemIdxs[from:to]
}

// ReachingItems returns every item from which a path of transitions and production steps leads to the item, including
// the item itself. The search for the shortest lookahead-sensitive path only needs to visit those, see section 6,
// "Finding shortest lookahead-sensitive path".
func (t *LookupTables) ReachingItems(itemIdx int) utils.Bitset {
	var result utils.Bitset
	result.Add(itemIdx)
	workList := []int{itemIdx}
	for len(workList) > 0 {
		currentItemIdx := workList[len(workList)-1]
		workList = workList[:len(workList)-1]
		for _, predecessorItemIdx := range t.ReverseTransitions(currentItemIdx) {
			if result.Add(predecessorItemIdx) {
				workList = append(workList, predecessorItemIdx)
			}
		}
		for _, predecessorItemIdx := range t.ReverseProductionSteps(currentItemIdx) {
			if result.Add(predecessorItemIdx) {
				workList = append(workList, predecessorItemIdx)
			}
		}
	}
	return result
}

// initProductionIdxsByNonterminalIdx initializes the helper variable productionIdxsByNonterminalIdx.
func (t *LookupTables) initProductionIdxsByNonterminalIdx() {
	t.productionIdxsByNonterminalIdx = make([][]int, len(t.grammar.Nonterminals))
	for productionIdx, production := range t.grammar.Productions {
		t.productionIdxsByNonterminalIdx[production.NonterminalIdx] = append(
			t.productionIdxsByNonterminalIdx[production.NonterminalIdx],
			productionIdx,
		)
	}
}

// initItems numbers the items of every state: the kernel items, followed by the closure items.
func (t *LookupTables) initItems() {
	t.itemIdxOffsetByStateIdx = make([]int, 0, len(t.states)+1)
	t.closureItemIdxOffsetByStateIdx = make([]int, 0, len(t.states))

	// We keep those variables outside the loop to re-use their allocated memory in subsequent loops.
	var closureNonterminalIdxs utils.Bitset
	var workList []int
	for stateIdx := range t.states {
		t.itemIdxOffsetByStateIdx = append(t.itemIdxOffsetByStateIdx, len(t.coreByItemIdx))
		for _, core := range t.states[stateIdx].KernelItems.All() {
			t.addItem(stateIdx, core)
		}

		t.closureItemIdxOffsetByStateIdx = append(t.closureItemIdxOffsetByStateIdx, len(t.coreByItemIdx))
		closureNonterminalIdxs.Clear()
		workList = t.collectClosureNonterminals(stateIdx, &closureNonterminalIdxs, workList[:0])
		for nonterminalIdx := range closureNonterminalIdxs.All() {
			for _, productionIdx := range t.productionIdxsByNonterminalIdx[nonterminalIdx] {
				t.addItem(stateIdx, backend.NewCore(productionIdx, 0))
			}
		}
	}
	t.itemIdxOffsetByStateIdx = append(t.itemIdxOffsetByStateIdx, len(t.coreByItemIdx))

	utils.DebugAssert(func() error {
		// The closure never adds the start production, as the start nonterminal of the augmented grammar is on no right
		// hand side, so it is the only kernel item with the dot at the start, and only of the start state.
		for stateIdx := range t.states {
			kernelFrom, kernelTo := t.kernelItems(stateIdx)
			for itemIdx := kernelFrom; itemIdx < kernelTo; itemIdx++ {
				if t.coreByItemIdx[itemIdx].Position() == 0 && (stateIdx != 0 || t.coreByItemIdx[itemIdx].ProductionIdx() != 0) {
					return fmt.Errorf("state %d has the kernel item %s with the dot at the start", stateIdx, t.coreByItemIdx[itemIdx])
				}
			}
		}
		return nil
	})
}

// addItem appends an item to the items of the state.
func (t *LookupTables) addItem(stateIdx int, core backend.Core) {
	t.stateIdxByItemIdx = append(t.stateIdxByItemIdx, stateIdx)
	t.coreByItemIdx = append(t.coreByItemIdx, core)
}

// collectClosureNonterminals adds the nonterminals whose productions make up the closure of the state to the set. The
// work list is scratch space which is returned for reuse.
func (t *LookupTables) collectClosureNonterminals(
	stateIdx int,
	nonterminalIdxs *utils.Bitset,
	workList []int,
) []int {
	for _, core := range t.states[stateIdx].KernelItems.All() {
		symbolRef, found := t.symbolAfterDot(core)
		if found && symbolRef.IsNonterminal() && nonterminalIdxs.Add(symbolRef.Idx()) {
			workList = append(workList, symbolRef.Idx())
		}
	}
	for len(workList) > 0 {
		nonterminalIdx := workList[len(workList)-1]
		workList = workList[:len(workList)-1]
		for _, productionIdx := range t.productionIdxsByNonterminalIdx[nonterminalIdx] {
			symbolRefs := t.grammar.Productions[productionIdx].SymbolRefs
			if len(symbolRefs) > 0 && symbolRefs[0].IsNonterminal() && nonterminalIdxs.Add(symbolRefs[0].Idx()) {
				workList = append(workList, symbolRefs[0].Idx())
			}
		}
	}
	return workList
}

// initTransitions finds the transition of every item and inverts them.
func (t *LookupTables) initTransitions() {
	t.transitionByItemIdx = make([]int, len(t.coreByItemIdx))
	t.reverseTransitionOffsetByItemIdx = make([]int, len(t.coreByItemIdx)+1)
	for itemIdx := range t.coreByItemIdx {
		targetItemIdx := t.findTransition(itemIdx)
		t.transitionByItemIdx[itemIdx] = targetItemIdx
		if targetItemIdx != noItemIdx {
			t.reverseTransitionOffsetByItemIdx[targetItemIdx+1]++
		}
	}

	// The counts become offsets, and every reverse transition is written behind the ones of its item written so far.
	for itemIdx := range t.coreByItemIdx {
		t.reverseTransitionOffsetByItemIdx[itemIdx+1] += t.reverseTransitionOffsetByItemIdx[itemIdx]
	}
	t.reverseTransitions = make([]int, t.reverseTransitionOffsetByItemIdx[len(t.coreByItemIdx)])
	writeOffsetByItemIdx := slices.Clone(t.reverseTransitionOffsetByItemIdx)
	for itemIdx, targetItemIdx := range t.transitionByItemIdx {
		if targetItemIdx == noItemIdx {
			continue
		}
		t.reverseTransitions[writeOffsetByItemIdx[targetItemIdx]] = itemIdx
		writeOffsetByItemIdx[targetItemIdx]++
	}
}

// findTransition returns the item which the transition on the symbol after the dot of the item leads to, or noItemIdx.
func (t *LookupTables) findTransition(itemIdx int) int {
	core := t.coreByItemIdx[itemIdx]
	symbolRef, found := t.symbolAfterDot(core)
	if !found {
		return noItemIdx
	}

	targetStateIdx, found := t.transitionStateIdx(t.stateIdxByItemIdx[itemIdx], symbolRef)
	if !found {
		// Only a declaration of the grammar removes a transition, see conflict.Resolve.
		return noItemIdx
	}

	targetCore := backend.NewCore(core.ProductionIdx(), core.Position()+1)
	targetItemIdx, found := t.kernelItemIdx(targetStateIdx, targetCore)
	utils.DebugAssert(func() error {
		if !found {
			return fmt.Errorf("state %d has no kernel item %s behind its transition", targetStateIdx, targetCore)
		}
		return nil
	})
	return targetItemIdx
}

// transitionStateIdx returns the state the transition of the state on the symbol leads to, if it has one.
func (t *LookupTables) transitionStateIdx(stateIdx int, symbolRef frontend.SymbolRef) (int, bool) {
	transitionActions := &t.states[stateIdx].TransitionActions
	transitionIdx := transitionActions.LowerBound(backend.NewTransitionAction(symbolRef, 0))
	if transitionIdx == transitionActions.Length() {
		return 0, false
	}
	if transitionActions.GetByIndex(transitionIdx).SymbolRef() != symbolRef {
		return 0, false
	}
	return transitionActions.GetByIndex(transitionIdx).StateIdx(), true
}

// kernelItems returns the range [from, to) of the kernel items of the state.
func (t *LookupTables) kernelItems(stateIdx int) (int, int) {
	return t.itemIdxOffsetByStateIdx[stateIdx], t.closureItemIdxOffsetByStateIdx[stateIdx]
}

// closureItems returns the range [from, to) of the closure items of the state. They are the last items of the state.
func (t *LookupTables) closureItems(stateIdx int) (int, int) {
	return t.closureItemIdxOffsetByStateIdx[stateIdx], t.itemIdxOffsetByStateIdx[stateIdx+1]
}

// isKernelItem reports if the item of the state is one of its kernel items.
func (t *LookupTables) isKernelItem(stateIdx int, itemIdx int) bool {
	return itemIdx < t.closureItemIdxOffsetByStateIdx[stateIdx]
}

// kernelItemIdx returns the kernel item of the state with the given core, if the state has one.
func (t *LookupTables) kernelItemIdx(stateIdx int, core backend.Core) (int, bool) {
	kernelFrom, kernelTo := t.kernelItems(stateIdx)
	offset, found := slices.BinarySearch(t.coreByItemIdx[kernelFrom:kernelTo], core)
	return kernelFrom + offset, found
}

// closureItemRange returns the range [from, to) of the closure items of the nonterminal in the state. A closure holds
// every production of its nonterminals, so the range has as many items as the nonterminal has productions.
func (t *LookupTables) closureItemRange(stateIdx int, nonterminalIdx int) (int, int) {
	closureFrom, closureTo := t.closureItems(stateIdx)
	offset, found := slices.BinarySearchFunc(
		t.coreByItemIdx[closureFrom:closureTo],
		nonterminalIdx,
		t.compareClosureCoreNonterminal,
	)
	utils.DebugAssert(func() error {
		if !found {
			return fmt.Errorf("state %d has no closure items of nonterminal %d", stateIdx, nonterminalIdx)
		}
		return nil
	})
	from := closureFrom + offset
	return from, from + len(t.productionIdxsByNonterminalIdx[nonterminalIdx])
}

// compareClosureCores orders closure items by their nonterminal and then by their production index, which is the order
// of the closure items of a state.
func (t *LookupTables) compareClosureCores(a backend.Core, b backend.Core) int {
	aNonterminalIdx := t.grammar.Productions[a.ProductionIdx()].NonterminalIdx
	bNonterminalIdx := t.grammar.Productions[b.ProductionIdx()].NonterminalIdx
	if result := cmp.Compare(aNonterminalIdx, bNonterminalIdx); result != 0 {
		return result
	}
	return cmp.Compare(a.ProductionIdx(), b.ProductionIdx())
}

// compareClosureCoreNonterminal compares the nonterminal of a closure item with a nonterminal.
func (t *LookupTables) compareClosureCoreNonterminal(core backend.Core, nonterminalIdx int) int {
	return cmp.Compare(t.grammar.Productions[core.ProductionIdx()].NonterminalIdx, nonterminalIdx)
}

// initItemIdxsByNextSymbol orders the items of every state by the symbol after the dot.
func (t *LookupTables) initItemIdxsByNextSymbol() {
	t.itemIdxsByNextSymbol = make([]int, len(t.coreByItemIdx))
	for itemIdx := range t.itemIdxsByNextSymbol {
		t.itemIdxsByNextSymbol[itemIdx] = itemIdx
	}
	for stateIdx := range t.states {
		itemIdxs := t.itemIdxsByNextSymbol[t.itemIdxOffsetByStateIdx[stateIdx]:t.itemIdxOffsetByStateIdx[stateIdx+1]]
		slices.SortStableFunc(itemIdxs, t.compareNextSymbols)
	}
}

// compareNextSymbols orders two items by the symbol after the dot, see nextSymbolKey.
func (t *LookupTables) compareNextSymbols(aItemIdx int, bItemIdx int) int {
	return cmp.Compare(t.nextSymbolKey(aItemIdx), t.nextSymbolKey(bItemIdx))
}

// compareNextSymbolKey compares the symbol after the dot of an item with the key of a symbol, see nextSymbolKey.
func (t *LookupTables) compareNextSymbolKey(itemIdx int, key int) int {
	return cmp.Compare(t.nextSymbolKey(itemIdx), key)
}

// nextSymbolKey returns the symbol after the dot of the item as an integer, or noNextSymbolKey for a reduce item.
func (t *LookupTables) nextSymbolKey(itemIdx int) int {
	symbolRef, found := t.NextSymbol(itemIdx)
	if !found {
		return noNextSymbolKey
	}
	return int(symbolRef)
}

// restBehindNextSymbol returns the symbols of the production of the item behind the symbol after the dot. The item must
// not be a reduce item. The returned slice must not be modified.
func (t *LookupTables) restBehindNextSymbol(itemIdx int) []frontend.SymbolRef {
	core := t.coreByItemIdx[itemIdx]
	return t.grammar.Productions[core.ProductionIdx()].SymbolRefs[core.Position()+1:]
}

// symbolAfterDot returns the symbol after the dot of the core. It reports false when the dot is at the end.
func (t *LookupTables) symbolAfterDot(core backend.Core) (frontend.SymbolRef, bool) {
	symbolRefs := t.grammar.Productions[core.ProductionIdx()].SymbolRefs
	if core.Position() == len(symbolRefs) {
		return 0, false
	}
	return symbolRefs[core.Position()], true
}
