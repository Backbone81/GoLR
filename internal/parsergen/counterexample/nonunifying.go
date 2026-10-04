package counterexample

import (
	"fmt"
	"slices"

	"github.com/backbone81/golr/internal/utils"
)

// nonunifyingBuilder builds the nonunifying counterexample of a pair of conflict items, see section 4. The first
// derivation completes a shortest lookahead-sensitive path to the reduce item. The second one walks back along that
// path from the other item until it meets an item of the path, and shares the path up to there.
type nonunifyingBuilder struct {
	tables        *LookupTables
	reduceItemIdx int
	otherItemIdx  int
	terminalIdx   int
	reach         *terminalReach
	reducePath    []pathItem
}

// walkNode is an item reached by a backward walk, and if the conflict terminal is still required behind its production.
type walkNode struct {
	itemIdx  int
	required bool
}

// walkStep is a node of a backward walk together with the node it was reached from.
type walkStep struct {
	node walkNode

	// nextStepIdx is the step the walk came from, which is the next one in forward direction, or noWalkStepIdx for the
	// item the walk started at.
	nextStepIdx int

	// productionStep reports if the step from this node to the next one is a production step, not a transition.
	productionStep bool
}

// noWalkStepIdx stands for a missing walk step, like the next step of the item the walk started at.
const noWalkStepIdx = -1

// newNonunifyingBuilder returns a builder for the counterexample of the reduce item and the other item on the terminal.
// The other item is an item which shifts the terminal, or a second reduce item. The path is the shortest
// lookahead-sensitive path to the reduce item.
func newNonunifyingBuilder(
	tables *LookupTables,
	reduceItemIdx int,
	otherItemIdx int,
	terminalIdx int,
	reducePath []pathItem,
) nonunifyingBuilder {
	return nonunifyingBuilder{
		tables:        tables,
		reduceItemIdx: reduceItemIdx,
		otherItemIdx:  otherItemIdx,
		terminalIdx:   terminalIdx,
		reach:         tables.terminalReach(terminalIdx),
		reducePath:    reducePath,
	}
}

// Build returns the counterexample. Its derivations start at the start symbol, or at the start production when the
// conflict terminal is the end of the input, which only the start production shows. It reports false when the
// declarations of the grammar leave the parser no input in which the reduction of a second reduce item is followed by
// the terminal.
func (b *nonunifyingBuilder) Build() (Counterexample, bool) {
	// The conflict terminal is the symbol after the dot of a shift item, so nothing is required behind it. A second
	// reduce item needs the terminal behind its production like the first one (p. 4, footnote 4).
	_, isShift := b.tables.NextSymbol(b.otherItemIdx)
	if !isShift && !b.reach.CanReduce(b.otherItemIdx) {
		return Counterexample{}, false
	}
	otherPath, found := b.walkBack(b.reducePath, walkNode{itemIdx: b.otherItemIdx, required: !isShift})
	if !found {
		// Only a second reduce item can miss the path, when the states the path goes through merged contexts in which
		// the terminal does not follow it. Its own shortest path shares less with the first one, but is valid.
		otherPath, found = b.tables.shortestLookaheadSensitivePath(b.otherItemIdx, b.terminalIdx)
		if !found {
			return Counterexample{}, false
		}
	}

	derivations := [2]Derivation{
		b.derivation(b.reducePath, true),
		b.derivation(otherPath, !isShift),
	}
	if b.terminalIdx != eofTerminalIdx {
		for i := range derivations {
			derivations[i] = derivations[i].Children[0]
		}
	}
	return Counterexample{
		NonterminalIdx: derivations[0].Symbol.Idx(),
		Derivations:    derivations,
	}, true
}

// walkBack explores backward from the node along the states of the path, until it meets an item of the path at the same
// point of the path, see figure 5(b). It returns the path up to that item, followed by the items the walk went through
// in forward direction, ending at the node. Points of the path are the runs of items between two transitions, which
// share a state. A later meeting point is preferred over an earlier one, because it shares more of the path.
//
// Within a state, the walk takes reverse production steps. A kernel item takes its reverse transition into the state of
// the previous point, which is unique, as that state holds the item with the dot one symbol earlier. While the terminal
// is required behind the production of the node, a reverse production step only goes to a parent from whose goto the
// parser carries the terminal to its shift, which satisfies it, or whose symbols behind the nonterminal vanish and
// whose production is reduced on the terminal, which keeps it required, see terminalReach. A meeting item then has to
// hold the terminal in its precise lookahead set. The paper leaves this out, as it only walks back from a shift item,
// behind which nothing is required.
//
// It reports false when the walk does not meet the path, which only happens while the terminal is required.
func (b *nonunifyingBuilder) walkBack(path []pathItem, node walkNode) ([]pathItem, bool) {
	pointStarts := pathPointStarts(path)
	steps := []walkStep{{node: node, nextStepIdx: noWalkStepIdx}}
	entryFrom := 0
	for point, pointFrom := range slices.Backward(pointStarts) {
		pointTo := len(path)
		if point+1 < len(pointStarts) {
			pointTo = pointStarts[point+1]
		}

		visited := make(map[walkNode]struct{})
		for stepIdx := entryFrom; stepIdx < len(steps); stepIdx++ {
			visited[steps[stepIdx].node] = struct{}{}
		}
		for stepIdx := entryFrom; stepIdx < len(steps); stepIdx++ {
			if pathIdx, found := b.meetingPathIdx(path[pointFrom:pointTo], steps[stepIdx].node); found {
				return b.joinWalk(path[:pointFrom+pathIdx+1], steps, stepIdx), true
			}
			steps = b.appendReverseProductionSteps(steps, stepIdx, visited)
		}

		if point == 0 {
			break
		}
		previousStateIdx := b.tables.StateIdx(path[pointFrom-1].itemIdx)
		nextEntryFrom := len(steps)
		steps = b.appendReverseTransitions(steps, entryFrom, previousStateIdx)
		entryFrom = nextEntryFrom
	}
	return nil, false
}

// pathPointStarts returns the index of the first item of every point of the path, see walkBack.
func pathPointStarts(path []pathItem) []int {
	result := []int{0}
	for pathIdx := 1; pathIdx < len(path); pathIdx++ {
		if !path[pathIdx].productionStep {
			result = append(result, pathIdx)
		}
	}
	return result
}

// meetingPathIdx returns the last index of the items of the point at which the node meets the path: the same item, with
// the terminal in its precise lookahead set while the terminal is required.
func (b *nonunifyingBuilder) meetingPathIdx(pointItems []pathItem, node walkNode) (int, bool) {
	for pathIdx, pointItem := range slices.Backward(pointItems) {
		if pointItem.itemIdx != node.itemIdx {
			continue
		}
		if !node.required || pointItem.terminalFollows {
			return pathIdx, true
		}
	}
	return 0, false
}

// appendReverseProductionSteps appends the parents of the node of the step within its state which were not visited yet.
func (b *nonunifyingBuilder) appendReverseProductionSteps(
	steps []walkStep,
	stepIdx int,
	visited map[walkNode]struct{},
) []walkStep {
	node := steps[stepIdx].node
	for _, parentItemIdx := range b.tables.ReverseProductionSteps(node.itemIdx) {
		parent, allowed := b.reverseProductionStep(node, parentItemIdx)
		if !allowed {
			continue
		}
		if _, found := visited[parent]; found {
			continue
		}
		visited[parent] = struct{}{}
		steps = append(steps, walkStep{node: parent, nextStepIdx: stepIdx, productionStep: true})
	}
	return steps
}

// reverseProductionStep returns the node of the parent item, and reports false when the parent is excluded because the
// parser can neither carry the required terminal from its goto to the shift, nor reduce its production on it.
func (b *nonunifyingBuilder) reverseProductionStep(node walkNode, parentItemIdx int) (walkNode, bool) {
	parent := walkNode{itemIdx: parentItemIdx}
	if !node.required {
		return parent, true
	}
	targetItemIdx, found := b.tables.Transition(parentItemIdx)
	if !found {
		return parent, false
	}
	if b.reach.CanShift(targetItemIdx) {
		return parent, true
	}
	parent.required = true
	return parent, b.reach.CanReduce(targetItemIdx)
}

// appendReverseTransitions appends the reverse transition into the previous state of every kernel item among the steps
// from the index on. The items with the dot at the start have none, they are closure items or the start item.
func (b *nonunifyingBuilder) appendReverseTransitions(
	steps []walkStep,
	entryFrom int,
	previousStateIdx int,
) []walkStep {
	stepCount := len(steps)
	for stepIdx := entryFrom; stepIdx < stepCount; stepIdx++ {
		node := steps[stepIdx].node
		for _, predecessorItemIdx := range b.tables.ReverseTransitions(node.itemIdx) {
			if b.tables.StateIdx(predecessorItemIdx) != previousStateIdx {
				continue
			}
			steps = append(steps, walkStep{
				node:        walkNode{itemIdx: predecessorItemIdx, required: node.required},
				nextStepIdx: stepIdx,
			})
		}
	}
	return steps
}

// joinWalk returns the shared part of the path, which ends at the meeting item, followed by the items of the walk from
// the step after the meeting item to the item the walk started at.
func (b *nonunifyingBuilder) joinWalk(sharedPath []pathItem, steps []walkStep, meetingStepIdx int) []pathItem {
	result := slices.Clone(sharedPath)
	for stepIdx := meetingStepIdx; steps[stepIdx].nextStepIdx != noWalkStepIdx; stepIdx = steps[stepIdx].nextStepIdx {
		result = append(result, pathItem{
			itemIdx:        steps[steps[stepIdx].nextStepIdx].node.itemIdx,
			productionStep: steps[stepIdx].productionStep,
		})
	}
	return result
}

// derivation completes every production on the path, from the innermost one outward, see section 4. The innermost
// production is the one of the last item of the path, which gets the dot. Every enclosing production expands the
// nonterminal its last item on the path has after the dot into the derivation of the production inside it. The symbols
// in front of the dot of every production are the symbols of the transitions on the path, and stay leaves.
//
// When the terminal is required, it has to follow the dot directly, see appendRest. The precise lookahead sets of the
// path guarantee that the start production satisfies it at the latest.
func (b *nonunifyingBuilder) derivation(path []pathItem, required bool) Derivation {
	var result Derivation
	productionTo := len(path)
	for productionFrom, item := range slices.Backward(path) {
		if !item.productionStep {
			continue
		}
		restItemIdx := path[productionTo-1].itemIdx
		core := b.tables.Core(restItemIdx)
		symbolRefs := b.tables.grammar.Productions[core.ProductionIdx()].SymbolRefs
		children := make([]Derivation, 0, len(symbolRefs)+1)
		for _, symbolRef := range symbolRefs[:core.Position()] {
			children = append(children, NewLeafDerivation(symbolRef))
		}

		if productionTo == len(path) {
			children = append(children, NewDotDerivation())
		} else {
			children = append(children, result)
			restItemIdx, _ = b.tables.Transition(restItemIdx)
		}
		var possible bool
		children, required, possible = b.appendRest(children, restItemIdx, required)
		utils.DebugAssert(func() error {
			if !possible {
				return fmt.Errorf("terminal %d cannot follow the dot of the derivation", b.terminalIdx)
			}
			return nil
		})
		result = NewExpandedDerivation(b.tables.grammar, core.ProductionIdx(), children)
		productionTo = productionFrom
	}

	utils.DebugAssert(func() error {
		if required {
			return fmt.Errorf("terminal %d does not follow the dot of the derivation", b.terminalIdx)
		}
		return nil
	})
	return result
}

// appendRest appends the symbols of the production of the item from its dot on, which are behind the dot or the
// expanded nonterminal of the production. While the terminal is required, they follow the cheapest way the parser
// carries it from the item to its shift, see terminalReach: the terminal itself, a nonterminal expanded into a
// derivation which begins with it, and nonterminals in front of it derived to the empty string, so the terminal
// directly follows the dot in the example. When there is no such way, but the rest vanishes and the production is
// reduced on the terminal, the rest is derived to the empty string and the terminal stays required, which passes the
// requirement on to the enclosing production. Once it is satisfied, the symbols stay leaves.
//
// It returns if the terminal is still required, and reports false when the parser can neither carry it to its shift
// nor reduce the production on it.
func (b *nonunifyingBuilder) appendRest(children []Derivation, itemIdx int, required bool) ([]Derivation, bool, bool) {
	switch {
	case !required:
		return b.tables.appendLeaves(children, itemIdx), false, true
	case b.reach.CanShift(itemIdx):
		return b.tables.appendShift(b.reach, children, itemIdx), false, true
	case b.reach.CanReduce(itemIdx):
		return b.tables.appendEmptyRest(b.reach, children, itemIdx), true, true
	default:
		return children, true, false
	}
}

// appendLeaves appends the symbols of the production of the item from its dot on as leaves.
func (t *LookupTables) appendLeaves(children []Derivation, itemIdx int) []Derivation {
	core := t.Core(itemIdx)
	for _, symbolRef := range t.grammar.Productions[core.ProductionIdx()].SymbolRefs[core.Position():] {
		children = append(children, NewLeafDerivation(symbolRef))
	}
	return children
}

// appendShift appends the symbols of the production of the item from its dot on, following the cheapest way the parser
// carries the terminal of the reach from the item to its shift. The parser must be able to. The symbols behind the
// terminal, or behind the nonterminal which begins with it, stay leaves, because nonterminals stay nonterminals where
// terminals are not germane (section 3.2).
func (t *LookupTables) appendShift(reach *terminalReach, children []Derivation, itemIdx int) []Derivation {
	utils.DebugAssert(func() error {
		if !reach.CanShift(itemIdx) {
			return fmt.Errorf("item %d cannot shift terminal %d", itemIdx, reach.terminalIdx)
		}
		return nil
	})
	for {
		if symbolRef, _ := t.NextSymbol(itemIdx); symbolRef.IsTerminal() {
			return t.appendLeaves(children, itemIdx)
		}
		targetItemIdx, _ := t.Transition(itemIdx)
		if closureItemIdx := int(reach.byItemIdx[itemIdx].toShift.closureItemIdx); closureItemIdx != noItemIdx {
			productionIdx := t.Core(closureItemIdx).ProductionIdx()
			children = append(children, NewExpandedDerivation(
				t.grammar,
				productionIdx,
				t.appendShift(reach, make([]Derivation, 0, len(t.grammar.Productions[productionIdx].SymbolRefs)), closureItemIdx),
			))
			return t.appendLeaves(children, targetItemIdx)
		}
		children = append(children, t.emptyDerivation(reach, itemIdx))
		itemIdx = targetItemIdx
	}
}

// appendEmptyRest appends the symbols of the production of the item from its dot on, each derived to the empty string
// along the cheapest way of the reach. The rest of the production must be able to vanish.
func (t *LookupTables) appendEmptyRest(reach *terminalReach, children []Derivation, itemIdx int) []Derivation {
	utils.DebugAssert(func() error {
		if !reach.CanReduce(itemIdx) {
			return fmt.Errorf("item %d cannot reduce on terminal %d", itemIdx, reach.terminalIdx)
		}
		return nil
	})
	for {
		if _, found := t.NextSymbol(itemIdx); !found {
			return children
		}
		children = append(children, t.emptyDerivation(reach, itemIdx))
		itemIdx, _ = t.Transition(itemIdx)
	}
}

// emptyDerivation returns the cheapest derivation of the empty string from the nonterminal after the dot of the item,
// whose reductions the declarations allow on the terminal of the reach. The nonterminal must be able to vanish.
func (t *LookupTables) emptyDerivation(reach *terminalReach, itemIdx int) Derivation {
	utils.DebugAssert(func() error {
		if !reach.CanVanish(itemIdx) {
			return fmt.Errorf("the nonterminal after the dot of item %d cannot vanish", itemIdx)
		}
		return nil
	})
	closureItemIdx := int(reach.byItemIdx[itemIdx].toVanish.closureItemIdx)
	productionIdx := t.Core(closureItemIdx).ProductionIdx()
	children := make([]Derivation, 0, len(t.grammar.Productions[productionIdx].SymbolRefs))
	return NewExpandedDerivation(t.grammar, productionIdx, t.appendEmptyRest(reach, children, closureItemIdx))
}

// shortestEmptyDerivation returns a shortest derivation of the empty string from the nonterminal, see EmptyDerivation.
// The nonterminal must be nullable.
func (t *LookupTables) shortestEmptyDerivation(nonterminalIdx int) Derivation {
	productionIdx, found := t.EmptyDerivation(nonterminalIdx)
	utils.DebugAssert(func() error {
		if !found {
			return fmt.Errorf("nonterminal %d cannot derive the empty string", nonterminalIdx)
		}
		return nil
	})

	symbolRefs := t.grammar.Productions[productionIdx].SymbolRefs
	children := make([]Derivation, 0, len(symbolRefs))
	for _, symbolRef := range symbolRefs {
		children = append(children, t.shortestEmptyDerivation(symbolRef.Idx()))
	}
	return NewExpandedDerivation(t.grammar, productionIdx, children)
}
