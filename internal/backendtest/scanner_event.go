package backendtest

import (
	"fmt"
)

// Token reports a rule which matched and produced a terminal for the parser.
type Token struct {
	RuleName string
	Start    int
	End      int
	Lexeme   string
}

// String returns the canonical trace line for the event, without the terminating newline.
func (t Token) String() string {
	return fmt.Sprintf(`TOKEN %s %d %d "%s"`, t.RuleName, t.Start, t.End, EscapeLexeme(t.Lexeme))
}

// Skip reports a rule which matched but produces no terminal, like whitespace or a comment.
type Skip struct {
	RuleName string
	Start    int
	End      int
	Lexeme   string
}

// String returns the canonical trace line for the event, without the terminating newline.
func (s Skip) String() string {
	return fmt.Sprintf(`SKIP %s %d %d "%s"`, s.RuleName, s.Start, s.End, EscapeLexeme(s.Lexeme))
}

// ScannerError reports the byte offset at which no rule matched. The offset is the finding, not an incidental detail,
// because backing up to the last accepting state is exactly where a driver can differ from the reference.
type ScannerError struct {
	Offset int
}

// String returns the canonical trace line for the event, without the terminating newline.
func (s ScannerError) String() string {
	return fmt.Sprintf("ERROR %d", s.Offset)
}

// EOF reports that the scanner consumed the whole input. It carries the offset so that a trace states the length of the
// input it belongs to.
type EOF struct {
	Offset int
}

// String returns the canonical trace line for the event, without the terminating newline.
func (e EOF) String() string {
	return fmt.Sprintf("EOF %d", e.Offset)
}
