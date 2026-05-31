package ielr1go

import (
	"math"

	"github.com/backbone81/golr/internal/utils"
)

// DigraphAlgorithm provides an implementation for algorithm Digraph as described by DeRemer and Pennello in
// "Efficient Computation of LALR(1) Look-Ahead Sets" at https://doi.org/10.1145/69622.357187. It provides functionality
// for propagating goto follow sets correctly across a directed graph which might contain loops and shortcuts.
type DigraphAlgorithm struct {
	gotoRecords []GotoRecord
	edges       []Edge

	gotoIdxWorkStack utils.Stack[int]
	processed        []int

	// TODO: We should not work with a merge function
	merge func(fromGotoIdx int, toGotoIdx int)
}

// NewDigraphAlgorithm creates a new instances for algorithm digraph.
func NewDigraphAlgorithm(
	gotoRecords []GotoRecord,
	edges []Edge,
	merge func(fromGotoIdx int, toGotoIdx int),
) DigraphAlgorithm {
	return DigraphAlgorithm{
		gotoRecords: gotoRecords,
		edges:       edges,
		merge:       merge,
		processed:   make([]int, len(gotoRecords)),
	}
}

// Execute runs the algorithm digraph on all the gotos.
func (d *DigraphAlgorithm) Execute() {
	for gotoIdx := range len(d.gotoRecords) {
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
	// TODO: iterating over all edges here is very inefficient
	for _, edge := range d.edges {
		if edge.FromIdx != gotoIdx {
			continue
		}
		if d.processed[edge.ToIdx] == 0 {
			// The target goto index for the edge has not been processed yet, so process that goto now.
			d.traverse(edge.ToIdx)
		}
		d.processed[gotoIdx] = min(d.processed[gotoIdx], d.processed[edge.ToIdx])
		d.merge(gotoIdx, edge.ToIdx)
	}
	if d.processed[gotoIdx] == currDepth {
		for {
			topOfStack := d.gotoIdxWorkStack.Top()
			d.processed[topOfStack] = math.MaxInt
			d.gotoIdxWorkStack.Pop()
			if topOfStack == gotoIdx {
				break
			}
		}
	}
}
