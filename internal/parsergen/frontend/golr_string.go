package frontend

import (
	"fmt"
	"strings"
)

// QuoteGoLRString returns the value as string literal of a GoLR grammar file. A string literal escapes double quotes
// and backslashes with a backslash and holds every other character as is. It returns an error for a value with a line
// feed, which a string literal can not hold.
func QuoteGoLRString(value string) (string, error) {
	if strings.ContainsRune(value, '\n') {
		return "", fmt.Errorf("the string %q has a line feed, which a GoLR string literal can not hold", value)
	}
	var builder strings.Builder
	builder.WriteByte('"')
	for _, r := range value {
		if r == '"' || r == '\\' {
			builder.WriteByte('\\')
		}
		builder.WriteRune(r)
	}
	builder.WriteByte('"')
	return builder.String(), nil
}

// UnquoteGoLRString returns the value of a string literal of a GoLR grammar file, which includes its double quotes. A
// backslash takes the character after it literally, which are only \" and \\ in what the GoLR scanner accepts.
func UnquoteGoLRString(literal string) string {
	inner := literal[1 : len(literal)-1]

	var builder strings.Builder
	for idx := 0; idx < len(inner); idx++ {
		if inner[idx] == '\\' && idx+1 < len(inner) {
			idx++
		}
		builder.WriteByte(inner[idx])
	}
	return builder.String()
}
