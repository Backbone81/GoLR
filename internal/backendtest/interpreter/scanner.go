package interpreter

import (
	"fmt"

	"github.com/backbone81/golr/internal/backendtest"
	"github.com/backbone81/golr/internal/scannergen/backend"
	"github.com/backbone81/golr/internal/scannergen/backend/table"
	"github.com/backbone81/golr/internal/scannergen/frontend"
)

// Match is what the scanner found for one attempt at a token: the rule which matched, and the extent of the input the
// scanner consumed for it.
//
// A match with a RuleIdx of table.NoRule is an attempt in which no rule matched. Start is then where the error is
// reported and End is where the scanner picks up again, which is one byte past the byte it could not consume. The bytes
// between the two are the error's lexeme in the trace: how far the scanner ran before it backed up is exactly what a
// driver can get wrong.
type Match struct {
	// RuleIdx is the index of the rule which matched, or table.NoRule when no rule matched.
	RuleIdx int

	// Start is the byte offset the match begins at.
	Start int

	// End is the byte offset one past the last byte of the match.
	End int
}

// Scanner is the reference scanner of the backend test harness. It reads the compressed tables the same way a generated
// table driven scanner reads them, and reports what it found one match at a time.
//
// The scan is a maximal munch. Every accepting state the automaton walks through is remembered as the longest match so
// far, and when the automaton runs into a state without a transition on the next byte, the scanner backs up to that
// remembered match. Two rules which match the same number of bytes are decided by rule order, which the DFA has already
// resolved by giving a state the lowest rule index of all rules it is part of.
type Scanner struct {
	dfa    table.CompressedDFA
	rules  []frontend.Rule
	source []byte

	// lineStarts holds the byte offset each line begins at, so Event can turn a match offset into a line and column.
	lineStarts []int

	// offset is where the next attempt at a token starts. It is the only state which survives a call to Next, so
	// that a scan is a function of the input alone.
	offset int
}

// NewScanner creates a scanner for the given DFA and input. It compresses the DFA itself, instead of taking the
// compressed tables, so that the rule names it reports and the tables it reads can never come from two different DFAs.
// The compression is lossless, which the exhaustive equivalence test of the table package proves for every state and
// every byte.
func NewScanner(dfa backend.DFA, source []byte) *Scanner {
	return &Scanner{
		dfa:        table.NewCompressedDFA(dfa),
		rules:      dfa.Rules,
		source:     source,
		lineStarts: newLineStarts(source),
	}
}

// Next scans the next match and reports whether there was one. It returns false once the whole input is consumed, which
// is the only way a scan ends: an input no rule matches produces a match with table.NoRule and the scan continues, so
// that a single bad byte does not hide everything behind it.
func (s *Scanner) Next() (Match, bool) {
	if len(s.source) <= s.offset {
		return Match{}, false
	}

	startIdx := s.offset
	ruleIdx := table.NoRule
	endIdx := startIdx

	stateIdx := 0
	peekIdx := startIdx
	for ; peekIdx < len(s.source); peekIdx++ {
		// The accepting state is checked before the byte is consumed, so the match which is remembered ends
		// where the automaton stands and not one byte further.
		if acceptRuleIdx := s.dfa.AcceptRuleIdxByStateIdx[stateIdx]; acceptRuleIdx != table.NoRule {
			ruleIdx = acceptRuleIdx
			endIdx = peekIdx
		}

		nextStateIdx := s.dfa.Transition(stateIdx, s.source[peekIdx])
		if nextStateIdx == table.NoTransition {
			// The state has no transition on this byte, so the token ends here and the scanner falls back
			// to the longest match it walked through.
			break
		}
		stateIdx = nextStateIdx
	}
	if peekIdx == len(s.source) {
		// The loop consumed the last byte of the input without ever looking at the state that byte led to.
		if acceptRuleIdx := s.dfa.AcceptRuleIdxByStateIdx[stateIdx]; acceptRuleIdx != table.NoRule {
			ruleIdx = acceptRuleIdx
			endIdx = peekIdx
		}
	}

	if startIdx < endIdx {
		s.offset = endIdx
		return Match{RuleIdx: ruleIdx, Start: startIdx, End: endIdx}, true
	}

	// No rule matched. The failed match is the bytes the automaton consumed, and the byte at peekIdx is not one of
	// them: that is the byte it could not consume, so it is where the next attempt starts. Consuming it here would
	// throw away the start of the token which follows.
	//
	// The automaton consumed nothing when it failed on the very first byte, and then that byte is the failed match
	// itself. Consuming at least one byte is what keeps the scan from standing still, and it is also what a rule
	// matching the empty string comes down to: such a rule accepts in the start state, leaves the match empty, and
	// therefore ends up here.
	//
	// No clamp is needed. This is only reached with startIdx below the length of the source, so startIdx+1 is at most
	// that length, and peekIdx never runs past it either.
	s.offset = max(startIdx+1, peekIdx)
	return Match{RuleIdx: table.NoRule, Start: startIdx, End: s.offset}, true
}

// rule returns the rule which produced the given match, and reports whether a rule produced it at all. A match which no
// rule produced is what a generated scanner hands on as its invalid token, which the parser then sees as a token which
// is no terminal of its grammar, see Parser.
func (s *Scanner) rule(match Match) (frontend.Rule, bool) {
	if match.RuleIdx == table.NoRule {
		return frontend.Rule{}, false
	}
	return s.rules[match.RuleIdx], true
}

// Event returns the trace event for the given match, positioned where the match starts. A match which no rule produced
// becomes an error carrying the bytes it could not match, every other match becomes a token, whether or not its rule
// was marked for skipping, see backendtest.Token.
func (s *Scanner) Event(match Match) fmt.Stringer {
	line, column := lineCol(s.lineStarts, match.Start)
	lexeme := string(s.source[match.Start:match.End])

	rule, ok := s.rule(match)
	if !ok {
		return backendtest.ScannerError{Line: line, Column: column, Lexeme: lexeme}
	}
	return backendtest.Token{Line: line, Column: column, RuleName: rule.Name, Lexeme: lexeme}
}

// ScanTrace scans the whole input and returns the canonical trace of it. Every trace ends with the end of input event,
// positioned one past the last byte, including the trace of an empty input, so that a trace always states where the
// input it belongs to ends.
func ScanTrace(dfa backend.DFA, source []byte) backendtest.Trace {
	scanner := NewScanner(dfa, source)

	var result backendtest.Trace
	for {
		match, ok := scanner.Next()
		if !ok {
			break
		}
		result = append(result, scanner.Event(match))
	}

	line, column := lineCol(scanner.lineStarts, len(source))
	return append(result, backendtest.EOF{Line: line, Column: column})
}
