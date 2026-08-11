package interpreter_test

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/backbone81/golr/internal/scannergen/backend"
	"github.com/backbone81/golr/internal/scannergen/frontend/dsl"
)

// keywordDFA is a scanner in which a keyword and an identifier can match the same input, which is the pair every
// tie break between two rules comes down to.
func keywordDFA() backend.DFA {
	return rulesToDFA(
		dsl.Rule("IF", dsl.Literal("if")),
		dsl.Rule("NAME", dsl.OneOrMore(dsl.CharClass(dsl.CharRange('a', 'z')))),
		dsl.SkipRule("WHITESPACE", dsl.OneOrMore(dsl.CharClass(dsl.CharRange(' ', ' ')))),
	)
}

// backupDFA is the smallest scanner which forces the maximal munch to back up: reading the third byte of "abc" leads
// into a state which only "abcd" can still reach an accepting state from.
func backupDFA() backend.DFA {
	return rulesToDFA(
		dsl.Rule("AB", dsl.Literal("ab")),
		dsl.Rule("ABCD", dsl.Literal("abcd")),
	)
}

// deadEndDFA is a scanner in which no prefix of its single rule accepts, so that an input which starts like the rule
// but does not complete it leaves the automaton without any match to fall back to.
func deadEndDFA() backend.DFA {
	return rulesToDFA(dsl.Rule("ABCD", dsl.Literal("abcd")))
}

var _ = Describe("Scanner", func() {
	Context("maximal munch", func() {
		It("prefers the longer match over the keyword", func() {
			expectScannerTrace(keywordDFA(), "iffy",
				`TOKEN NAME 0 4 "iffy"`,
				"EOF 4",
			)
		})

		It("breaks a tie at equal length by rule order", func() {
			expectScannerTrace(keywordDFA(), "if",
				`TOKEN IF 0 2 "if"`,
				"EOF 2",
			)
		})

		It("backs up to the last accepting state when it runs into a dead state", func() {
			expectScannerTrace(backupDFA(), "abc",
				`TOKEN AB 0 2 "ab"`,
				"ERROR 2",
				"EOF 3",
			)
		})

		It("consumes the longer alternative when the input completes it", func() {
			expectScannerTrace(backupDFA(), "abcd",
				`TOKEN ABCD 0 4 "abcd"`,
				"EOF 4",
			)
		})

		It("takes the last accepting state up to the very last byte of the input", func() {
			expectScannerTrace(backupDFA(), "abcab",
				`TOKEN AB 0 2 "ab"`,
				"ERROR 2",
				`TOKEN AB 3 5 "ab"`,
				"EOF 5",
			)
		})
	})

	Context("skipped rules", func() {
		It("reports them as ordinary tokens, because a runner is never told which rules are skipped", func() {
			expectScannerTrace(keywordDFA(), "if fy",
				`TOKEN IF 0 2 "if"`,
				`TOKEN WHITESPACE 2 3 " "`,
				`TOKEN NAME 3 5 "fy"`,
				"EOF 5",
			)
		})
	})

	Context("input no rule matches", func() {
		It("reports the offset the failed match started at", func() {
			expectScannerTrace(keywordDFA(), "a?b",
				`TOKEN NAME 0 1 "a"`,
				"ERROR 1",
				`TOKEN NAME 2 3 "b"`,
				"EOF 3",
			)
		})

		It("keeps scanning instead of giving up on the first error", func() {
			expectScannerTrace(keywordDFA(), "??",
				"ERROR 0",
				"ERROR 1",
				"EOF 2",
			)
		})

		It("consumes everything it looked at, including the byte it could not consume", func() {
			expectScannerTrace(deadEndDFA(), "abcx",
				"ERROR 0",
				"EOF 4",
			)
		})

		It("reports an input which ends in the middle of a token", func() {
			expectScannerTrace(backupDFA(), "abc",
				`TOKEN AB 0 2 "ab"`,
				"ERROR 2",
				"EOF 3",
			)
		})

		It("stops at the end of the input rather than one byte past it", func() {
			expectScannerTrace(deadEndDFA(), "abc",
				"ERROR 0",
				"EOF 3",
			)
		})
	})

	Context("bytes", func() {
		It("reports offsets in bytes and escapes a lexeme byte by byte", func() {
			expectScannerTrace(rulesToDFA(dsl.Rule("UMLAUT", dsl.Literal("ä"))), "ää",
				`TOKEN UMLAUT 0 2 "\xc3\xa4"`,
				`TOKEN UMLAUT 2 4 "\xc3\xa4"`,
				"EOF 4",
			)
		})

		It("scans the byte at the lower end of the byte range", func() {
			expectScannerTrace(rulesToDFA(dsl.Rule("NUL", dsl.CharClass(dsl.CharRange(0, 0)))), "\x00",
				`TOKEN NUL 0 1 "\x00"`,
				"EOF 1",
			)
		})
	})

	Context("edge cases", func() {
		It("traces an empty input as the end of input alone", func() {
			expectScannerTrace(keywordDFA(), "",
				"EOF 0",
			)
		})

		It("never lets a rule which matches the empty string match nothing", func() {
			emptyDFA := rulesToDFA(dsl.Rule("AS", dsl.ZeroOrMore(dsl.CharClass(dsl.CharRange('a', 'a')))))

			expectScannerTrace(emptyDFA, "aa",
				`TOKEN AS 0 2 "aa"`,
				"EOF 2",
			)
			expectScannerTrace(emptyDFA, "b",
				"ERROR 0",
				"EOF 1",
			)
		})
	})

	Context("the GoLR specification", func() {
		It("traces a real scanner of non-trivial size", func() {
			expectScannerTrace(golrSpecDFA(), `@parser { a: "b"; }`,
				`TOKEN PARSER 0 7 "@parser"`,
				`TOKEN WHITESPACE 7 8 " "`,
				`TOKEN LBRACE 8 9 "{"`,
				`TOKEN WHITESPACE 9 10 " "`,
				`TOKEN NAME 10 11 "a"`,
				`TOKEN COLON 11 12 ":"`,
				`TOKEN WHITESPACE 12 13 " "`,
				`TOKEN STRING 13 16 "\"b\""`,
				`TOKEN SEMI 16 17 ";"`,
				`TOKEN WHITESPACE 17 18 " "`,
				`TOKEN RBRACE 18 19 "}"`,
				"EOF 19",
			)
		})

		It("reports a string which the end of the input cuts short", func() {
			expectScannerTrace(golrSpecDFA(), `"abc`,
				"ERROR 0",
				"EOF 4",
			)
		})
	})
})
