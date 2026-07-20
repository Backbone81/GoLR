package oracle_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
	ielr1golrcore "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	"github.com/backbone81/golr/internal/parsergen/core/ielr1/golr/oracle"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// runToCompletion drives the interpreter to the end and returns the full action sequence together with the final
// action. The final action is the last element of the sequence, returned separately for readable assertions.
func runToCompletion(parser backend.Parser, input []int) (oracle.ParserAction, []oracle.ParserAction) {
	interpreter := oracle.NewParserInterpreter(parser, input)
	var actions []oracle.ParserAction
	for interpreter.Next() {
		actions = append(actions, interpreter.Value())
	}
	return interpreter.Value(), actions
}

var _ = Describe("Parser Interpreter", func() {
	Describe("on a hand-built resolved table", func() {
		// A minimal resolved LR table for the grammar `S -> a`, augmented to `$accept -> S $end`. Building the table by
		// hand keeps the interpreter test independent of the IELR(1) pipeline and pins down the exact table shape the
		// interpreter runs against, including the accept state whose reduce action carries an empty lookahead set.
		//
		// Terminals:    $end (0), a (1)
		// Nonterminals: $accept (0), S (1)
		// Productions:  0: $accept -> S $end     1: S -> a
		grammar := frontend.Grammar{
			Terminals:    []frontend.Symbol{{Name: "$end"}, {Name: "a"}},
			Nonterminals: []frontend.Symbol{{Name: "$accept"}, {Name: "S"}},
			Productions: []frontend.Production{
				{
					NonterminalIdx: 0,
					SymbolRefs: []frontend.SymbolRef{
						frontend.NewNonterminalRef(1), // S
						frontend.NewTerminalRef(0),    // $end
					},
				},
				{
					NonterminalIdx: 1,
					SymbolRefs: []frontend.SymbolRef{
						frontend.NewTerminalRef(1), // a
					},
				},
			},
			StartNonterminalIdx: 0,
		}
		parser := backend.Parser{
			Grammar: grammar,
			States: []backend.State{
				// state 0: $accept -> . S $end ; S -> . a
				{
					KernelItems: backend.NewCoreSet(backend.NewCore(0, 0)),
					TransitionActions: backend.NewTransitionActionSet(
						backend.NewTransitionAction(frontend.NewTerminalRef(1), 2),    // shift a -> state 2
						backend.NewTransitionAction(frontend.NewNonterminalRef(1), 1), // goto S -> state 1
					),
				},
				// state 1: $accept -> S . $end
				{
					KernelItems: backend.NewCoreSet(backend.NewCore(0, 1)),
					TransitionActions: backend.NewTransitionActionSet(
						backend.NewTransitionAction(frontend.NewTerminalRef(0), 3), // shift $end -> state 3
					),
				},
				// state 2: S -> a .
				{
					KernelItems: backend.NewCoreSet(backend.NewCore(1, 1)),
					ReduceActions: backend.NewReduceActionSet(
						backend.NewReduceAction(backend.NewLookaheadSet(0), 1), // reduce S -> a on $end
					),
				},
				// state 3: $accept -> S $end . — the accept, encoded as a reduce of production 0 with empty lookahead.
				{
					KernelItems: backend.NewCoreSet(backend.NewCore(0, 2)),
					ReduceActions: backend.NewReduceActionSet(
						backend.NewReduceAction(backend.NewLookaheadSet(), 0),
					),
				},
			},
		}

		It("accepts a valid sentence and reports the full shift/reduce/accept sequence", func() {
			final, sequence := runToCompletion(parser, []int{1}) // a
			Expect(final.Kind).To(Equal(oracle.ParserActionAccept))
			Expect(sequence).To(Equal([]oracle.ParserAction{
				{Kind: oracle.ParserActionShift, TerminalIdx: 1, ProductionIdx: -1},  // shift a
				{Kind: oracle.ParserActionReduce, TerminalIdx: -1, ProductionIdx: 1}, // reduce S -> a
				{Kind: oracle.ParserActionShift, TerminalIdx: 0, ProductionIdx: -1},  // shift $end
				{Kind: oracle.ParserActionAccept, TerminalIdx: -1, ProductionIdx: -1},
			}))
		})

		It("rejects empty input, which the grammar does not derive", func() {
			final, _ := runToCompletion(parser, []int{})
			Expect(final.Kind).To(Equal(oracle.ParserActionReject))
		})

		It("rejects input with a trailing token past a complete sentence", func() {
			final, _ := runToCompletion(parser, []int{1, 1}) // a a
			Expect(final.Kind).To(Equal(oracle.ParserActionReject))
		})

		It("keeps returning the final action once the parse is done", func() {
			interpreter := oracle.NewParserInterpreter(parser, []int{1})
			for interpreter.Next() {
			}
			// Next reports no further progress, and Value keeps yielding the accept.
			Expect(interpreter.Next()).To(BeFalse())
			Expect(interpreter.Value().Kind).To(Equal(oracle.ParserActionAccept))
		})
	})

	Describe("on a real resolved table from GrammarToParser", func() {
		// UnambiguousTestGrammarFig1 (`S -> aAa | bAb`, `A -> a | aa`) is unambiguous but not LR(1): after `aa` with a
		// single terminal of lookahead the parser cannot tell whether the middle A is `a` or `aa`. Building against the
		// real IELR(1) pipeline grounds the interpreter's assumptions about the actual table shape.
		//
		// The augmentation fixes the terminal indexes: $end = 0, a = 1, b = 2, and the production indexes:
		// 0: $accept -> S $end, 1: S -> a A a, 2: S -> b A b, 3: A -> a, 4: A -> a a.
		var parser backend.Parser
		BeforeEach(func() {
			var err error
			parser, _, err = ielr1golrcore.GrammarToParser(ielr1golrcore.UnambiguousTestGrammarFig1)
			Expect(err).NotTo(HaveOccurred())
		})

		It("accepts `bab`, where the middle A is decided by the following b", func() {
			final, sequence := runToCompletion(parser, []int{2, 1, 2}) // b a b
			Expect(final.Kind).To(Equal(oracle.ParserActionAccept))
			Expect(sequence).To(Equal([]oracle.ParserAction{
				{Kind: oracle.ParserActionShift, TerminalIdx: 2, ProductionIdx: -1},  // shift b
				{Kind: oracle.ParserActionShift, TerminalIdx: 1, ProductionIdx: -1},  // shift a
				{Kind: oracle.ParserActionReduce, TerminalIdx: -1, ProductionIdx: 3}, // reduce A -> a
				{Kind: oracle.ParserActionShift, TerminalIdx: 2, ProductionIdx: -1},  // shift b
				{Kind: oracle.ParserActionReduce, TerminalIdx: -1, ProductionIdx: 2}, // reduce S -> b A b
				{Kind: oracle.ParserActionShift, TerminalIdx: 0, ProductionIdx: -1},  // shift $end
				{Kind: oracle.ParserActionAccept, TerminalIdx: -1, ProductionIdx: -1},
			}))
		})

		It("accepts `aaaa`, which the shift-over-reduce resolution parses as S -> a (A -> a a) a", func() {
			final, _ := runToCompletion(parser, []int{1, 1, 1, 1})
			Expect(final.Kind).To(Equal(oracle.ParserActionAccept))
		})

		It("rejects `aaa`: valid in the language but unreachable once the conflict is resolved toward shift", func() {
			// The resolved table commits to A -> a a on the shift, then runs out of input. This is the resolved-table
			// behavior the interpreter must reproduce, not a claim that `aaa` is outside the language.
			final, _ := runToCompletion(parser, []int{1, 1, 1})
			Expect(final.Kind).To(Equal(oracle.ParserActionReject))
		})

		It("rejects `ba`, which is not a sentence", func() {
			final, _ := runToCompletion(parser, []int{2, 1})
			Expect(final.Kind).To(Equal(oracle.ParserActionReject))
		})
	})
})
