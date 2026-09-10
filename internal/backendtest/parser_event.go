package backendtest

import (
	"fmt"
	"strings"
)

// traceLine formats one trace line, scanner or parser: the source position, the keyword, and an optional payload. No
// trailing whitespace.
func traceLine(line int, column int, keyword string, payload string) string {
	location := fmt.Sprintf("%d:%d", line, column)
	if payload == "" {
		return fmt.Sprintf("%-7s %s", location, keyword)
	}
	return fmt.Sprintf("%-7s %-7s %s", location, keyword, payload)
}

// Shift reports a terminal, including the end of input symbol, being shifted onto the parse stack.
type Shift struct {
	Line         int
	Column       int
	TerminalName string
	Lexeme       string
}

// String returns the canonical trace line for the event, without the terminating newline.
func (s Shift) String() string {
	return traceLine(s.Line, s.Column, "SHIFT", s.TerminalName+` "`+EscapeLexeme(s.Lexeme)+`"`)
}

// Reduce reports a production being reduced, as "lhs => rhs" or "lhs => ε" for an empty right hand side.
type Reduce struct {
	Line          int
	Column        int
	LeftHandSide  string
	RightHandSide []string
}

// String returns the canonical trace line for the event, without the terminating newline.
func (r Reduce) String() string {
	payload := r.LeftHandSide + " =>"
	if len(r.RightHandSide) == 0 {
		payload += " ε"
	}
	var payloadSb44 strings.Builder
	for _, name := range r.RightHandSide {
		payloadSb44.WriteString(" " + name)
	}
	payload += payloadSb44.String()
	return traceLine(r.Line, r.Column, "REDUCE", payload)
}

// ParserError reports a syntax error at the current lookahead. Suppressed marks an error the parser hits while still
// recovering and does not report to the caller.
type ParserError struct {
	Line       int
	Column     int
	Detail     string
	Suppressed bool
}

// String returns the canonical trace line for the event, without the terminating newline.
func (p ParserError) String() string {
	detail := p.Detail
	if p.Suppressed {
		detail = "(suppressed) " + detail
	}
	return traceLine(p.Line, p.Column, "ERROR", detail)
}

// Discard reports error recovery dropping the lookahead it keeps failing on.
type Discard struct {
	Line         int
	Column       int
	TerminalName string
	Lexeme       string
}

// String returns the canonical trace line for the event, without the terminating newline.
func (d Discard) String() string {
	return traceLine(d.Line, d.Column, "DISCARD", d.TerminalName+` "`+EscapeLexeme(d.Lexeme)+`"`)
}

// Pop reports error recovery dropping one state and the symbol it had parsed.
type Pop struct {
	Line       int
	Column     int
	SymbolName string
}

// String returns the canonical trace line for the event, without the terminating newline.
func (p Pop) String() string {
	return traceLine(p.Line, p.Column, "POP", p.SymbolName)
}

// Resync reports error recovery shifting the error symbol and resuming.
type Resync struct {
	Line   int
	Column int
}

// String returns the canonical trace line for the event, without the terminating newline.
func (r Resync) String() string {
	return traceLine(r.Line, r.Column, "RESYNC", "")
}

// Fail reports the parse being given up. A parse ends on either Accept or Fail.
type Fail struct {
	Line   int
	Column int
}

// String returns the canonical trace line for the event, without the terminating newline.
func (f Fail) String() string {
	return traceLine(f.Line, f.Column, "FAIL", "")
}

// Accept reports that the parser accepted the input.
type Accept struct {
	Line   int
	Column int
}

// String returns the canonical trace line for the event, without the terminating newline.
func (a Accept) String() string {
	return traceLine(a.Line, a.Column, "ACCEPT", "")
}
