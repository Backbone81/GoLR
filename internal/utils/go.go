package utils

import (
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
