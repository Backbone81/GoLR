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

// partialMatchDFA is a scanner in which a failed match can be followed by a real token: "-" leads into the state which
// only ARROW completes, and the byte which ends that attempt is one an IDENT can start with.
func partialMatchDFA() backend.DFA {
	return rulesToDFA(
		dsl.Rule("ARROW", dsl.Literal("->")),
		dsl.Rule("IDENT", dsl.OneOrMore(dsl.CharClass(dsl.CharRange('a', 'z')))),
	)
}

var _ = Describe("Scanner", func() {
	Context("maximal munch", func() {
		It("prefers the longer match over the keyword", func() {
			expectScannerTrace(keywordDFA(), "iffy",
				`1:1     TOKEN   NAME "iffy"`,
				"1:5     EOF",
			)
		})

		It("breaks a tie at equal length by rule order", func() {
			expectScannerTrace(keywordDFA(), "if",
				`1:1     TOKEN   IF "if"`,
				"1:3     EOF",
			)
		})

		It("backs up to the last accepting state when it runs into a dead state", func() {
			expectScannerTrace(backupDFA(), "abc",
				`1:1     TOKEN   AB "ab"`,
				`1:3     ERROR   "c"`,
				"1:4     EOF",
			)
		})

		It("consumes the longer alternative when the input completes it", func() {
			expectScannerTrace(backupDFA(), "abcd",
				`1:1     TOKEN   ABCD "abcd"`,
				"1:5     EOF",
			)
		})

		It("takes the last accepting state up to the very last byte of the input", func() {
			expectScannerTrace(backupDFA(), "abcab",
				`1:1     TOKEN   AB "ab"`,
				`1:3     ERROR   "c"`,
				`1:4     TOKEN   AB "ab"`,
				"1:6     EOF",
			)
		})
	})

	Context("skipped rules", func() {
		It("reports them as ordinary tokens, because a runner is never told which rules are skipped", func() {
			expectScannerTrace(keywordDFA(), "if fy",
				`1:1     TOKEN   IF "if"`,
				`1:3     TOKEN   WHITESPACE " "`,
				`1:4     TOKEN   NAME "fy"`,
				"1:6     EOF",
			)
		})
	})

	Context("input no rule matches", func() {
		It("reports the position the failed match started at", func() {
			expectScannerTrace(keywordDFA(), "a?b",
				`1:1     TOKEN   NAME "a"`,
				`1:2     ERROR   "?"`,
				`1:3     TOKEN   NAME "b"`,
				"1:4     EOF",
			)
		})

		It("keeps scanning instead of giving up on the first error", func() {
			expectScannerTrace(keywordDFA(), "??",
				`1:1     ERROR   "?"`,
				`1:2     ERROR   "?"`,
				"1:3     EOF",
			)
		})

		It("carries the whole stretch of bytes it could not match", func() {
			expectScannerTrace(deadEndDFA(), "abcx",
				`1:1     ERROR   "abc"`,
				`1:4     ERROR   "x"`,
				"1:5     EOF",
			)
		})

		It("leaves the token which follows a failed match whole", func() {
			expectScannerTrace(partialMatchDFA(), "-ab-",
				`1:1     ERROR   "-"`,
				`1:2     TOKEN   IDENT "ab"`,
				`1:4     ERROR   "-"`,
				"1:5     EOF",
			)
		})

		It("reports an input which ends in the middle of a token", func() {
			expectScannerTrace(backupDFA(), "abc",
				`1:1     TOKEN   AB "ab"`,
				`1:3     ERROR   "c"`,
				"1:4     EOF",
			)
		})

		It("stops at the end of the input rather than one byte past it", func() {
			expectScannerTrace(deadEndDFA(), "abc",
				`1:1     ERROR   "abc"`,
				"1:4     EOF",
			)
		})
	})

	Context("bytes", func() {
		It("counts the column in bytes and escapes a lexeme byte by byte", func() {
			expectScannerTrace(rulesToDFA(dsl.Rule("UMLAUT", dsl.Literal("ä"))), "ää",
				`1:1     TOKEN   UMLAUT "\xc3\xa4"`,
				`1:3     TOKEN   UMLAUT "\xc3\xa4"`,
				"1:5     EOF",
			)
		})

		It("scans the byte at the lower end of the byte range", func() {
			expectScannerTrace(rulesToDFA(dsl.Rule("NUL", dsl.CharClass(dsl.CharRange(0, 0)))), "\x00",
				`1:1     TOKEN   NUL "\x00"`,
				"1:2     EOF",
			)
		})
	})

	Context("edge cases", func() {
		It("traces an empty input as the end of input alone", func() {
			expectScannerTrace(keywordDFA(), "",
				"1:1     EOF",
			)
		})

		It("never lets a rule which matches the empty string match nothing", func() {
			emptyDFA := rulesToDFA(dsl.Rule("AS", dsl.ZeroOrMore(dsl.CharClass(dsl.CharRange('a', 'a')))))

			expectScannerTrace(emptyDFA, "aa",
				`1:1     TOKEN   AS "aa"`,
				"1:3     EOF",
			)
			expectScannerTrace(emptyDFA, "b",
				`1:1     ERROR   "b"`,
				"1:2     EOF",
			)
		})
	})

	Context("the GoLR specification", func() {
		It("traces a real scanner of non-trivial size", func() {
			expectScannerTrace(golrSpecDFA(), `@parser { a: "b"; }`,
				`1:1     TOKEN   PARSER "@parser"`,
				`1:8     TOKEN   WHITESPACE " "`,
				`1:9     TOKEN   LBRACE "{"`,
				`1:10    TOKEN   WHITESPACE " "`,
				`1:11    TOKEN   IDENTIFIER "a"`,
				`1:12    TOKEN   COLON ":"`,
				`1:13    TOKEN   WHITESPACE " "`,
				`1:14    TOKEN   STRING "\"b\""`,
				`1:17    TOKEN   SEMI ";"`,
				`1:18    TOKEN   WHITESPACE " "`,
				`1:19    TOKEN   RBRACE "}"`,
				"1:20    EOF",
			)
		})

		It("reports a string which the end of the input cuts short", func() {
			expectScannerTrace(golrSpecDFA(), `"abc`,
				`1:1     ERROR   "\"abc"`,
				"1:5     EOF",
			)
		})
	})
})
