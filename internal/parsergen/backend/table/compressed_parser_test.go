package table_test

import (
	"fmt"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/backend/table"
	"github.com/backbone81/golr/internal/parsergen/core"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
	"github.com/backbone81/golr/testdata"
)

// ptr returns a pointer to a copy of the value, for building the DefaultReduceProductionIdx of a state.
func ptr[T any](value T) *T {
	return &value
}

// handBuiltGrammar returns a grammar with the given number of terminals and nonterminals, the last terminal being the
// error symbol. Laying the tables out only reads how many symbols there are and where the error symbol sits, so the
// productions are left out.
func handBuiltGrammar(terminalCount int, nonterminalCount int) frontend.Grammar {
	grammar := frontend.Grammar{
		Terminals:    make([]frontend.Symbol, terminalCount),
		Nonterminals: make([]frontend.Symbol, nonterminalCount),
	}
	for terminalIdx := range grammar.Terminals {
		grammar.Terminals[terminalIdx] = frontend.Symbol{Name: "t" + strconv.Itoa(terminalIdx)}
	}
	for nonterminalIdx := range grammar.Nonterminals {
		grammar.Nonterminals[nonterminalIdx] = frontend.Symbol{Name: "n" + strconv.Itoa(nonterminalIdx)}
	}
	grammar.Terminals[terminalCount-1] = frontend.SymbolError
	return grammar
}

// packingDensity returns the share of the packed array which holds a real entry. The remaining cells are the holes
// which no row could be placed on, and are what the placement heuristic is judged by.
func packingDensity(displacement utils.RowDisplacement) float64 {
	if len(displacement.Check) == 0 {
		return 1
	}

	var occupiedCells int
	for _, colIdx := range displacement.Check {
		if colIdx != utils.NoColumn {
			occupiedCells++
		}
	}
	return float64(occupiedCells) / float64(len(displacement.Check))
}

// expectDecodeEquivalence asserts that for every state, every terminal and every nonterminal the compressed lookup
// returns exactly what reading the parser states directly returns, and that the error recovery finds the same
// resynchronization points. This covers the default actions and the row displacement together, because a lookup goes
// through both.
//
// The comparison is a plain conditional which only reports through Gomega once it found a mismatch. The well-known
// grammars have millions of cells between them, and a matcher call per cell would have the suite spend its runtime
// inside Gomega instead of inside the code under test. The first mismatch ends the check: a lookup which decodes a
// single cell differently from the state model is a defect of the compression, and the cells behind it have nothing
// left to add.
func expectDecodeEquivalence(parser backend.Parser) table.CompressedParser {
	compressed := table.NewCompressedParser(parser)

	errorRef, hasErrorTerminal := frontend.ErrorTerminalRef(parser.Grammar)
	errorIdx := table.NoTerminal
	if hasErrorTerminal {
		errorIdx = errorRef.Idx()
	}

	Expect(compressed.StateCount()).To(Equal(len(parser.States)))
	Expect(compressed.ErrorTerminalIdx).To(Equal(errorIdx))

	for stateIdx := range parser.States {
		state := parser.States[stateIdx]

		for terminalIdx := range parser.Grammar.Terminals {
			want := referenceAction(state, terminalIdx, errorIdx)
			if got := compressed.Action(stateIdx, terminalIdx); got != want {
				Expect(got).To(Equal(want), "state %d on terminal %d", stateIdx, terminalIdx)
			}
		}

		for nonterminalIdx := range parser.Grammar.Nonterminals {
			want := referenceGoto(state, nonterminalIdx)
			if got := compressed.Goto(stateIdx, nonterminalIdx); got != want {
				Expect(got).To(Equal(want), "state %d on nonterminal %d", stateIdx, nonterminalIdx)
			}
		}

		wantStateIdx, wantOk := 0, false
		if hasErrorTerminal {
			wantStateIdx, wantOk = backend.ErrorShiftStateIdx(&parser.States[stateIdx], errorRef)
		}
		gotStateIdx, gotOk := compressed.ErrorShiftStateIdx(stateIdx)
		if gotOk != wantOk || (wantOk && gotStateIdx != wantStateIdx) {
			// At least one of the two fails, and whichever does ends the check. The state index is only compared once
			// both agree that the state shifts the error symbol at all, which the first assertion establishes.
			Expect(gotOk).To(Equal(wantOk), "state %d shifting the error symbol", stateIdx)
			Expect(gotStateIdx).To(Equal(wantStateIdx), "state %d shifting the error symbol", stateIdx)
		}
	}

	return compressed
}

var _ = Describe("CompressedParser", func() {
	It("takes the default action of the state for a terminal without an action of its own", func() {
		parser := backend.Parser{
			Grammar: handBuiltGrammar(4, 2),
			States: []backend.State{{
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(0), 0),
				),
				DefaultReduceProductionIdx: ptr(3),
			}},
		}

		compressed := expectDecodeEquivalence(parser)

		Expect(compressed.Action(0, 0)).To(Equal(table.NewShiftAction(0)))
		Expect(compressed.Action(0, 1)).To(Equal(table.NewReduceAction(3)))
		Expect(compressed.Action(0, 2)).To(Equal(table.NewReduceAction(3)))
	})

	It("keeps a terminal which the state rejects an error, even though the state reduces on anything else", func() {
		// This is what a terminal declared nonassociative asks for. The rejection only survives the default reduction
		// because it is an entry of its own instead of an absent one.
		parser := backend.Parser{
			Grammar: handBuiltGrammar(4, 2),
			States: []backend.State{{
				DefaultReduceProductionIdx: ptr(3),
				RejectedTerminals:          backend.NewLookaheadSet(2),
			}},
		}

		compressed := expectDecodeEquivalence(parser)

		Expect(compressed.Action(0, 1)).To(Equal(table.NewReduceAction(3)))
		Expect(compressed.Action(0, 2)).To(Equal(table.NewErrorAction()))
	})

	It("reports a syntax error for a state which has neither an action nor a default action", func() {
		parser := backend.Parser{
			Grammar: handBuiltGrammar(4, 2),
			States:  []backend.State{{}},
		}

		compressed := expectDecodeEquivalence(parser)

		Expect(compressed.Action(0, 1)).To(Equal(table.NoAction))
	})

	It("reports the accept for both of the ways a core encodes it", func() {
		// The Bison backed cores report the accept as `$default accept`.
		bisonStyle := backend.Parser{
			Grammar: handBuiltGrammar(4, 2),
			States:  []backend.State{{DefaultReduceProductionIdx: ptr(0)}},
		}
		// The native GoLR cores keep it a reduce of the accept production with an empty reduction lookahead set.
		golrStyle := backend.Parser{
			Grammar: handBuiltGrammar(4, 2),
			States: []backend.State{{
				ReduceActions: backend.NewReduceActionSet(backend.NewReduceAction(backend.NewLookaheadSet(), 0)),
			}},
		}

		bisonCompressed := expectDecodeEquivalence(bisonStyle)
		golrCompressed := expectDecodeEquivalence(golrStyle)

		Expect(bisonCompressed.Action(0, 1)).To(Equal(table.NewAcceptAction()))
		Expect(golrCompressed.Action(0, 1)).To(Equal(table.NewAcceptAction()))
	})

	It("finds the resynchronization points of the error recovery in the column of the error terminal", func() {
		grammar := handBuiltGrammar(3, 1)
		errorRef, _ := frontend.ErrorTerminalRef(grammar)
		parser := backend.Parser{
			Grammar: grammar,
			States: []backend.State{
				{
					TransitionActions: backend.NewTransitionActionSet(
						backend.NewTransitionAction(errorRef, 1),
					),
				},
				{},
			},
		}

		compressed := expectDecodeEquivalence(parser)

		Expect(compressed.HasErrorRecovery()).To(BeTrue())
		stateIdx, ok := compressed.ErrorShiftStateIdx(0)
		Expect(ok).To(BeTrue())
		Expect(stateIdx).To(Equal(1))

		_, ok = compressed.ErrorShiftStateIdx(1)
		Expect(ok).To(BeFalse())
	})

	It("reports no error recovery for a grammar which does not use the error symbol", func() {
		parser := backend.Parser{
			Grammar: frontend.Grammar{
				Terminals:    []frontend.Symbol{{Name: "t0"}},
				Nonterminals: []frontend.Symbol{{Name: "n0"}},
			},
			States: []backend.State{{}},
		}

		compressed := expectDecodeEquivalence(parser)

		Expect(compressed.ErrorTerminalIdx).To(Equal(table.NoTerminal))
		Expect(compressed.HasErrorRecovery()).To(BeFalse())
	})

	It("supports a parser without any state", func() {
		compressed := table.NewCompressedParser(backend.Parser{})

		Expect(compressed.StateCount()).To(Equal(0))
		Expect(compressed.Actions.Base).To(BeEmpty())
		Expect(compressed.Gotos.Base).To(BeEmpty())
		Expect(compressed.ErrorTerminalIdx).To(Equal(table.NoTerminal))
		Expect(compressed.HasErrorRecovery()).To(BeFalse())
	})

	for _, wellKnownGrammar := range testdata.WellKnownGrammars {
		It("decodes like the parser states for "+wellKnownGrammar.Title, func() {
			parser := wellKnownParser(wellKnownGrammar)

			compressed := expectDecodeEquivalence(parser)

			AddReportEntry("tables", fmt.Sprintf(
				"%d states, %d terminals, %d nonterminals, %d action cells, %d goto cells",
				len(parser.States), len(parser.Grammar.Terminals), len(parser.Grammar.Nonterminals),
				len(compressed.Actions.Next), len(compressed.Gotos.Next),
			))
		})
	}

	It("decodes like the parser states for tables which were not compressed by default reductions", func() {
		// Without the default reductions every state lists every one of its reduce actions explicitly, which makes the
		// rows of the action table far denser than the row displacement usually sees them. The lookups have to come out
		// the same either way.
		parser := wellKnownParser(wellKnownGrammarByTitle("Go 1.5.4"), core.WithoutDefaultReductions())

		expectDecodeEquivalence(parser)
	})

	It("needs far fewer cells than the uncompressed tables", func() {
		parser := wellKnownParser(wellKnownGrammarByTitle("Go 1.5.4"))

		compressed := expectDecodeEquivalence(parser)

		denseActionCells := len(parser.States) * len(parser.Grammar.Terminals)
		denseGotoCells := len(parser.States) * len(parser.Grammar.Nonterminals)
		AddReportEntry("cells", fmt.Sprintf(
			"%d states, %d terminals, %d nonterminals, %d action cells packed instead of %d (%.1f%% filled), "+
				"%d goto cells packed instead of %d (%.1f%% filled)",
			len(parser.States), len(parser.Grammar.Terminals), len(parser.Grammar.Nonterminals),
			len(compressed.Actions.Next), denseActionCells, 100*packingDensity(compressed.Actions),
			len(compressed.Gotos.Next), denseGotoCells, 100*packingDensity(compressed.Gotos),
		))
		Expect(len(compressed.Actions.Next)).To(BeNumerically("<", denseActionCells/4))
		Expect(len(compressed.Gotos.Next)).To(BeNumerically("<", denseGotoCells/4))

		// Holes which no row could use are the cost of the placement. Parser tables do not reach the 96 to 98 percent
		// the scanner tables pack to, and cannot: a scanner row has an entry in most of its few byte classes, while a
		// parser row holds a handful of entries spread over all terminals or all nonterminals, and every row which is
		// not identical to another one needs a displacement of its own. The Go grammar packs to roughly 74 percent of
		// the action cells and 57 percent of the goto cells, so half the array holding real entries is the level at
		// which something has regressed rather than the level to aim for.
		Expect(packingDensity(compressed.Actions)).To(BeNumerically(">", 0.5))
		Expect(packingDensity(compressed.Gotos)).To(BeNumerically(">", 0.5))
	})
})
