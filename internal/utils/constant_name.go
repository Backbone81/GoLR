package utils

import (
	"strings"
	"unicode"
)

// upperSnakeName returns the given identifier in upper case with underscores, so that TransitionBase becomes
// TRANSITION_BASE. A word boundary is a lower case letter or a digit followed by an upper case one, which is what keeps
// an identifier which is already one word from gaining an underscore.
//
// It is shared because several languages spell a constant this way. Each of them still has an entry point of its own,
// so a language which stops agreeing changes its own function instead of everyone else's.
func upperSnakeName(identifier string) string {
	var builder strings.Builder
	for i, r := range identifier {
		if i > 0 && unicode.IsUpper(r) && !unicode.IsUpper(rune(identifier[i-1])) {
			builder.WriteByte('_')
		}
		builder.WriteRune(unicode.ToUpper(r))
	}
	return builder.String()
}
