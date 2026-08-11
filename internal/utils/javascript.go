package utils

import (
	"math"
)

// JavaScriptUintArrayType returns the name of the narrowest unsigned JavaScript typed array constructor which can hold
// every value from zero up to the given one. It is the JavaScript counterpart of GoUintType.
//
// Picking the array from the values it actually holds is what keeps the generated tables small, and it is also what
// makes a generated JavaScript scanner or parser run into the same width boundaries as every other backend instead of
// hiding them behind the arbitrary precision of a plain JavaScript array.
func JavaScriptUintArrayType(maxValue int) string {
	switch {
	case maxValue <= math.MaxUint8:
		return "Uint8Array"
	case maxValue <= math.MaxUint16:
		return "Uint16Array"
	default:
		return "Uint32Array"
	}
}

// NewJavaScriptIntArray returns the given values as a table typed by the largest value it has to hold. It is the
// JavaScript counterpart of NewIntArray, which picks a Go type for the same values.
func NewJavaScriptIntArray(values []int) IntArray {
	var maxValue int
	for _, value := range values {
		maxValue = max(maxValue, value)
	}
	return NewTypedIntArray(JavaScriptUintArrayType(maxValue), values)
}
