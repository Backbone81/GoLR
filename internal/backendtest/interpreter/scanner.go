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
// A match with a RuleIdx of table.NoRule is an attempt in which no rule matched. Start is then the offset the trace
// reports the error at, and End is where the scanner picks up again, which is one byte past the byte it could not
// consume. The extent of a failed attempt is deliberately part of the match instead of being implicit, because it is
// what decides the offsets of everything the scanner reports after an error.
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
		dfa:    table.NewCompressedDFA(dfa),
		rules:  dfa.Rules,
		source: source,
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

	// No rule matched. The scanner consumes everything it looked at, including the byte it could not consume, and
	// tries again after it. Consuming at least one byte is what keeps the scan from standing still, and it is also
	// what a rule matching the empty string comes down to: such a rule accepts in the start state, leaves the match
	// empty, and therefore ends up here.
	s.offset = min(peekIdx+1, len(s.source))
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

// Event returns the trace event for the given match. A match which no rule produced becomes an error carrying the
// offset it started at, and every other match becomes a token or a skip depending on the rule which matched.
func (s *Scanner) Event(match Match) fmt.Stringer {
	rule, ok := s.rule(match)
	if !ok {
		return backendtest.ScannerError{Offset: match.Start}
	}

	lexeme := string(s.source[match.Start:match.End])
	if rule.Skip {
		return backendtest.Skip{RuleName: rule.Name, Start: match.Start, End: match.End, Lexeme: lexeme}
	}
	return backendtest.Token{RuleName: rule.Name, Start: match.Start, End: match.End, Lexeme: lexeme}
}

// ScanTrace scans the whole input and returns the canonical trace of it. Every trace ends with the end of input event,
// including the trace of an empty input, so that a trace always states the length of the input it belongs to.
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
	return append(result, backendtest.EOF{Offset: len(source)})
}
