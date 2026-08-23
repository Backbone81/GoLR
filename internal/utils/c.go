package utils

import (
	"math"
	"strings"
)

// CUintType returns the name of the narrowest C unsigned integer type which can hold every value from zero up to the
// given one. It is the C counterpart of GoUintType.
//
// The fixed width types of <stdint.h> are used rather than the built in ones, because the width of unsigned char, short
// and int is what the implementation says it is, while the size of a generated table has to be the same everywhere.
//
// Every one of them converts to size_t in an expression without a cast, so a lookup in a table typed this way composes
// with an index or with another lookup as it is.
func CUintType(maxValue int) string {
	switch {
	case maxValue <= math.MaxUint8:
		return "uint8_t"
	case maxValue <= math.MaxUint16:
		return "uint16_t"
	default:
		return "uint32_t"
	}
}

// NewCIntArray returns the given values as a table typed by the largest value it has to hold. It is the C counterpart
// of NewIntArray, which picks a Go type for the same values.
func NewCIntArray(values []int) IntArray {
	var maxValue int
	for _, value := range values {
		maxValue = max(maxValue, value)
	}
	return NewTypedIntArray(CUintType(maxValue), values)
}

// CTypeName returns the name a generated C type goes by, which is the prefix followed by the name the other backends
// give the same type. C has no namespaces, so the prefix is what keeps two generated parsers in one program apart.
//
// The spelling follows the C libraries a reader is most likely to know, where a type is written in camel case and a
// function in lower case with underscores.
func CTypeName(prefix string, name string) string {
	return GoIdentifier(prefix) + name
}

// CFunctionName returns the name a generated C function goes by, which is the prefix and the name in lower case with
// underscores. See CTypeName for why the prefix is part of the name.
func CFunctionName(prefix string, name string) string {
	return lowerSnakeName(GoIdentifier(prefix) + name)
}

// CConstantName returns the name a generated C constant or enumerator goes by, which is the prefix and the name in
// upper case with underscores. See CTypeName for why the prefix is part of the name.
//
// C spells a constant this way, which is the same reason the Java and the Python backends do, so the cross language
// rule that a token is called the same everywhere loses to the convention of the language here.
func CConstantName(prefix string, name string) string {
	return upperSnakeName(GoIdentifier(prefix) + name)
}

// lowerSnakeName returns the given identifier in lower case with underscores, so that TransitionBase becomes
// transition_base. It is upperSnakeName in the other case, and shares its notion of a word boundary.
func lowerSnakeName(identifier string) string {
	return strings.ToLower(upperSnakeName(identifier))
}
