package counterexample_test

import (
	"bytes"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	ielr1golr "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	lalr1golr "github.com/backbone81/golr/internal/parsergen/core/lalr1/golr"
	lr1golr "github.com/backbone81/golr/internal/parsergen/core/lr1/golr"
	"github.com/backbone81/golr/internal/parsergen/counterexample"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	bisonfrontend "github.com/backbone81/golr/internal/parsergen/frontend/bison"
	golrfrontend "github.com/backbone81/golr/internal/parsergen/frontend/golr"
	"github.com/backbone81/golr/internal/utils"
	"github.com/backbone81/golr/testdata"
)

// figure1Spec is the ambiguous grammar of figure 1 of the paper, with the dangling else and the challenging conflict of
// section 3.1.
var figure1Spec = utils.HereDoc(`
	@scanner {
	    WHITESPACE: /[ \t\r\n]+/ @skip;
	    IF:         "if";
	    THEN:       "then";
	    ELSE:       "else";
	    QUESTION:   "?";
	    ARR:        "arr";
	    LBRACKET:   "[";
	    RBRACKET:   "]";
	    ASSIGN:     ":=";
	    PLUS:       "+";
	    DIGIT:      /[0-9]/;
	}

	@parser {
	    stmt
	        : "if" expr "then" stmt "else" stmt
	        | "if" expr "then" stmt
	        | expr "?" stmt stmt
	        | "arr" "[" expr "]" ":=" expr
	        ;

	    expr
	        : num
	        | expr "+" expr
	        ;

	    num
	        : DIGIT
	        | num DIGIT
	        ;
	}
`)

// figure3Spec is the unambiguous LR(2) grammar of figure 3 of the paper.
var figure3Spec = utils.HereDoc(`
	@scanner {
	    A: "a";
	    B: "b";
	}

	@parser {
	    s : t | s t ;
	    t : x | y ;
	    x : "a" ;
	    y : "a" "a" "b" ;
	}
`)

// figure7Spec is the ambiguous grammar of figure 7 of the paper.
var figure7Spec = utils.HereDoc(`
	@scanner {
	    N: "n";
	    A: "a";
	    B: "b";
	    C: "c";
	    D: "d";
	}

	@parser {
	    s : n | n "c" ;
	    n : "n" n "d" | "n" n "c" | "n" a "b" | "n" b ;
	    a : "a" ;
	    b : "a" "b" "c" | "a" "b" "d" ;
	}
`)

// emptySpec is a grammar whose lookaheads and derivations reach through nonterminals which can vanish.
var emptySpec = utils.HereDoc(`
	@scanner {
	    P: "p";
	    Q: "q";
	    W: "w";
	    Z: "z";
	}

	@parser {
	    s : x y "z" | "w" s ;
	    x : @empty | "p" ;
	    y : x | "q" ;
	}
`)

// precedenceSpec is an ambiguous expression grammar whose conflicts are all decided by precedence declarations.
var precedenceSpec = utils.HereDoc(`
	@scanner {
	    INTEGER:  /[0-9]+/;
	    PLUS:     "+";
	    MULTIPLY: "*";
	}

	@parser {
	    @precedence {
	        @left: "*";
	        @left: "+";
	    }

	    expression
	        : INTEGER
	        | expression "+" expression
	        | expression "*" expression
	        ;
	}
`)

// unresolvedParser builds the parser tables of a core with their conflicts intact.
type unresolvedParser func(frontend.Grammar, conflict.PolicyFactory) (backend.Parser, []utils.Warning, error)

// unresolvedParsers are the cores whose tables the lookup tables are built on.
var unresolvedParsers = map[string]unresolvedParser{
	"ielr1": ielr1golr.GrammarToUnresolvedParser,
	"lalr1": lalr1golr.GrammarToUnresolvedParser,
	"lr1":   lr1golr.GrammarToUnresolvedParser,
}

var _ = Describe("LookupTables", func() {
	DescribeTable("should propagate the lookaheads every core computed for its reductions",
		func(spec string) {
			_, grammar, err := golrfrontend.GrammarFromString(spec)
			Expect(err).ToNot(HaveOccurred())
			for coreName, grammarToUnresolvedParser := range unresolvedParsers {
				parser, _, err := grammarToUnresolvedParser(grammar, conflict.DefaultPolicy)
				Expect(err).ToNot(HaveOccurred())
				tables := counterexample.NewLookupTables(parser, conflict.DefaultPolicy(parser.Grammar))
				expectReductionLookaheads(tables, parser, coreName)
			}
		},
		Entry("for figure 1", figure1Spec),
		Entry("for figure 3", figure3Spec),
		Entry("for figure 7", figure7Spec),
		Entry("for nonterminals which can vanish", emptySpec),
		Entry("for precedence declarations", precedenceSpec),
	)

	Context("well known grammars", func() {
		for _, wellKnownGrammar := range testdata.WellKnownGrammars {
			It("should propagate the lookaheads the cores computed for the reductions of "+wellKnownGrammar.Title, func() {
				grammar, err := bisonfrontend.ToGrammar(
					bytes.NewBuffer(wellKnownGrammar.Content()),
					wellKnownGrammar.FileName,
				)
				Expect(err).ToNot(HaveOccurred())
				// Canonical LR(1) is left out, as its tables take too long to build for a unit test.
				for _, coreName := range []string{"ielr1", "lalr1"} {
					parser, _, err := unresolvedParsers[coreName](grammar, conflict.DefaultPolicy)
					Expect(err).ToNot(HaveOccurred())
					tables := counterexample.NewLookupTables(parser, conflict.DefaultPolicy(parser.Grammar))
					expectReductionLookaheads(tables, parser, coreName)
				}
			})
		}
	})

	DescribeTable("should invert the transitions and the production steps",
		func(spec string) {
			_, grammar, err := golrfrontend.GrammarFromString(spec)
			Expect(err).ToNot(HaveOccurred())
			for coreName, grammarToUnresolvedParser := range unresolvedParsers {
				parser, _, err := grammarToUnresolvedParser(grammar, conflict.DefaultPolicy)
				Expect(err).ToNot(HaveOccurred())
				tables := counterexample.NewLookupTables(parser, conflict.DefaultPolicy(parser.Grammar))
				expectConsistentEdges(tables, parser, coreName)
			}
		},
		Entry("for figure 1", figure1Spec),
		Entry("for figure 7", figure7Spec),
		Entry("for nonterminals which can vanish", emptySpec),
	)

	It("should find the items of the dangling else", func() {
		_, grammar, err := golrfrontend.GrammarFromString(figure1Spec)
		Expect(err).ToNot(HaveOccurred())
		parser, _, _, err := ielr1golr.GrammarToParser(grammar, conflict.DefaultPolicy)
		Expect(err).ToNot(HaveOccurred())
		tables := counterexample.NewLookupTables(parser, conflict.DefaultPolicy(parser.Grammar))
		elseIdx := terminalIdx(parser.Grammar, `"else"`)

		for _, reduceItemIdx := range itemsWithCore(tables, parser.Grammar, `stmt -> "if" expr "then" stmt`, 4) {
			stateIdx := tables.StateIdx(reduceItemIdx)
			Expect(tables.Lookahead(reduceItemIdx).Contains(elseIdx)).To(BeTrue())

			shiftItemIdxs := tables.ItemsWithNextSymbol(stateIdx, frontend.NewTerminalRef(elseIdx))
			Expect(shiftItemIdxs).To(HaveLen(1))
			Expect(tables.Core(shiftItemIdxs[0])).To(Equal(backend.NewCore(
				productionIdx(parser.Grammar, `stmt -> "if" expr "then" stmt "else" stmt`), 4,
			)))

			reachingItemIdxs := tables.ReachingItems(reduceItemIdx)
			startItemIdx, found := tables.ItemIdx(0, backend.NewCore(0, 0))
			Expect(found).To(BeTrue())
			Expect(reachingItemIdxs.Contains(startItemIdx)).To(BeTrue())
			Expect(reachingItemIdxs.Contains(reduceItemIdx)).To(BeTrue())
			for _, otherReduceItemIdx := range itemsWithCore(tables, parser.Grammar, "num -> num DIGIT", 2) {
				Expect(reachingItemIdxs.Contains(otherReduceItemIdx)).To(BeFalse())
			}
		}
	})

	Context("IsActionAllowed", func() {
		It("should allow both actions of a conflict a rule of last resort decided", func() {
			_, grammar, err := golrfrontend.GrammarFromString(figure1Spec)
			Expect(err).ToNot(HaveOccurred())
			parser, _, _, err := ielr1golr.GrammarToParser(grammar, conflict.DefaultPolicy)
			Expect(err).ToNot(HaveOccurred())
			tables := counterexample.NewLookupTables(parser, conflict.DefaultPolicy(parser.Grammar))
			elseIdx := terminalIdx(parser.Grammar, `"else"`)
			reduceItemIdxs := itemsWithCore(tables, parser.Grammar, `stmt -> "if" expr "then" stmt`, 4)
			Expect(reduceItemIdxs).ToNot(BeEmpty())

			for _, reduceItemIdx := range reduceItemIdxs {
				stateIdx := tables.StateIdx(reduceItemIdx)
				reduce := conflict.NewReduceContribution(tables.Core(reduceItemIdx).ProductionIdx())
				// Shift over reduce removed the reduction from the tables, but the grammar author still has to see it.
				Expect(tables.IsActionAllowed(stateIdx, elseIdx, reduce)).To(BeTrue())
				Expect(tables.IsActionAllowed(stateIdx, elseIdx, conflict.NewShiftContribution())).To(BeTrue())
			}
		})

		It("should only allow the action a precedence declaration chose", func() {
			_, grammar, err := golrfrontend.GrammarFromString(precedenceSpec)
			Expect(err).ToNot(HaveOccurred())
			resolvedParser, _, _, err := ielr1golr.GrammarToParser(grammar, conflict.DefaultPolicy)
			Expect(err).ToNot(HaveOccurred())
			unresolvedParser, _, err := ielr1golr.GrammarToUnresolvedParser(grammar, conflict.DefaultPolicy)
			Expect(err).ToNot(HaveOccurred())

			// The filter does not depend on the tables being resolved already.
			for _, parser := range []backend.Parser{resolvedParser, unresolvedParser} {
				tables := counterexample.NewLookupTables(parser, conflict.DefaultPolicy(parser.Grammar))
				plusIdx := terminalIdx(parser.Grammar, `"+"`)
				reduceItemIdxs := itemsWithCore(tables, parser.Grammar, `expression -> expression "+" expression`, 3)
				Expect(reduceItemIdxs).ToNot(BeEmpty())

				for _, reduceItemIdx := range reduceItemIdxs {
					stateIdx := tables.StateIdx(reduceItemIdx)
					reduce := conflict.NewReduceContribution(tables.Core(reduceItemIdx).ProductionIdx())
					// "+" is left associative, so the reduction wins over the shift.
					Expect(tables.IsActionAllowed(stateIdx, plusIdx, reduce)).To(BeTrue())
					Expect(tables.IsActionAllowed(stateIdx, plusIdx, conflict.NewShiftContribution())).To(BeFalse())
				}
			}
		})

		It("should not allow an action the state does not have", func() {
			_, grammar, err := golrfrontend.GrammarFromString(figure1Spec)
			Expect(err).ToNot(HaveOccurred())
			parser, _, _, err := ielr1golr.GrammarToParser(grammar, conflict.DefaultPolicy)
			Expect(err).ToNot(HaveOccurred())
			tables := counterexample.NewLookupTables(parser, conflict.DefaultPolicy(parser.Grammar))
			thenIdx := terminalIdx(parser.Grammar, `"then"`)
			for _, reduceItemIdx := range itemsWithCore(tables, parser.Grammar, "num -> DIGIT", 1) {
				stateIdx := tables.StateIdx(reduceItemIdx)
				Expect(tables.IsActionAllowed(stateIdx, thenIdx, conflict.NewShiftContribution())).To(BeFalse())
			}
		})
	})

	Context("DerivationStartingWith", func() {
		It("should derive a nonterminal beginning with a terminal", func() {
			_, grammar, err := golrfrontend.GrammarFromString(figure1Spec)
			Expect(err).ToNot(HaveOccurred())
			parser, _, _, err := ielr1golr.GrammarToParser(grammar, conflict.DefaultPolicy)
			Expect(err).ToNot(HaveOccurred())
			tables := counterexample.NewLookupTables(parser, conflict.DefaultPolicy(parser.Grammar))
			digitIdx := terminalIdx(parser.Grammar, "DIGIT")

			expectDerivationStep(tables, parser.Grammar, "stmt", digitIdx, `stmt -> expr "?" stmt stmt`, 0)
			expectDerivationStep(tables, parser.Grammar, "expr", digitIdx, "expr -> num", 0)
			expectDerivationStep(tables, parser.Grammar, "num", digitIdx, "num -> DIGIT", 0)

			_, _, found := tables.DerivationStartingWith(
				nonterminalIdx(parser.Grammar, "stmt"),
				terminalIdx(parser.Grammar, `"else"`),
			)
			Expect(found).To(BeFalse())
		})

		It("should derive the symbols in front of the terminal to the empty string", func() {
			_, grammar, err := golrfrontend.GrammarFromString(emptySpec)
			Expect(err).ToNot(HaveOccurred())
			parser, _, _, err := ielr1golr.GrammarToParser(grammar, conflict.DefaultPolicy)
			Expect(err).ToNot(HaveOccurred())
			tables := counterexample.NewLookupTables(parser, conflict.DefaultPolicy(parser.Grammar))

			expectDerivationStep(tables, parser.Grammar, "s", terminalIdx(parser.Grammar, `"z"`), `s -> x y "z"`, 2)
			expectDerivationStep(tables, parser.Grammar, "s", terminalIdx(parser.Grammar, `"q"`), `s -> x y "z"`, 1)
			expectDerivationStep(tables, parser.Grammar, "s", terminalIdx(parser.Grammar, `"w"`), `s -> "w" s`, 0)

			emptyProductionIdx, found := tables.EmptyDerivation(nonterminalIdx(parser.Grammar, "y"))
			Expect(found).To(BeTrue())
			Expect(emptyProductionIdx).To(Equal(productionIdx(parser.Grammar, "y -> x")))
			emptyProductionIdx, found = tables.EmptyDerivation(nonterminalIdx(parser.Grammar, "x"))
			Expect(found).To(BeTrue())
			Expect(emptyProductionIdx).To(Equal(productionIdx(parser.Grammar, "x -> (empty)")))
			_, found = tables.EmptyDerivation(nonterminalIdx(parser.Grammar, "s"))
			Expect(found).To(BeFalse())
		})
	})
})

// expectReductionLookaheads checks that the lookahead set of every reduce item is the lookahead set the core computed
// for its reduction.
func expectReductionLookaheads(tables counterexample.LookupTables, parser backend.Parser, coreName string) {
	for stateIdx, state := range parser.States {
		for _, reduceAction := range state.ReduceActions.All() {
			production := parser.Grammar.Productions[reduceAction.ProductionIdx]
			itemIdx, found := tables.ItemIdx(stateIdx, backend.NewCore(reduceAction.ProductionIdx, len(production.SymbolRefs)))
			Expect(found).To(BeTrue(), "core %s, state %d, reduction %d", coreName, stateIdx, reduceAction.ProductionIdx)
			Expect(tables.Lookahead(itemIdx).Equal(reduceAction.LookaheadSet)).To(
				BeTrue(),
				"core %s, state %d, reduction %d: propagated %s, computed by the core %s",
				coreName, stateIdx, reduceAction.ProductionIdx,
				tables.Lookahead(itemIdx).String(), reduceAction.LookaheadSet.String(),
			)
		}
	}

	// Every reduce item with a lookahead has a reduction, so the reductions checked above are all there are.
	for itemIdx := range tables.ItemCount() {
		if _, hasNextSymbol := tables.NextSymbol(itemIdx); hasNextSymbol || tables.Lookahead(itemIdx).IsEmpty() {
			continue
		}
		productionIdxs := reducedProductionIdxs(parser.States[tables.StateIdx(itemIdx)])
		Expect(productionIdxs).To(
			ContainElement(tables.Core(itemIdx).ProductionIdx()),
			"core %s, state %d", coreName, tables.StateIdx(itemIdx),
		)
	}
}

// reducedProductionIdxs returns the productions the state has a reduce action for.
func reducedProductionIdxs(state backend.State) []int {
	var result []int
	for _, reduceAction := range state.ReduceActions.All() {
		result = append(result, reduceAction.ProductionIdx)
	}
	return result
}

// expectConsistentEdges checks that every edge of the lookup tables is found again in the opposite direction, and that
// the items name their state and core.
func expectConsistentEdges(tables counterexample.LookupTables, parser backend.Parser, coreName string) {
	for itemIdx := range tables.ItemCount() {
		description := fmt.Sprintf("core %s, item %d", coreName, itemIdx)
		stateIdx := tables.StateIdx(itemIdx)
		core := tables.Core(itemIdx)

		foundItemIdx, found := tables.ItemIdx(stateIdx, core)
		Expect(found).To(BeTrue(), description)
		Expect(foundItemIdx).To(Equal(itemIdx), description)

		if targetItemIdx, found := tables.Transition(itemIdx); found {
			Expect(tables.Core(targetItemIdx)).To(Equal(backend.NewCore(core.ProductionIdx(), core.Position()+1)), description)
			Expect(tables.ReverseTransitions(targetItemIdx)).To(ContainElement(itemIdx), description)
		}
		for _, sourceItemIdx := range tables.ReverseTransitions(itemIdx) {
			targetItemIdx, found := tables.Transition(sourceItemIdx)
			Expect(found).To(BeTrue(), description)
			Expect(targetItemIdx).To(Equal(itemIdx), description)
		}

		from, to := tables.ProductionSteps(itemIdx)
		nextSymbol, hasNextSymbol := tables.NextSymbol(itemIdx)
		if hasNextSymbol && nextSymbol.IsNonterminal() {
			Expect(to-from).To(Equal(productionCount(parser.Grammar, nextSymbol.Idx())), description)
		} else {
			Expect(to-from).To(BeZero(), description)
		}
		for targetItemIdx := from; targetItemIdx < to; targetItemIdx++ {
			Expect(tables.StateIdx(targetItemIdx)).To(Equal(stateIdx), description)
			targetCore := tables.Core(targetItemIdx)
			Expect(targetCore.Position()).To(BeZero(), description)
			targetNonterminalIdx := parser.Grammar.Productions[targetCore.ProductionIdx()].NonterminalIdx
			Expect(targetNonterminalIdx).To(Equal(nextSymbol.Idx()), description)
			Expect(tables.ReverseProductionSteps(targetItemIdx)).To(ContainElement(itemIdx), description)
		}
		for _, sourceItemIdx := range tables.ReverseProductionSteps(itemIdx) {
			sourceFrom, sourceTo := tables.ProductionSteps(sourceItemIdx)
			Expect(itemIdx).To(BeNumerically(">=", sourceFrom), description)
			Expect(itemIdx).To(BeNumerically("<", sourceTo), description)
		}
	}
}

// expectDerivationStep checks the first step of the shortest derivation of the nonterminal beginning with the terminal.
func expectDerivationStep(
	tables counterexample.LookupTables,
	grammar frontend.Grammar,
	nonterminalName string,
	terminalIdx int,
	wantProduction string,
	wantPosition int,
) {
	gotProductionIdx, gotPosition, found := tables.DerivationStartingWith(
		nonterminalIdx(grammar, nonterminalName),
		terminalIdx,
	)
	Expect(found).To(BeTrue(), "%s beginning with %s", nonterminalName, grammar.Terminals[terminalIdx])
	Expect(frontend.FormatProduction(grammar, gotProductionIdx)).To(Equal(wantProduction))
	Expect(gotPosition).To(Equal(wantPosition))
}

// itemsWithCore returns the items of all states with the production and the position of the dot.
func itemsWithCore(
	tables counterexample.LookupTables,
	grammar frontend.Grammar,
	production string,
	position int,
) []int {
	core := backend.NewCore(productionIdx(grammar, production), position)
	var result []int
	for itemIdx := range tables.ItemCount() {
		if tables.Core(itemIdx) == core {
			result = append(result, itemIdx)
		}
	}
	return result
}

// productionIdx returns the production which FormatProduction writes as the text.
func productionIdx(grammar frontend.Grammar, text string) int {
	for productionIdx := range grammar.Productions {
		if frontend.FormatProduction(grammar, productionIdx) == text {
			return productionIdx
		}
	}
	Fail("no production " + text)
	return 0
}

// productionCount returns the number of productions of the nonterminal.
func productionCount(grammar frontend.Grammar, nonterminalIdx int) int {
	result := 0
	for _, production := range grammar.Productions {
		if production.NonterminalIdx == nonterminalIdx {
			result++
		}
	}
	return result
}

// terminalIdx returns the terminal which is written as the text.
func terminalIdx(grammar frontend.Grammar, text string) int {
	for terminalIdx, terminal := range grammar.Terminals {
		if terminal.String() == text {
			return terminalIdx
		}
	}
	Fail("no terminal " + text)
	return 0
}

// nonterminalIdx returns the nonterminal with the name.
func nonterminalIdx(grammar frontend.Grammar, name string) int {
	for nonterminalIdx, nonterminal := range grammar.Nonterminals {
		if nonterminal.Name == name {
			return nonterminalIdx
		}
	}
	Fail("no nonterminal " + name)
	return 0
}
