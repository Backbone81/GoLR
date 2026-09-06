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
			Entry("TOKEN",
				backendtest.Token{RuleName: "IDENTIFIER", Start: 0, End: 5, Lexeme: "hello"},
				`TOKEN IDENTIFIER 0 5 "hello"`),
			Entry("ERROR",
				backendtest.ScannerError{Offset: 7},
				"ERROR 7"),
			Entry("EOF",
				backendtest.EOF{Offset: 12},
				"EOF 12"),
			// The quotes are what let a lexeme contain the space which separates the other fields.
			Entry("a lexeme with spaces",
				backendtest.Token{RuleName: "WHITESPACE", Start: 0, End: 3, Lexeme: "  x"},
				`TOKEN WHITESPACE 0 3 "  x"`),
			Entry("a lexeme which needs escaping",
				backendtest.Token{RuleName: "STRING", Start: 0, End: 6, Lexeme: "\"a\\b\"\n"},
				`TOKEN STRING 0 6 "\"a\\b\"\n"`),
			Entry("an empty lexeme",
				backendtest.Token{RuleName: "EMPTY", Start: 4, End: 4, Lexeme: ""},
				`TOKEN EMPTY 4 4 ""`),
		)

		It("writes a whole trace with one event per line and a trailing newline", func() {
			trace := backendtest.Trace{
				backendtest.Token{RuleName: "IF", Start: 0, End: 2, Lexeme: "if"},
				backendtest.Token{RuleName: "WHITESPACE", Start: 2, End: 3, Lexeme: " "},
				backendtest.ScannerError{Offset: 3},
				backendtest.EOF{Offset: 4},
			}

			Expect(trace.String()).To(Equal(`TOKEN IF 0 2 "if"` + "\n" + `TOKEN WHITESPACE 2 3 " "` + "\nERROR 3\nEOF 4\n"))
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
