package backendtest_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/backendtest"
)

var _ = Describe("Scanner events", func() {
	Context("Scanner", func() {
		DescribeTable("writes an event as its canonical line",
			func(event fmt.Stringer, expected string) {
				Expect(event.String()).To(Equal(expected))
			},
			Entry("TOKEN carries the position, the rule and the escaped lexeme",
				backendtest.Token{Line: 1, Column: 1, RuleName: "IDENTIFIER", Lexeme: "hello"},
				`1:1     TOKEN   IDENTIFIER "hello"`),
			Entry("ERROR carries the position and the bytes no rule matched",
				backendtest.ScannerError{Line: 1, Column: 8, Lexeme: "?"},
				`1:8     ERROR   "?"`),
			Entry("EOF carries only the position",
				backendtest.EOF{Line: 2, Column: 1},
				"2:1     EOF"),
			Entry("a long position pushes the rest right instead of being truncated",
				backendtest.Token{Line: 1234, Column: 56, RuleName: "NAME", Lexeme: "x"},
				`1234:56 TOKEN   NAME "x"`),
			// The quotes are what let a lexeme contain the space which separates the other fields.
			Entry("a lexeme with spaces",
				backendtest.Token{Line: 1, Column: 1, RuleName: "WHITESPACE", Lexeme: "  x"},
				`1:1     TOKEN   WHITESPACE "  x"`),
			Entry("a lexeme which needs escaping",
				backendtest.Token{Line: 1, Column: 1, RuleName: "STRING", Lexeme: "\"a\\b\"\n"},
				`1:1     TOKEN   STRING "\"a\\b\"\n"`),
			Entry("an empty lexeme",
				backendtest.Token{Line: 4, Column: 2, RuleName: "EMPTY", Lexeme: ""},
				`4:2     TOKEN   EMPTY ""`),
			Entry("an error whose bytes need escaping",
				backendtest.ScannerError{Line: 1, Column: 1, Lexeme: "\xff"},
				`1:1     ERROR   "\xff"`),
		)

		It("writes a whole trace with one event per line and a trailing newline", func() {
			trace := backendtest.Trace{
				backendtest.Token{Line: 1, Column: 1, RuleName: "IF", Lexeme: "if"},
				backendtest.Token{Line: 1, Column: 3, RuleName: "WHITESPACE", Lexeme: " "},
				backendtest.ScannerError{Line: 1, Column: 4, Lexeme: "?"},
				backendtest.EOF{Line: 1, Column: 5},
			}

			Expect(trace.String()).To(Equal(
				`1:1     TOKEN   IF "if"` + "\n" +
					`1:3     TOKEN   WHITESPACE " "` + "\n" +
					`1:4     ERROR   "?"` + "\n" +
					"1:5     EOF\n",
			))
		})
	})

	Context("Parser", func() {
		DescribeTable("writes an event as its canonical line",
			func(event fmt.Stringer, expected string) {
				Expect(event.String()).To(Equal(expected))
			},
			Entry("SHIFT carries the position, the terminal and the escaped lexeme",
				backendtest.Shift{Line: 1, Column: 1, TerminalName: "NUMBER", Lexeme: "42"},
				`1:1     SHIFT   NUMBER "42"`),
			Entry("REDUCE spells out the right hand side",
				backendtest.Reduce{Line: 1, Column: 3, LeftHandSide: "expr", RightHandSide: []string{"expr", "PLUS", "term"}},
				"1:3     REDUCE  expr => expr PLUS term"),
			Entry("REDUCE of an empty production uses an epsilon",
				backendtest.Reduce{Line: 2, Column: 5, LeftHandSide: "opt"},
				"2:5     REDUCE  opt => ε"),
			Entry("a long position pushes the rest right instead of being truncated",
				backendtest.Reduce{Line: 1234, Column: 56, LeftHandSide: "stmt", RightHandSide: []string{"expr"}},
				"1234:56 REDUCE  stmt => expr"),
			Entry("ERROR carries the position and a description",
				backendtest.ParserError{Line: 3, Column: 8, Detail: "unexpected token RBRACE"},
				"3:8     ERROR   unexpected token RBRACE"),
			Entry("ERROR marks a suppressed error",
				backendtest.ParserError{Line: 3, Column: 8, Detail: "unexpected token RBRACE", Suppressed: true},
				"3:8     ERROR   (suppressed) unexpected token RBRACE"),
			Entry("DISCARD carries the dropped terminal and its lexeme",
				backendtest.Discard{Line: 3, Column: 8, TerminalName: "RBRACE", Lexeme: "}"},
				`3:8     DISCARD RBRACE "}"`),
			Entry("POP names the dropped symbol",
				backendtest.Pop{Line: 3, Column: 8, SymbolName: "term"},
				"3:8     POP     term"),
			Entry("RESYNC carries only the position",
				backendtest.Resync{Line: 3, Column: 8},
				"3:8     RESYNC"),
			Entry("FAIL carries only the position",
				backendtest.Fail{Line: 3, Column: 8},
				"3:8     FAIL"),
			Entry("ACCEPT carries only the position",
				backendtest.Accept{Line: 1, Column: 10},
				"1:10    ACCEPT"),
		)

		It("writes a whole trace with one event per line and a trailing newline", func() {
			trace := backendtest.Trace{
				backendtest.Shift{Line: 1, Column: 1, TerminalName: "NUMBER", Lexeme: "1"},
				backendtest.Reduce{Line: 1, Column: 3, LeftHandSide: "expr", RightHandSide: []string{"NUMBER"}},
				backendtest.ParserError{Line: 1, Column: 3, Detail: "unexpected token PLUS"},
				backendtest.Pop{Line: 1, Column: 3, SymbolName: "expr"},
				backendtest.Resync{Line: 1, Column: 3},
				backendtest.Accept{Line: 1, Column: 5},
			}

			Expect(trace.String()).To(Equal(
				`1:1     SHIFT   NUMBER "1"` + "\n" +
					"1:3     REDUCE  expr => NUMBER\n" +
					"1:3     ERROR   unexpected token PLUS\n" +
					"1:3     POP     expr\n" +
					"1:3     RESYNC\n" +
					"1:5     ACCEPT\n",
			))
		})

		It("writes an empty trace as the empty string", func() {
			Expect(backendtest.Trace{}.String()).To(Equal(""))
		})
	})
})
