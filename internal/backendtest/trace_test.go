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
			Entry("SHIFT",
				backendtest.Shift{TerminalName: "NUMBER"},
				"SHIFT NUMBER"),
			Entry("REDUCE",
				backendtest.Reduce{NonterminalName: "expr", RightHandSideLength: 3},
				"REDUCE expr 3"),
			Entry("REDUCE of an empty production",
				backendtest.Reduce{NonterminalName: "opt", RightHandSideLength: 0},
				"REDUCE opt 0"),
			Entry("ERROR",
				backendtest.ParserError{Offset: 5},
				"ERROR 5"),
			Entry("RESYNC carries no arguments",
				backendtest.Resync{},
				"RESYNC"),
			Entry("ACCEPT carries no arguments",
				backendtest.Accept{},
				"ACCEPT"),
		)

		It("writes a whole trace with one event per line and a trailing newline", func() {
			trace := backendtest.Trace{
				backendtest.ParserError{Offset: 2},
				backendtest.Shift{TerminalName: "NUMBER"},
				backendtest.Reduce{NonterminalName: "expr", RightHandSideLength: 1},
				backendtest.Resync{},
				backendtest.Accept{},
			}

			Expect(trace.String()).To(Equal(
				"ERROR 2\n" +
					"SHIFT NUMBER\n" +
					"REDUCE expr 1\n" +
					"RESYNC\n" +
					"ACCEPT\n",
			))
		})

		It("writes an empty trace as the empty string", func() {
			Expect(backendtest.Trace{}.String()).To(Equal(""))
		})
	})
})
