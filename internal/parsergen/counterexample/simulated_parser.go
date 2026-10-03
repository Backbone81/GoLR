package counterexample

import (
	"hash/fnv"
	"unsafe"

	"github.com/backbone81/golr/internal/utils"
)

// SimulatedParser is one of the two parsers of a ProductConfiguration: a sequence of items, linked by transitions and
// production steps, and the partial derivations of its transitions, one per transition in order.
type SimulatedParser struct {
	items       utils.Deque[int]
	derivations utils.Deque[Derivation]

	// depth is how deep the last item is nested in productions entered behind the conflict item, or completedDepth
	// once the production of the conflict item is reduced. That completes stage 1 for the first parser and stage 2 for
	// the second one (p. 5).
	depth int
}

// completedDepth is the depth of a parser which reduced the production of its conflict item.
const completedDepth = -1

// NewSimulatedParser returns a parser which holds nothing but the conflict item.
func NewSimulatedParser(conflictItemIdx int) SimulatedParser {
	return SimulatedParser{items: utils.Deque[int]{}.PushBack(conflictItemIdx)}
}

// Head returns the first item.
func (p *SimulatedParser) Head() int {
	return p.items.First()
}

// Tail returns the last item.
func (p *SimulatedParser) Tail() int {
	return p.items.Last()
}

// StageCompleted reports if the parser reduced the production of its conflict item.
func (p *SimulatedParser) StageCompleted() bool {
	return p.depth == completedDepth
}

// Hash calculates a hash over the items and the depth. The derivations are not part of it, so of two configurations
// which differ only in their derivations the search continues the one queued first. The items are collected in the
// buffer, which is returned for reuse.
func (p *SimulatedParser) Hash(buffer []int) (uint64, []int) {
	var values [2]uint64
	values[0], buffer = p.items.Hash(buffer)
	values[1] = uint64(p.depth) //nolint:gosec // The bits of the depth are hashed.

	// We reinterpret the values as a slice of bytes. We do this with unsafe pointer arithmetic to avoid encoding the
	// values only for the hash.
	//nolint:gosec // unsafe is required for better performance
	valueBytes := unsafe.Slice((*byte)(unsafe.Pointer(&values)), unsafe.Sizeof(values))

	hash := fnv.New64a()
	if _, err := hash.Write(valueBytes); err != nil {
		panic(err)
	}
	return hash.Sum64(), buffer
}
