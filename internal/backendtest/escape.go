package backendtest

import (
	"fmt"
	"strings"
)

// printableLow and printableHigh delimit the range of bytes which are written to a trace as they are. Everything
// outside of it is escaped, which keeps a trace free of control characters and of anything a target language could
// interpret as an encoding of its own.
const (
	printableLow  = 0x20
	printableHigh = 0x7E
)

// EscapeLexeme escapes the bytes of a lexeme for a trace line. The caller writes the quotes around the result, so that
// the format string of an event shows the whole shape of the line.
//
// The escaping is deliberately minimal, so that no target language needs an escaping library to produce it: a
// backslash, a quote, the three whitespace characters which would otherwise break the one event per line rule, and
// \xHH for every byte outside of the printable ASCII range. Together with the quotes that is a minimal C style string
// literal, which every target language can write with a switch.
//
// Go's %q is not an alternative. On the raw lexeme it applies Go's own escaping, which writes a multi byte UTF-8
// sequence as the character it stands for instead of byte by byte, and which reaches for \a, \b, \f, \v and \u. On an
// already escaped lexeme it would escape the backslashes a second time.
//
// The input is treated as a sequence of bytes and not as a sequence of runes. A multi byte UTF-8 sequence therefore
// becomes one \xHH escape per byte, because bytes are the only unit every target language agrees on without
// conversion.
func EscapeLexeme(lexeme string) string {
	var result strings.Builder
	result.Grow(len(lexeme))

	for byteIdx := range len(lexeme) {
		value := lexeme[byteIdx]
		switch {
		case value == '\\':
			result.WriteString(`\\`)
		case value == '"':
			result.WriteString(`\"`)
		case value == '\n':
			result.WriteString(`\n`)
		case value == '\r':
			result.WriteString(`\r`)
		case value == '\t':
			result.WriteString(`\t`)
		case printableLow <= value && value <= printableHigh:
			result.WriteByte(value)
		default:
			// %02x is lower case and zero padded, which is exactly what the escape sequence asks for. The error can
			// only come from the writer, and a strings.Builder never fails.
			_, _ = fmt.Fprintf(&result, `\x%02x`, value)
		}
	}

	return result.String()
}
