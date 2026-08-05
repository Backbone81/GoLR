package utils

import (
	"math"
	"strings"
	"unicode"
)

// GoIdentifier creates a camel case name which is suitable as a Go identifier for functions or variables. Is used for
// code generation.
func GoIdentifier(text string) string {
	words := strings.FieldsFunc(text, func(r rune) bool {
		return r == '_' || r == ' ' || r == '\t'
	})

	var builder strings.Builder
	for _, word := range words {
		if len(word) == 0 {
			continue
		}

		cleaned := replaceSpecialCharacters(word)
		capitalized := capitalizeFirstChar(cleaned)
		builder.WriteString(capitalized)
	}
	return builder.String()
}

// GoUintType returns the name of the narrowest unsigned Go integer type which can hold every value from zero up to the
// given one. Is used for code generation, where picking the type of a lookup table from the values it actually holds is
// what keeps the generated tables small.
func GoUintType(maxValue int) string {
	switch {
	case maxValue <= math.MaxUint8:
		return "uint8"
	case maxValue <= math.MaxUint16:
		return "uint16"
	default:
		return "uint32"
	}
}

func replaceSpecialCharacters(text string) string {
	var builder strings.Builder
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
		} else {
			builder.WriteByte('_')
		}
	}
	return builder.String()
}

func capitalizeFirstChar(text string) string {
	var builder strings.Builder
	for i, r := range text {
		if i == 0 {
			builder.WriteRune(unicode.ToUpper(r))
		} else {
			builder.WriteRune(unicode.ToLower(r))
		}
	}
	return builder.String()
}
