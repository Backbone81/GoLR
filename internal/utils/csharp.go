//nolint:dupl
package utils

import (
	"fmt"
	"math"
	"strings"
)

// CSharpIntType returns the name of the narrowest C# integer type which can hold every value from zero up to the given
// one. It is the C# counterpart of GoUintType.
//
// The ladder stops at int instead of continuing to uint, because byte, ushort and int all widen to int in an
// expression. A lookup in a table typed this way therefore composes with an index or with another lookup without a
// cast, which uint would need.
func CSharpIntType(maxValue int) string {
	switch {
	case maxValue <= math.MaxUint8:
		return "byte"
	case maxValue <= math.MaxUint16:
		return "ushort"
	default:
		return "int"
	}
}

// NewCSharpIntArray returns the given values as a table typed by the largest value it has to hold. It is the C#
// counterpart of NewIntArray, which picks a Go type for the same values.
func NewCSharpIntArray(values []int) IntArray {
	var maxValue int
	for _, value := range values {
		maxValue = max(maxValue, value)
	}
	return NewTypedIntArray(CSharpIntType(maxValue), values)
}

// CSharpStringLiteral returns the given string as a C# string literal. Control characters are written as \u escapes,
// which unlike a \x escape have a fixed length and cannot run into a hex digit following them. Every other byte is
// written as it is, so UTF-8 stays readable.
func CSharpStringLiteral(value string) string {
	var builder strings.Builder
	builder.WriteByte('"')
	for i := range len(value) {
		char := value[i]
		switch {
		case char == '\\':
			builder.WriteString(`\\`)
		case char == '"':
			builder.WriteString(`\"`)
		case char == '\n':
			builder.WriteString(`\n`)
		case char == '\r':
			builder.WriteString(`\r`)
		case char == '\t':
			builder.WriteString(`\t`)
		case char < 0x20 || char == 0x7f:
			fmt.Fprintf(&builder, `\u%04x`, char)
		default:
			builder.WriteByte(char)
		}
	}
	builder.WriteByte('"')
	return builder.String()
}
