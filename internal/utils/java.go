package utils

import (
	"math"
	"strings"
	"unicode"
)

// javaValuesPerMethod is the number of table entries one generated method holds.
//
// The Java virtual machine limits a method to 64 KB of bytecode, and an array literal costs about eight bytes per
// entry, which is why a table is split over several methods instead of being written as one literal. Measured against
// the compiler, a method takes 8000 entries and fails at 9000, so this leaves room for the wider entry types and for
// the entries which cost more than the average.
//
// The limit applies per method and not per literal, so moving the entries out of the field initializer is what buys
// the room: every static field of a class is initialized in the one class initializer the compiler writes, and a
// grammar of any size would otherwise have to fit into that single method.
const javaValuesPerMethod = 4000

// JavaIntType returns the name of the narrowest Java integer type which can hold every value from zero up to the given
// one. It is the Java counterpart of GoUintType.
//
// Java has no unsigned integer types, so the ladder is bounded by the positive range of each signed type rather than
// by its width. All three widen to int in an expression, so a lookup in a table typed this way composes with an index
// or with another lookup without a cast.
func JavaIntType(maxValue int) string {
	switch {
	case maxValue <= math.MaxInt8:
		return "byte"
	case maxValue <= math.MaxInt16:
		return "short"
	default:
		return "int"
	}
}

// NewJavaIntArray returns the given values as a table typed by the largest value it has to hold. It is the Java
// counterpart of NewIntArray, which picks a Go type for the same values.
func NewJavaIntArray(values []int) IntArray {
	var maxValue int
	for _, value := range values {
		maxValue = max(maxValue, value)
	}
	return NewTypedIntArray(JavaIntType(maxValue), values)
}

// JavaConstantName returns the given identifier in the upper case with underscores which Java spells a constant with,
// so that TransitionBase becomes TRANSITION_BASE. A word boundary is a lower case letter or a digit followed by an
// upper case one, which is what keeps an identifier which is already one word from gaining an underscore.
func JavaConstantName(identifier string) string {
	var builder strings.Builder
	for i, r := range identifier {
		if i > 0 && unicode.IsUpper(r) && !unicode.IsUpper(rune(identifier[i-1])) {
			builder.WriteByte('_')
		}
		builder.WriteRune(unicode.ToUpper(r))
	}
	return builder.String()
}

// JavaNameTable is a lookup table whose entries are the names of constants rather than numbers. It is chunked the same
// way a JavaTable is, because the entries cost the same bytecode either way.
type JavaNameTable struct {
	// Name is the name of the constant holding the whole table.
	Name string

	// Method is the name of the methods holding the entries. Each one carries the index of its chunk as a suffix.
	Method string

	// Type is the Java type of one entry.
	Type string

	// Chunks are the names of the entries, split into pieces small enough for one method each.
	Chunks [][]string
}

// NewJavaNameTable returns the given entries under the given method name, split into as many chunks as they need.
func NewJavaNameTable(method string, typeName string, names []string) JavaNameTable {
	var chunks [][]string
	for start := 0; start < len(names); start += javaValuesPerMethod {
		chunks = append(chunks, names[start:min(start+javaValuesPerMethod, len(names))])
	}
	if len(chunks) == 0 {
		// A table with no entries at all still needs the one method the constant is built from.
		chunks = append(chunks, nil)
	}

	return JavaNameTable{
		Name:   JavaConstantName(method),
		Method: method,
		Type:   typeName,
		Chunks: chunks,
	}
}

// JavaTable is a lookup table in the form the generated Java code holds it: a constant built from the entries of
// several methods, because no single method can hold them all.
type JavaTable struct {
	// Name is the name of the constant holding the whole table.
	Name string

	// Method is the name of the methods holding the entries. Each one carries the index of its chunk as a suffix.
	Method string

	// Type is the Java type of one entry.
	Type string

	// Chunks are the entries, split into pieces small enough for one method each.
	Chunks []IntArray
}

// NewJavaTable returns the given table under the given method name, split into as many chunks as it needs. The name of
// the constant follows from the method name, so that the two can never drift apart.
func NewJavaTable(method string, array IntArray) JavaTable {
	var chunks []IntArray
	for start := 0; start < len(array.Values); start += javaValuesPerMethod {
		end := min(start+javaValuesPerMethod, len(array.Values))
		chunks = append(chunks, NewTypedIntArray(array.Type, array.Values[start:end]))
	}
	if len(chunks) == 0 {
		// A table with no entries at all still needs the one method the constant is built from.
		chunks = append(chunks, NewTypedIntArray(array.Type, nil))
	}

	return JavaTable{
		Name:   JavaConstantName(method),
		Method: method,
		Type:   array.Type,
		Chunks: chunks,
	}
}
