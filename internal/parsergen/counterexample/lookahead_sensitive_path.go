package counterexample

import (
	"slices"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	"github.com/backbone81/golr/internal/utils"
)

// eofTerminalIdx is the terminal index of the end of the input, which the augmented grammar puts first.
const eofTerminalIdx = 0

// pathItem is an item on a path of transitions and production steps, together with the edge which entered it.
type pathItem struct {
	// itemIdx is the item.
	itemIdx int

	// productionStep reports if a production step entered the item, which starts a production of its own. The first
	// item of a path counts as entered by a production step.
	productionStep bool

	// terminalFollows reports if the precise lookahead set L of the vertex (s, itm, L) of section 4 holds the conflict
	// terminal. It is only known on a shortest lookahead-sensitive path, and false on the items a backward walk adds.
	terminalFollows bool
}

// lookaheadSensitiveVertex is a vertex (s, itm, L) of the lookahead-sensitive graph of section 4, held in two fields
// instead of three. Deviating from the paper, L is not kept as a set, only if it holds the conflict terminal, because
// that bit alone decides the edges of the vertex, see lookaheadSensitivePathBuilder.
type lookaheadSensitiveVertex struct {
	// itemIdx is the item, which names both s and itm.
	itemIdx int

	// terminalFollows reports if L holds the conflict terminal.
	terminalFollows bool
}

// lookaheadSensitivePathBuilder finds a shortest lookahead-sensitive path from the start item to a reduce item with the
// conflict terminal in its precise lookahead set, see section 4. It is a breadth first search over the
// lookahead-sensitive graph, whose vertices and edges are created as they are discovered (p. 4), and which only enters
// items that can reach the reduce item (section 6, "Finding shortest lookahead-sensitive path").
//
// Deviating from the paper, a vertex does not hold the precise lookahead set L, only if L holds the conflict terminal.
// followL of an item holds the terminal when the symbols behind the nonterminal can begin with it, or when they can
// vanish and L holds it, so the edges of a vertex and whether their targets hold the terminal only depend on the item
// and that bit. The search finds paths of the same length, but visits at most two vertices per item instead of one per
// item and distinct lookahead set, which are hundreds of thousands on large grammars.
//
// Deviating from the paper as well, the search takes the declarations of the grammar into account, see section 6,
// "Exploiting precedence". It only takes transitions on terminals whose shift the declarations leave to the parser,
// and L only holds the terminal when the parser can carry it to its shift, see terminalReach: the symbols behind the
// nonterminal begin with the terminal or vanish with every reduction allowed on it, and the production is reduced on
// it. So the reduction of the reduce item at the end of the path is followed by an allowed shift of the terminal.
type lookaheadSensitivePathBuilder struct {
	tables        *LookupTables
	reduceItemIdx int
	terminalIdx   int
	reach         *terminalReach

	reachingItemIdxs utils.Bitset

	// visitedItemIdxs holds the items of the vertices found so far, separately for both values of terminalFollows.
	visitedItemIdxs [2]utils.Bitset

	// The vertices in the order they were found, which is the queue of the search, with their parents and how they were
	// entered.
	vertices                  []lookaheadSensitiveVertex
	parentVertexIdxs          []int
	productionStepByVertexIdx []bool
}

// noVertexIdx stands for a missing vertex, like the parent of the first vertex of the search.
const noVertexIdx = -1

// newLookaheadSensitivePathBuilder returns a builder for the shortest lookahead-sensitive path to the reduce item with
// the terminal in its precise lookahead set.
func newLookaheadSensitivePathBuilder(
	tables *LookupTables,
	reduceItemIdx int,
	terminalIdx int,
) lookaheadSensitivePathBuilder {
	return lookaheadSensitivePathBuilder{
		tables:           tables,
		reduceItemIdx:    reduceItemIdx,
		terminalIdx:      terminalIdx,
		reach:            tables.terminalReach(terminalIdx),
		reachingItemIdxs: tables.ReachingItems(reduceItemIdx),
	}
}

// shortestLookaheadSensitivePath returns a shortest lookahead-sensitive path from (s0, START -> • S $, {$}) to the
// reduce item with the terminal in its precise lookahead set. It reports false when the declarations of the grammar
// leave the parser no input in which the reduction of the item is followed by the terminal.
func (t *LookupTables) shortestLookaheadSensitivePath(reduceItemIdx int, terminalIdx int) ([]pathItem, bool) {
	builder := newLookaheadSensitivePathBuilder(t, reduceItemIdx, terminalIdx)
	return builder.Build()
}

// Build returns the path, or reports false when there is none.
func (b *lookaheadSensitivePathBuilder) Build() ([]pathItem, bool) {
	startItemIdx, found := b.tables.ItemIdx(0, backend.NewCore(0, 0))
	if !found || !b.reachingItemIdxs.Contains(startItemIdx) {
		return nil, false
	}
	b.addVertex(lookaheadSensitiveVertex{
		itemIdx:         startItemIdx,
		terminalFollows: b.terminalIdx == eofTerminalIdx,
	}, noVertexIdx, true)

	//nolint:intrange // The loop bound grows while the loop runs, the vertices are the queue of the search.
	for vertexIdx := 0; vertexIdx < len(b.vertices); vertexIdx++ {
		vertex := b.vertices[vertexIdx]
		if vertex.itemIdx == b.reduceItemIdx && vertex.terminalFollows && b.reach.CanReduce(vertex.itemIdx) {
			return b.path(vertexIdx), true
		}
		b.addTransition(vertexIdx)
		b.addProductionSteps(vertexIdx)
	}
	return nil, false
}

// addTransition adds the vertex the transition of the vertex leads to, which keeps its lookahead set, see figure 4(a).
func (b *lookaheadSensitivePathBuilder) addTransition(vertexIdx int) {
	vertex := b.vertices[vertexIdx]
	targetItemIdx, found := b.tables.Transition(vertex.itemIdx)
	if !found || !b.reachingItemIdxs.Contains(targetItemIdx) {
		return
	}
	symbolRef, _ := b.tables.NextSymbol(vertex.itemIdx)
	stateIdx := b.tables.StateIdx(vertex.itemIdx)
	if symbolRef.IsTerminal() &&
		!b.tables.IsActionAllowed(stateIdx, symbolRef.Idx(), conflict.NewShiftContribution()) {
		return
	}
	b.addVertex(lookaheadSensitiveVertex{
		itemIdx:         targetItemIdx,
		terminalFollows: vertex.terminalFollows,
	}, vertexIdx, false)
}

// addProductionSteps adds the vertices the production steps of the vertex lead to, whose lookahead set is followL of
// its item, see figure 4(b).
func (b *lookaheadSensitivePathBuilder) addProductionSteps(vertexIdx int) {
	vertex := b.vertices[vertexIdx]
	from, to := b.tables.ProductionSteps(vertex.itemIdx)
	if from == to {
		return
	}
	terminalFollows := b.followLHoldsTerminal(vertex)
	for targetItemIdx := from; targetItemIdx < to; targetItemIdx++ {
		if b.reachingItemIdxs.Contains(targetItemIdx) {
			b.addVertex(lookaheadSensitiveVertex{
				itemIdx:         targetItemIdx,
				terminalFollows: terminalFollows,
			}, vertexIdx, true)
		}
	}
}

// followLHoldsTerminal reports if the precise follow set followL of the item of the vertex (p. 4) holds the terminal:
// when the parser can carry the terminal from behind the nonterminal after the dot to its shift, or when the symbols
// behind the nonterminal can vanish, the production is reduced on the terminal, and the lookahead set of the vertex
// holds it. These are the four cases of the paper unrolled over the rest of the production, restricted to the actions
// the declarations leave to the parser, see terminalReach.
func (b *lookaheadSensitivePathBuilder) followLHoldsTerminal(vertex lookaheadSensitiveVertex) bool {
	targetItemIdx, found := b.tables.Transition(vertex.itemIdx)
	if !found {
		return false
	}
	return b.reach.CanShift(targetItemIdx) || b.reach.CanReduce(targetItemIdx) && vertex.terminalFollows
}

// addVertex adds the vertex to the search, unless it was found before.
func (b *lookaheadSensitivePathBuilder) addVertex(
	vertex lookaheadSensitiveVertex,
	parentVertexIdx int,
	productionStep bool,
) {
	visitedIdx := 0
	if vertex.terminalFollows {
		visitedIdx = 1
	}
	if !b.visitedItemIdxs[visitedIdx].Add(vertex.itemIdx) {
		return
	}
	b.vertices = append(b.vertices, vertex)
	b.parentVertexIdxs = append(b.parentVertexIdxs, parentVertexIdx)
	b.productionStepByVertexIdx = append(b.productionStepByVertexIdx, productionStep)
}

// path returns the path from the start item to the vertex, following the parents of the vertices.
func (b *lookaheadSensitivePathBuilder) path(vertexIdx int) []pathItem {
	var result []pathItem
	for ; vertexIdx != noVertexIdx; vertexIdx = b.parentVertexIdxs[vertexIdx] {
		vertex := b.vertices[vertexIdx]
		result = append(result, pathItem{
			itemIdx:         vertex.itemIdx,
			productionStep:  b.productionStepByVertexIdx[vertexIdx],
			terminalFollows: vertex.terminalFollows,
		})
	}
	slices.Reverse(result)
	return result
}
