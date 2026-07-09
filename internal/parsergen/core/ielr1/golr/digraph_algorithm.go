package golr

import (
	"math"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/utils"
)

// DigraphAlgorithm provides an implementation for algorithm Digraph as described by DeRemer and Pennello in
// "Efficient Computation of LALR(1) Look-Ahead Sets" at https://doi.org/10.1145/69622.357187. It provides functionality
// for propagating goto follow sets correctly across a directed graph which might contain loops and shortcuts.
type DigraphAlgorithm struct {
	// follows holds the follow set being propagated, indexed by goto index. The caller owns this slice and seeds it
	// with the direct contributions before calling Execute. The algorithm merges follow sets along the relation in
	// place, so after Execute the slice holds the fully propagated follow sets.
	follows []backend.LookaheadSet

	// successorGotoIdxs holds the target goto index of every edge in the relation, grouped by source goto, so that
	// traverse only looks at the outgoing edges of a goto instead of scanning the whole edge list on every call. The
	// successors of source goto g are the slice successorGotoIdxs[successorGotoIdxOffsets[g]:successorGotoIdxOffsets[g+1]].
	// This compressed-sparse-row layout keeps the whole relation in two allocations regardless of the number of gotos,
	// and stores the successors of each goto contiguously.
	successorGotoIdxs       []int
	successorGotoIdxOffsets []int

	// gotoIdxWorkStack holds the gotos on the current depth-first traversal path whose strongly connected component has
	// not been closed yet. It is the stack "S" of the Digraph algorithm. When a component root is found, its members sit
	// contiguously on top of the stack and are popped together. The zero value is an empty stack, ready to use.
	gotoIdxWorkStack utils.Stack[int]

	// processed records the traversal state of every goto, indexed by goto index. It is the array "N" of the Digraph
	// algorithm and encodes three states in one int:
	//   - 0:           the goto has not been visited yet (the zero value, set up by NewDigraphAlgorithm).
	//   - 1..len:      the goto is on gotoIdxWorkStack; the value is its depth on the stack, lowered to the smallest
	//                  depth reachable through a back edge so the component root is the goto whose value still equals its
	//                  own depth.
	//   - math.MaxInt: the goto's component has been closed; its follow set is final and it is ignored from then on.
	processed []int
}

// NewDigraphAlgorithm creates a new instances for algorithm digraph. The follows slice is indexed by goto index, owned
// by the caller, and propagated in place: seed it with the direct contributions before calling Execute and read the
// results back from it afterward.
func NewDigraphAlgorithm(
	follows []backend.LookaheadSet,
	edges []Edge,
) DigraphAlgorithm {
	// Build the compressed-sparse-row adjacency. First count the outgoing edges per goto to get the offsets, then place
	// each edge target into the slot reserved for its source goto.
	successorGotoIdxOffsets := make([]int, len(follows)+1)
	for _, edge := range edges {
		successorGotoIdxOffsets[edge.FromIdx+1]++
	}
	for gotoIdx := 1; gotoIdx < len(successorGotoIdxOffsets); gotoIdx++ {
		successorGotoIdxOffsets[gotoIdx] += successorGotoIdxOffsets[gotoIdx-1]
	}
	successorGotoIdxs := make([]int, len(edges))
	nextSuccessorSlot := make([]int, len(follows))
	copy(nextSuccessorSlot, successorGotoIdxOffsets)
	for _, edge := range edges {
		successorGotoIdxs[nextSuccessorSlot[edge.FromIdx]] = edge.ToIdx
		nextSuccessorSlot[edge.FromIdx]++
	}

	return DigraphAlgorithm{
		follows:                 follows,
		successorGotoIdxs:       successorGotoIdxs,
		successorGotoIdxOffsets: successorGotoIdxOffsets,
		processed:               make([]int, len(follows)),
	}
}

// Execute runs the algorithm digraph on all the gotos.
func (d *DigraphAlgorithm) Execute() {
	for gotoIdx := range len(d.follows) {
		if d.processed[gotoIdx] != 0 {
			// This goto index has already been processed and does not need any more processing.
			continue
		}
		d.traverse(gotoIdx)
	}
}

// traverse executes the algorithm digraph on the given goto and recurses for unprocessed gotos which are targeted by
// some edge.
func (d *DigraphAlgorithm) traverse(gotoIdx int) {
	d.gotoIdxWorkStack.Push(gotoIdx)
	currDepth := d.gotoIdxWorkStack.Size()
	d.processed[gotoIdx] = currDepth
	for _, toIdx := range d.successorGotoIdxs[d.successorGotoIdxOffsets[gotoIdx]:d.successorGotoIdxOffsets[gotoIdx+1]] {
		if d.processed[toIdx] == 0 {
			// The target goto index for the edge has not been processed yet, so process that goto now.
			d.traverse(toIdx)
		}
		d.processed[gotoIdx] = min(d.processed[gotoIdx], d.processed[toIdx])
		d.follows[gotoIdx].Merge(&d.follows[toIdx])
	}
	if d.processed[gotoIdx] == currDepth {
		for {
			topOfStack := d.gotoIdxWorkStack.Top()
			d.processed[topOfStack] = math.MaxInt
			// All members of a strongly connected component share the same follow set, which is fully accumulated in
			// the root of the component (gotoIdx). Copy it into each member, otherwise members other than the root keep
			// an incomplete set. This is the "F(Top of S) <- F x" step of the Digraph algorithm by DeRemer and Pennello.
			d.follows[topOfStack].Merge(&d.follows[gotoIdx])
			d.gotoIdxWorkStack.Pop()
			if topOfStack == gotoIdx {
				break
			}
		}
	}
}
