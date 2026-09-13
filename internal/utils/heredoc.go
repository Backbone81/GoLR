package utils

import "strings"

// HereDoc cleans up an indented raw string literal so it can be written aligned with the surrounding Go code
// instead of flush against the left margin. It drops a leading blank line and then removes the longest leading
// whitespace prefix shared by every remaining non-blank line, leaving the relative indentation of the content
// intact.
func HereDoc(s string) string {
	s = strings.TrimPrefix(s, "\n")
	lines := strings.Split(s, "\n")

	indent := commonIndent(lines)
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			// Force blank lines to be fully empty instead of relying on TrimPrefix below: a blank line may be
			// shorter than indent (e.g. an editor stripped its trailing whitespace), in which case TrimPrefix
			// would leave its leftover whitespace in place rather than stripping it.
			lines[i] = ""
			continue
		}
		lines[i] = strings.TrimPrefix(line, indent)
	}
	return strings.Join(lines, "\n")
}

// commonIndent returns the longest leading whitespace prefix shared by every non-blank line.
func commonIndent(lines []string) string {
	var indent string
	found := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			// Blank lines carry no indentation information and must not narrow the common prefix: a short or
			// empty blank line would otherwise collapse indent to less than what the actual content agrees on.
			continue
		}
		lineIndent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		if !found {
			// Seed indent with the first non-blank line's own indentation. Without this, indent would start as
			// "" and commonPrefix("", lineIndent) is always "", so no line could ever establish an initial value.
			indent = lineIndent
			found = true
			continue
		}
		indent = commonPrefix(indent, lineIndent)
	}
	return indent
}

// commonPrefix returns the longest prefix shared by a and b.
func commonPrefix(a, b string) string {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}
	return a[:i]
}
