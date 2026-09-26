//nolint:dupl
package utils

import (
	"fmt"
	"math"
	"strings"
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

// RustStringLiteral returns the given string as a Rust string literal. Control characters are written as \x escapes,
// which always take two digits and cannot run into a hex digit following them. Every other byte is written as it is, so
// UTF-8 stays readable.
func RustStringLiteral(value string) string {
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
			fmt.Fprintf(&builder, `\x%02x`, char)
		default:
			builder.WriteByte(char)
		}
	}
	builder.WriteByte('"')
	return builder.String()
}
