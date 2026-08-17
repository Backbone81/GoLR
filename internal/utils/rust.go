package utils

import (
	"math"
)

// RustUintType returns the name of the narrowest Rust unsigned integer type which can hold every value from zero up to
// the given one. It is the Rust counterpart of GoUintType.
//
// Rust never widens an integer on its own, so a lookup in a table typed this way is cast to usize where it is used as
// an index or added to another lookup. Picking the type from the values the table actually holds is therefore free: it
// keeps the table small without costing a conversion the wider type would not also cost.
func RustUintType(maxValue int) string {
	switch {
	case maxValue <= math.MaxUint8:
		return "u8"
	case maxValue <= math.MaxUint16:
		return "u16"
	default:
		return "u32"
	}
}

// NewRustIntArray returns the given values as a table typed by the largest value it has to hold. It is the Rust
// counterpart of NewIntArray, which picks a Go type for the same values.
func NewRustIntArray(values []int) IntArray {
	var maxValue int
	for _, value := range values {
		maxValue = max(maxValue, value)
	}
	return NewTypedIntArray(RustUintType(maxValue), values)
}
