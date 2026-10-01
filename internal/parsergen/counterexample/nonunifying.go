package counterexample

import (
	"fmt"
	"slices"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
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
// The other item is an item which shifts the terminal, or a second reduce item.
func newNonunifyingBuilder(
	tables *LookupTables,
	reduceItemIdx int,
	otherItemIdx int,
	terminalIdx int,
) nonunifyingBuilder {
	return nonunifyingBuilder{
		tables:        tables,
		reduceItemIdx: reduceItemIdx,
		otherItemIdx:  otherItemIdx,
		terminalIdx:   terminalIdx,
	}
}

// Build returns the counterexample. Its derivations start at the start symbol, or at the start production when the
// conflict terminal is the end of the input, which only the start production shows.
func (b *nonunifyingBuilder) Build() Counterexample {
	reducePath := b.tables.shortestLookaheadSensitivePath(b.reduceItemIdx, b.terminalIdx)

	// The conflict terminal is the symbol after the dot of a shift item, so nothing is required behind it. A second
	// reduce item needs the terminal behind its production like the first one (p. 4, footnote 4).
	_, isShift := b.tables.NextSymbol(b.otherItemIdx)
	otherPath, found := b.walkBack(reducePath, walkNode{itemIdx: b.otherItemIdx, required: !isShift})
	if !found {
		// Only a second reduce item can miss the path, when the states the path goes through merged contexts in which
		// the terminal does not follow it. Its own shortest path shares less with the first one, but is valid.
		otherPath = b.tables.shortestLookaheadSensitivePath(b.otherItemIdx, b.terminalIdx)
	}

	derivations := [2]Derivation{
		b.derivation(reducePath, true),
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
	}
}

// walkBack explores backward from the node along the states of the path, until it meets an item of the path at the same
// point of the path, see figure 5(b). It returns the path up to that item, followed by the items the walk went through
// in forward direction, ending at the node. Points of the path are the runs of items between two transitions, which
// share a state. A later meeting point is preferred over an earlier one, because it shares more of the path.
//
// Within a state, the walk takes reverse production steps. A kernel item takes its reverse transition into the state of
// the previous point, which is unique, as that state holds the item with the dot one symbol earlier. While the terminal
// is required behind the production of the node, a reverse production step only goes to a parent whose symbols behind
// the nonterminal can begin with the terminal, which satisfies it, or can vanish, which keeps it required. A meeting
// item then has to hold the terminal in its precise lookahead set. The paper leaves this out, as it only walks back
// from a shift item, behind which nothing is required.
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

// reverseProductionStep returns the node of the parent item, and reports false when the parent is excluded because its
// symbols behind the nonterminal can neither begin with the required terminal nor vanish.
func (b *nonunifyingBuilder) reverseProductionStep(node walkNode, parentItemIdx int) (walkNode, bool) {
	parent := walkNode{itemIdx: parentItemIdx}
	if !node.required {
		return parent, true
	}
	var first backend.LookaheadSet
	nullable := b.tables.firstSets.FirstOfSequence(b.tables.restBehindNextSymbol(parentItemIdx), &first)
	if first.Contains(b.terminalIdx) {
		return parent, true
	}
	parent.required = true
	return parent, nullable
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
		core := b.tables.Core(path[productionTo-1].itemIdx)
		symbolRefs := b.tables.grammar.Productions[core.ProductionIdx()].SymbolRefs
		children := make([]Derivation, 0, len(symbolRefs)+1)
		for _, symbolRef := range symbolRefs[:core.Position()] {
			children = append(children, NewLeafDerivation(symbolRef))
		}

		rest := symbolRefs[core.Position():]
		if productionTo == len(path) {
			children = append(children, NewDotDerivation())
		} else {
			children = append(children, result)
			rest = rest[1:]
		}
		children, required = b.appendRest(children, rest, required)
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

// appendRest appends the symbols behind the dot or the expanded nonterminal of a production. While the terminal is
// required, the symbols are examined in order: the terminal satisfies it, a nonterminal which can begin with it is
// expanded into a shortest derivation which does, and a nonterminal which cannot is derived to the empty string, so the
// terminal directly follows the dot in the example. Once it is satisfied, the symbols stay leaves. It returns if the
// terminal is still required, which passes the requirement on to the enclosing production.
func (b *nonunifyingBuilder) appendRest(
	children []Derivation,
	rest []frontend.SymbolRef,
	required bool,
) ([]Derivation, bool) {
	for _, symbolRef := range rest {
		switch {
		case !required || symbolRef.IsTerminal():
			children = append(children, NewLeafDerivation(symbolRef))
			if symbolRef == frontend.NewTerminalRef(b.terminalIdx) {
				required = false
			}
		case b.tables.firstSets.IsNullable(symbolRef.Idx()) && !b.canBeginWithTerminal(symbolRef.Idx()):
			children = append(children, b.tables.emptyDerivation(symbolRef.Idx()))
		default:
			children = append(children, b.tables.derivationStartingWith(symbolRef.Idx(), b.terminalIdx))
			required = false
		}
	}
	return children, required
}

// canBeginWithTerminal reports if a derivation of the nonterminal can begin with the terminal.
func (b *nonunifyingBuilder) canBeginWithTerminal(nonterminalIdx int) bool {
	_, _, found := b.tables.DerivationStartingWith(nonterminalIdx, b.terminalIdx)
	return found
}

// derivationStartingWith returns a shortest derivation of the nonterminal which begins with the terminal, see
// DerivationStartingWith. A derivation must exist.
func (t *LookupTables) derivationStartingWith(nonterminalIdx int, terminalIdx int) Derivation {
	productionIdx, position, found := t.DerivationStartingWith(nonterminalIdx, terminalIdx)
	utils.DebugAssert(func() error {
		if !found {
			return fmt.Errorf("nonterminal %d cannot begin with terminal %d", nonterminalIdx, terminalIdx)
		}
		return nil
	})

	symbolRefs := t.grammar.Productions[productionIdx].SymbolRefs
	children := make([]Derivation, 0, len(symbolRefs))
	for _, symbolRef := range symbolRefs[:position] {
		children = append(children, t.emptyDerivation(symbolRef.Idx()))
	}
	if symbolRef := symbolRefs[position]; symbolRef.IsTerminal() {
		children = append(children, NewLeafDerivation(symbolRef))
	} else {
		children = append(children, t.derivationStartingWith(symbolRef.Idx(), terminalIdx))
	}
	for _, symbolRef := range symbolRefs[position+1:] {
		children = append(children, NewLeafDerivation(symbolRef))
	}
	return NewExpandedDerivation(t.grammar, productionIdx, children)
}

// emptyDerivation returns a shortest derivation of the empty string from the nonterminal, see EmptyDerivation. The
// nonterminal must be nullable.
func (t *LookupTables) emptyDerivation(nonterminalIdx int) Derivation {
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
		children = append(children, t.emptyDerivation(symbolRef.Idx()))
	}
	return NewExpandedDerivation(t.grammar, productionIdx, children)
}
