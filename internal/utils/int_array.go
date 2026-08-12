package utils

import (
	"fmt"
	"strings"
)

// valuesPerLine is the number of table entries written per line of generated source code. Tables can hold thousands of
// entries, so they are wrapped to keep the generated file readable.
const valuesPerLine = 16

// IntArray is a table of non-negative integers together with the narrowest integer type which can hold all of them. Is
// used for code generation, where picking the type per grammar instead of using a single wide type is what keeps the
// generated tables small, because the entries of a small grammar fit into a single byte each.
type IntArray struct {
	// Type is the name of the type of an entry.
	Type string

	// Values are the entries of the table.
	Values []int
}

// NewIntArray returns the given values as a table typed by the largest value it has to hold.
func NewIntArray(values []int) IntArray {
	var maxValue int
	for _, value := range values {
		maxValue = max(maxValue, value)
	}
	return IntArray{
		Type:   GoUintType(maxValue),
		Values: values,
	}
}

// NewTypedIntArray returns the given values as a table of the given type. Use this where two tables are compared
// against each other at runtime and therefore have to share a type, instead of each getting the narrowest type its own
// values would allow.
func NewTypedIntArray(typeName string, values []int) IntArray {
	return IntArray{
		Type:   typeName,
		Values: values,
	}
}

// Literal returns the entries of the table as the body of a composite literal, wrapped over several lines and with
// every line starting with the given indentation.
// No line ends in whitespace, for the same reason: an editor or a commit hook which strips trailing whitespace must
// not be able to change a generated file.
func (a IntArray) Literal(indent string) string {
	var builder strings.Builder
	for i, value := range a.Values {
		if i%valuesPerLine == 0 {
			builder.WriteString("\n")
			builder.WriteString(indent)
		} else {
			builder.WriteString(" ")
		}
		fmt.Fprintf(&builder, "%d,", value)
	}
	builder.WriteString("\n")
	return builder.String()
}
