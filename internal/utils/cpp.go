package utils

import (
	"math"
)

// CppUintType returns the name of the narrowest C++ unsigned integer type which can hold every value from zero up to
// the given one. It is the C++ counterpart of GoUintType.
//
// The fixed width types of <cstdint> are used rather than the built in ones, because the width of unsigned char, short
// and int is what the implementation says it is, while the size of a generated table has to be the same everywhere.
//
// Every one of them widens to std::size_t in an expression without a cast, so a lookup in a table typed this way
// composes with an index or with another lookup as it is.
func CppUintType(maxValue int) string {
	switch {
	case maxValue <= math.MaxUint8:
		return "std::uint8_t"
	case maxValue <= math.MaxUint16:
		return "std::uint16_t"
	default:
		return "std::uint32_t"
	}
}

// NewCppIntArray returns the given values as a table typed by the largest value it has to hold. It is the C++
// counterpart of NewIntArray, which picks a Go type for the same values.
func NewCppIntArray(values []int) IntArray {
	var maxValue int
	for _, value := range values {
		maxValue = max(maxValue, value)
	}
	return NewTypedIntArray(CppUintType(maxValue), values)
}
