package backendtest

// Token reports a rule which matched, at the position its lexeme starts.
type Token struct {
	Line     int
	Column   int
	RuleName string
	Lexeme   string
}

// String returns the canonical trace line for the event, without the terminating newline.
func (t Token) String() string {
	return traceLine(t.Line, t.Column, "TOKEN", t.RuleName+` "`+EscapeLexeme(t.Lexeme)+`"`)
}

// ScannerError reports a stretch of bytes no rule matched, at the position it starts. Its lexeme is those bytes: how
// far the scanner ran before it backed up to the last accepting state is exactly where a driver can differ from the
// reference.
type ScannerError struct {
	Line   int
	Column int
	Lexeme string
}

// String returns the canonical trace line for the event, without the terminating newline.
func (s ScannerError) String() string {
	return traceLine(s.Line, s.Column, "ERROR", `"`+EscapeLexeme(s.Lexeme)+`"`)
}

// EOF reports that the scanner consumed the whole input. Its position is one past the last byte, so a trace states
// where the input it belongs to ends.
type EOF struct {
	Line   int
	Column int
}

// String returns the canonical trace line for the event, without the terminating newline.
func (e EOF) String() string {
	return traceLine(e.Line, e.Column, "EOF", "")
}
