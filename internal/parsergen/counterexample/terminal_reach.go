package counterexample

import (
	"math"

	"github.com/backbone81/golr/internal/parsergen/conflict"
	"github.com/backbone81/golr/internal/utils"
)

// terminalReach tells for every item how the parser carries a terminal from the item on, taking only actions the
// declarations of the grammar leave to it, see LookupTables.IsActionAllowed. The terminal is the next one of the input
// all along: it is the lookahead of every reduction, and the parser shifts it at the end.
//
// Deviating from the paper, the lookahead-sensitive graph of section 4 and the derivations behind the dot follow the
// reach instead of FIRST sets and nullable nonterminals. Those only tell that the grammar lets the terminal follow,
// not that the parser does: a declaration can remove the shift of the terminal, or the reduction it needs before,
// in a state the counterexample goes through, and the counterexample would show an input the parser rejects. The paper
// exploits precedence in the unifying search only (section 6, "Exploiting precedence").
//
// For example, with the terminal "+" and the grammar s : ε | s "+" s | "+" "x", the item s -> s "+" • s of the state
// behind s "+" reaches the shift of "+" by entering s -> • "+" "x", and lets s vanish by entering s -> • and reducing
// it on "+". It reduces its own production on "+" when s vanishes and the state behind s "+" s reduces
// s -> s "+" s on "+", which only a left associative "+" allows.
type terminalReach struct {
	terminalIdx int

	// byItemIdx holds the reach of the terminal from every item.
	byItemIdx []itemReach
}

// itemReach is the reach of the terminal from one item, see terminalReach.
type itemReach struct {
	// toShift is the cheapest way from the item to the shift of the terminal: the terminal is the symbol after the dot,
	// or the parser enters a production of the nonterminal after the dot which begins with the terminal, or the
	// nonterminal vanishes and the way goes on behind it. Only entering a production names a closure item.
	toShift way

	// toVanish is the cheapest derivation of the empty string from the nonterminal after the dot, which enters a
	// production of the nonterminal whose rest vanishes and is reduced, see toReduce.
	toVanish way

	// toReduce is the cost of the cheapest way to derive the rest of the production behind the dot to the empty string,
	// followed by the reduction of the production. It is 0 for a reduce item whose reduction the declarations allow, and
	// noCost when there is no way. The way enters no closure item of the item itself, the nonterminals of the rest
	// vanish along toVanish of their items.
	toReduce int
}

// way is the cheapest way the parser takes from an item, see itemReach.
type way struct {
	// cost is the number of productions of the derivation the way builds, or noCost when there is no way. The cost only
	// picks the shortest derivation, which keeps the counterexample short (section 3.2).
	cost int

	// closureItemIdx is the closure item of the nonterminal after the dot which the way enters next, or noItemIdx when
	// it enters none.
	closureItemIdx int
}

// noCost stands for a missing way, which is more expensive than any way there is.
const noCost = math.MaxInt

// noWay is a way which does not exist.
var noWay = way{cost: noCost, closureItemIdx: noItemIdx}

// CanShift reports if the parser can carry the terminal from the item to its shift.
func (r *terminalReach) CanShift(itemIdx int) bool {
	return r.byItemIdx[itemIdx].toShift.cost != noCost
}

// CanVanish reports if the nonterminal after the dot of the item can derive the empty string.
func (r *terminalReach) CanVanish(itemIdx int) bool {
	return r.byItemIdx[itemIdx].toVanish.cost != noCost
}

// CanReduce reports if the rest of the production of the item can derive the empty string, so the parser reduces the
// production.
func (r *terminalReach) CanReduce(itemIdx int) bool {
	return r.byItemIdx[itemIdx].toReduce != noCost
}

// terminalReach returns the reach of the terminal. It is computed on first use of the terminal.
func (t *LookupTables) terminalReach(terminalIdx int) *terminalReach {
	if reach, found := t.terminalReachByTerminalIdx[terminalIdx]; found {
		return reach
	}
	builder := newTerminalReachBuilder(t, terminalIdx)
	reach := builder.Build()
	t.terminalReachByTerminalIdx[terminalIdx] = reach
	return reach
}

// terminalReachBuilder computes the reach of a terminal in a fixed-point computation over the items, which repeats the
// items whose reach depends on an item whose reach got cheaper until none does.
type terminalReachBuilder struct {
	tables *LookupTables
	result *terminalReach
}

// newTerminalReachBuilder returns a builder for the reach of the terminal.
func newTerminalReachBuilder(tables *LookupTables, terminalIdx int) terminalReachBuilder {
	result := &terminalReach{
		terminalIdx: terminalIdx,
		byItemIdx:   make([]itemReach, tables.ItemCount()),
	}
	for itemIdx := range result.byItemIdx {
		result.byItemIdx[itemIdx] = itemReach{toShift: noWay, toVanish: noWay, toReduce: noCost}
	}
	return terminalReachBuilder{tables: tables, result: result}
}

// Build returns the reach.
func (b *terminalReachBuilder) Build() *terminalReach {
	itemCount := b.tables.ItemCount()
	workList := utils.NewDynamicRingBufferWithCapacity[int](itemCount)
	queued := make([]bool, itemCount)
	for itemIdx := range itemCount {
		workList.Add(itemIdx)
		queued[itemIdx] = true
	}
	for !workList.IsEmpty() {
		itemIdx := workList.Remove()
		queued[itemIdx] = false
		if !b.update(itemIdx) {
			continue
		}
		// The items whose reach reads the reach of this one: those whose transition leads to it, and, for a closure
		// item, those which enter it by a production step.
		for _, dependentItemIdx := range b.tables.ReverseTransitions(itemIdx) {
			if !queued[dependentItemIdx] {
				workList.Add(dependentItemIdx)
				queued[dependentItemIdx] = true
			}
		}
		for _, dependentItemIdx := range b.tables.ReverseProductionSteps(itemIdx) {
			if !queued[dependentItemIdx] {
				workList.Add(dependentItemIdx)
				queued[dependentItemIdx] = true
			}
		}
	}
	return b.result
}

// update computes the reach of the item from the reach known so far, and reports if any part of it got cheaper.
func (b *terminalReachBuilder) update(itemIdx int) bool {
	symbolRef, found := b.tables.NextSymbol(itemIdx)
	if !found {
		return b.updateReduceItem(itemIdx)
	}
	if symbolRef.IsTerminal() {
		return b.updateTerminalItem(itemIdx)
	}

	// The cheapest closure items of the nonterminal after the dot, to the shift and to the empty string.
	toShift, toVanish := noWay, noWay
	from, to := b.tables.ProductionSteps(itemIdx)
	for closureItemIdx := from; closureItemIdx < to; closureItemIdx++ {
		closureReach := b.result.byItemIdx[closureItemIdx]
		if cost := addCosts(1, closureReach.toShift.cost); cost < toShift.cost {
			toShift = way{cost: cost, closureItemIdx: closureItemIdx}
		}
		if cost := addCosts(1, closureReach.toReduce); cost < toVanish.cost {
			toVanish = way{cost: cost, closureItemIdx: closureItemIdx}
		}
	}

	// The nonterminal vanishes, and the way goes on behind it.
	toReduce := noCost
	if targetItemIdx, found := b.tables.Transition(itemIdx); found {
		targetReach := b.result.byItemIdx[targetItemIdx]
		toReduce = addCosts(toVanish.cost, targetReach.toReduce)
		if cost := addCosts(toVanish.cost, targetReach.toShift.cost); cost < toShift.cost {
			toShift = way{cost: cost, closureItemIdx: noItemIdx}
		}
	}

	reach := &b.result.byItemIdx[itemIdx]
	changed := false
	if toShift.cost < reach.toShift.cost {
		reach.toShift = toShift
		changed = true
	}
	if toVanish.cost < reach.toVanish.cost {
		reach.toVanish = toVanish
		changed = true
	}
	if toReduce < reach.toReduce {
		reach.toReduce = toReduce
		changed = true
	}
	return changed
}

// updateReduceItem makes the reduction of the item free when the declarations allow it on the terminal, and reports
// if that changed the reach.
func (b *terminalReachBuilder) updateReduceItem(itemIdx int) bool {
	reach := &b.result.byItemIdx[itemIdx]
	contribution := conflict.NewReduceContribution(b.tables.Core(itemIdx).ProductionIdx())
	if reach.toReduce == 0 ||
		!b.tables.IsActionAllowed(b.tables.StateIdx(itemIdx), b.result.terminalIdx, contribution) {
		return false
	}
	reach.toReduce = 0
	return true
}

// updateTerminalItem makes the shift of the item free when the terminal after the dot is the terminal and the
// declarations allow its shift, and reports if that changed the reach.
func (b *terminalReachBuilder) updateTerminalItem(itemIdx int) bool {
	reach := &b.result.byItemIdx[itemIdx]
	symbolRef, _ := b.tables.NextSymbol(itemIdx)
	if reach.toShift.cost == 0 || symbolRef.Idx() != b.result.terminalIdx {
		return false
	}
	if _, found := b.tables.Transition(itemIdx); !found ||
		!b.tables.IsActionAllowed(b.tables.StateIdx(itemIdx), b.result.terminalIdx, conflict.NewShiftContribution()) {
		return false
	}
	reach.toShift = way{cost: 0, closureItemIdx: noItemIdx}
	return true
}

// addCosts returns the sum of the costs, which is noCost when either one is.
func addCosts(a int, b int) int {
	if a == noCost || b == noCost {
		return noCost
	}
	return a + b
}
