package counterexample

import (
	"hash/fnv"
	"unsafe"
)

// ProductConfiguration is a configuration of the product parser, see figure 8: the two parsers it simulates, the first
// one reducing the reduce item and the second one shifting with the other item or reducing it.
type ProductConfiguration struct {
	Parsers [2]SimulatedParser

	// TerminalShifted reports if the parsers shifted the conflict terminal behind the dot.
	TerminalShifted bool

	// Cost is the sum of the costs of the actions which led to the ProductConfiguration.
	Cost int
}

// Successor returns a copy of the ProductConfiguration with the cost added. The sequences are immutable, so the copy
// shares them.
func (c *ProductConfiguration) Successor(cost int) *ProductConfiguration {
	result := *c
	result.Cost += cost
	return &result
}

// Hash calculates a hash over both parsers and if the conflict terminal was shifted, which tells configurations apart
// that the search treats as different, see simulatedParser.Hash.
func (c *ProductConfiguration) Hash() uint64 {
	values := [3]uint64{
		c.Parsers[0].Hash(),
		c.Parsers[1].Hash(),
		0,
	}
	if c.TerminalShifted {
		values[2] = 1
	}

	// We reinterpret the values as a slice of bytes. We do this with unsafe pointer arithmetic to avoid encoding the
	// values only for the hash.
	//nolint:gosec // unsafe is required for better performance
	valueBytes := unsafe.Slice((*byte)(unsafe.Pointer(&values)), unsafe.Sizeof(values))

	hash := fnv.New64a()
	if _, err := hash.Write(valueBytes); err != nil {
		panic(err)
	}
	return hash.Sum64()
}
