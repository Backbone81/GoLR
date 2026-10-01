package counterexample

import (
	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/utils"
)

// Lookahead returns the lookahead set of the item, which holds every terminal that can follow the production of the
// item in its state. It is the union of the precise lookahead sets of section 4 over all vertices of the
// lookahead-sensitive graph with this item. Stage 1 of the unifying search requires a prepended item to hold the
// conflict terminal in it (section 5.3). The returned set must not be modified.
func (t *LookupTables) Lookahead(itemIdx int) *backend.LookaheadSet {
	return &t.lookaheads[t.lookaheadIdxByItemIdx[itemIdx]]
}

// initLookaheads computes the lookahead set of every item by propagating lookaheads over the states the core built, the
// way canonical LR(1) computes the lookahead sets of its items. On the states of any core, the lookahead set of a
// reduce item is then the lookahead set the core computed for the reduction, before the conflicts were resolved.
// Nothing is taken from the reduce actions, which the conflict resolution changed.
//
// The closure items of a nonterminal in a state always have the same lookahead set, so they share one. That keeps the
// sets few for grammars with large closures.
func (t *LookupTables) initLookaheads() {
	t.initLookaheadIdxs()
	t.lookaheads = make([]backend.LookaheadSet, len(t.itemIdxOffsetByLookaheadIdx)-1)
	for itemIdx := range t.coreByItemIdx {
		t.addSpontaneousLookaheads(itemIdx)
	}
	t.propagateLookaheads()
}

// initLookaheadIdxs assigns a lookahead set to every item: one of its own to every kernel item, and one to the closure
// items of every nonterminal of a state. The items of a lookahead set are adjacent.
func (t *LookupTables) initLookaheadIdxs() {
	t.lookaheadIdxByItemIdx = make([]int, len(t.coreByItemIdx))
	for itemIdx := range t.coreByItemIdx {
		if t.startsLookahead(itemIdx) {
			t.itemIdxOffsetByLookaheadIdx = append(t.itemIdxOffsetByLookaheadIdx, itemIdx)
		}
		t.lookaheadIdxByItemIdx[itemIdx] = len(t.itemIdxOffsetByLookaheadIdx) - 1
	}
	t.itemIdxOffsetByLookaheadIdx = append(t.itemIdxOffsetByLookaheadIdx, len(t.coreByItemIdx))
}

// startsLookahead reports if the item gets a lookahead set of its own instead of sharing the one of the item before it.
// That is every kernel item, the first closure item, and every closure item whose nonterminal differs from the item
// before it.
func (t *LookupTables) startsLookahead(itemIdx int) bool {
	stateIdx := t.stateIdxByItemIdx[itemIdx]
	if t.isKernelItem(stateIdx, itemIdx) {
		return true
	}
	closureFrom, _ := t.closureItems(stateIdx)
	return itemIdx == closureFrom || t.productionNonterminalIdx(itemIdx-1) != t.productionNonterminalIdx(itemIdx)
}

// productionNonterminalIdx returns the nonterminal on the left hand side of the production of the item.
func (t *LookupTables) productionNonterminalIdx(itemIdx int) int {
	return t.grammar.Productions[t.coreByItemIdx[itemIdx].ProductionIdx()].NonterminalIdx
}

// addSpontaneousLookaheads adds the terminals which start the rest of the item behind the nonterminal after the dot to
// the lookahead set of the closure items of that nonterminal. They are the terminal and FIRST cases of followL of
// section 4, which hold whatever the lookahead set of the item is.
func (t *LookupTables) addSpontaneousLookaheads(itemIdx int) {
	from, to := t.ProductionSteps(itemIdx)
	if from == to {
		return
	}
	core := t.coreByItemIdx[itemIdx]
	rest := t.grammar.Productions[core.ProductionIdx()].SymbolRefs[core.Position()+1:]
	t.firstSets.FirstOfSequence(rest, &t.lookaheads[t.lookaheadIdxByItemIdx[from]])
}

// propagateLookaheads merges the lookahead set of every item into the items it leads to until nothing changes: along a
// transition, and along a production step when the rest of the item behind the nonterminal can vanish, which is the
// last case of followL of section 4.
func (t *LookupTables) propagateLookaheads() {
	workList := utils.NewDynamicRingBufferWithCapacity[int](len(t.lookaheads))
	queued := make([]bool, len(t.lookaheads))
	for lookaheadIdx := range t.lookaheads {
		workList.Add(lookaheadIdx)
		queued[lookaheadIdx] = true
	}

	for !workList.IsEmpty() {
		lookaheadIdx := workList.Remove()
		queued[lookaheadIdx] = false
		firstItemIdx := t.itemIdxOffsetByLookaheadIdx[lookaheadIdx]
		endItemIdx := t.itemIdxOffsetByLookaheadIdx[lookaheadIdx+1]
		for itemIdx := firstItemIdx; itemIdx < endItemIdx; itemIdx++ {
			if targetItemIdx, found := t.Transition(itemIdx); found {
				t.mergeLookahead(lookaheadIdx, t.lookaheadIdxByItemIdx[targetItemIdx], &workList, queued)
			}
			from, to := t.ProductionSteps(itemIdx)
			if from == to {
				continue
			}
			core := t.coreByItemIdx[itemIdx]
			rest := t.grammar.Productions[core.ProductionIdx()].SymbolRefs[core.Position()+1:]
			if t.firstSets.IsSequenceNullable(rest) {
				t.mergeLookahead(lookaheadIdx, t.lookaheadIdxByItemIdx[from], &workList, queued)
			}
		}
	}
}

// mergeLookahead merges one lookahead set into another, and queues the other when that made it grow.
func (t *LookupTables) mergeLookahead(
	fromLookaheadIdx int,
	toLookaheadIdx int,
	workList *utils.DynamicRingBuffer[int],
	queued []bool,
) {
	if !t.lookaheads[toLookaheadIdx].Merge(&t.lookaheads[fromLookaheadIdx]) || queued[toLookaheadIdx] {
		return
	}
	workList.Add(toLookaheadIdx)
	queued[toLookaheadIdx] = true
}
