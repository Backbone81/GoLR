package counterexample_test

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	"github.com/backbone81/golr/internal/parsergen/core"
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

// mergedReduceReduceSpec is an unambiguous LR(1) grammar whose LALR(1) tables merge the states behind "a" "c" and
// "b" "c", which gives them a reduce/reduce conflict on "d" and on "e".
var mergedReduceReduceSpec = utils.HereDoc(`
	@scanner {
	    A: "a";
	    B: "b";
	    C: "c";
	    D: "d";
	    E: "e";
	}

	@parser {
	    s : "a" x "d" | "b" y "d" | "a" y "e" | "b" x "e" ;
	    x : "c" ;
	    y : "c" ;
	}
`)

// blockedInnermostSpec is an ambiguous grammar whose unifying search reaches configurations which show a nonterminal
// both parsers are in, but whose productions keep the conflict terminal from directly following the dot.
var blockedInnermostSpec = utils.HereDoc(`
	@scanner {
	    A: "a";
	    B: "b";
	    C: "c";
	}

	@parser {
	    s : "a" | "b" u v | s s "c" "b" ;
	    v : @empty ;
	    u : @empty | v u s | "c" "a" ;
	}
`)

// leftEmptyOperandSpec is an ambiguous grammar with an empty operand and a left associative "+". Behind s "+", the
// empty operand conflicts with "+" "x" on "+", and its reduction is followed by s "+" s • "+", where the parser
// reduces.
var leftEmptyOperandSpec = utils.HereDoc(`
	@scanner {
	    PLUS: "+";
	    X:    "x";
	}

	@parser {
	    @precedence {
	        @left: "+";
	    }

	    s : @empty | s "+" s | "+" "x" ;
	}
`)

// rightEmptyOperandSpec is leftEmptyOperandSpec with a right associative "+", so the parser shifts at s "+" s • "+".
var rightEmptyOperandSpec = utils.HereDoc(`
	@scanner {
	    PLUS: "+";
	    X:    "x";
	}

	@parser {
	    @precedence {
	        @right: "+";
	    }

	    s : @empty | s "+" s | "+" "x" ;
	}
`)

// nonassociativeEmptyOperandSpec is leftEmptyOperandSpec with a non-associative "+", so the parser rejects "+" at
// s "+" s • "+".
var nonassociativeEmptyOperandSpec = utils.HereDoc(`
	@scanner {
	    PLUS: "+";
	    X:    "x";
	}

	@parser {
	    @precedence {
	        @none: "+";
	    }

	    s : @empty | s "+" s | "+" "x" ;
	}
`)

// nestedNonassociativeSpec is a grammar whose empty reduction behind "a" s has the lookahead "a" only from nested
// "a" "a", which the non-associative "a" rejects.
var nestedNonassociativeSpec = utils.HereDoc(`
	@scanner {
	    A: "a";
	}

	@parser {
	    @precedence {
	        @none: "a";
	    }

	    s : @empty | "a" | "a" s s ;
	}
`)

// resolvedParser builds the parser tables of a core with their conflicts resolved.
type resolvedParser func(
	frontend.Grammar,
	conflict.PolicyFactory,
	...core.Option,
) (backend.Parser, []conflict.Conflict, []utils.Warning, error)

// resolvedParsers are the cores the counterexamples are searched on.
var resolvedParsers = map[string]resolvedParser{
	"ielr1": ielr1golr.GrammarToParser,
	"lalr1": lalr1golr.GrammarToParser,
	"lr1":   lr1golr.GrammarToParser,
}

var _ = Describe("Find", func() {
	It("should find the unifying counterexample of the challenging conflict", func() {
		Expect(findText(figure1Spec, "ielr1", "DIGIT")).To(Equal(utils.HereDoc(`
			example: expr "?" "arr" "[" expr "]" ":=" num • DIGIT DIGIT "?" stmt stmt
			using the reduction:
			  stmt -> expr "?" [stmt] [stmt]
			    stmt -> "arr" "[" expr "]" ":=" [expr]
			      expr -> num •
			    stmt -> [expr] "?" stmt stmt
			      expr -> [num]
			        num -> [num] DIGIT
			          num -> DIGIT
			using the shift:
			  stmt -> expr "?" [stmt] [stmt]
			    stmt -> "arr" "[" expr "]" ":=" [expr]
			      expr -> [num]
			        num -> num • DIGIT
			    stmt -> [expr] "?" stmt stmt
			      expr -> [num]
			        num -> DIGIT
		`)))
	})

	It("should find the unifying counterexample of the dangling else", func() {
		Expect(findText(figure1Spec, "ielr1", `"else"`)).To(Equal(utils.HereDoc(`
			example: "if" expr "then" "if" expr "then" stmt • "else" stmt
			using the reduction:
			  stmt -> "if" expr "then" [stmt] "else" stmt
			    stmt -> "if" expr "then" stmt •
			using the shift:
			  stmt -> "if" expr "then" [stmt]
			    stmt -> "if" expr "then" stmt • "else" stmt
		`)))
	})

	It("should find the unifying counterexample of the innermost nonterminal", func() {
		Expect(findText(figure1Spec, "ielr1", `"+"`)).To(Equal(utils.HereDoc(`
			example: expr "+" expr • "+" expr
			using the reduction:
			  expr -> [expr] "+" expr
			    expr -> expr "+" expr •
			using the shift:
			  expr -> expr "+" [expr]
			    expr -> expr • "+" expr
		`)))
	})

	It("should find unifying counterexamples which leave the shortest lookahead-sensitive path", func() {
		Expect(findText(figure7Spec, "ielr1", `"b"`)).To(Equal(utils.HereDoc(`
			example: "n" "a" • "b" "c"
			using the reduction:
			  s -> [n] "c"
			    n -> "n" [a] "b"
			      a -> "a" •
			using the shift:
			  s -> [n]
			    n -> "n" [b]
			      b -> "a" • "b" "c"

			example: "n" "n" "a" • "b" "d" "c"
			using the reduction:
			  s -> [n] "c"
			    n -> "n" [n] "d"
			      n -> "n" [a] "b"
			        a -> "a" •
			using the shift:
			  s -> [n]
			    n -> "n" [n] "c"
			      n -> "n" [b]
			        b -> "a" • "b" "d"
		`)))
	})

	It("should find the unifying counterexample of two reductions", func() {
		Expect(findText(endOfInputSpec, "ielr1", "$end")).To(Equal(utils.HereDoc(`
			example: "x" •
			using the first reduction:
			  s -> "x" •
			using the second reduction:
			  s -> "x" [t]
			    t -> (empty) •
		`)))
	})

	It("should fall back to the nonunifying counterexample of a grammar which is not ambiguous", func() {
		Expect(findText(figure3Spec, "ielr1", `"a"`)).To(
			Equal(findText(figure3Spec, "ielr1", `"a"`, counterexample.WithTotalTimeLimit(0))),
		)
		Expect(findText(mergedReduceReduceSpec, "lalr1", `"d"`)).To(
			Equal(findText(mergedReduceReduceSpec, "lalr1", `"d"`, counterexample.WithTotalTimeLimit(0))),
		)
	})

	It("should find the nonunifying counterexample of the challenging conflict", func() {
		Expect(findText(figure1Spec, "ielr1", "DIGIT", counterexample.WithTotalTimeLimit(0))).To(Equal(utils.HereDoc(`
			example: expr "?" "arr" "[" expr "]" ":=" num • DIGIT "?" stmt stmt
			using the reduction:
			  stmt -> expr "?" [stmt] [stmt]
			    stmt -> "arr" "[" expr "]" ":=" [expr]
			      expr -> num •
			    stmt -> [expr] "?" stmt stmt
			      expr -> [num]
			        num -> DIGIT
			example: expr "?" "arr" "[" expr "]" ":=" num • DIGIT stmt
			using the shift:
			  stmt -> expr "?" [stmt] stmt
			    stmt -> "arr" "[" expr "]" ":=" [expr]
			      expr -> [num]
			        num -> num • DIGIT
		`)))
	})

	It("should share the path of the reduction up to the latest point of divergence", func() {
		Expect(findText(figure1Spec, "ielr1", `"else"`, counterexample.WithTotalTimeLimit(0))).To(Equal(utils.HereDoc(`
			example: "if" expr "then" "if" expr "then" stmt • "else" stmt
			using the reduction:
			  stmt -> "if" expr "then" [stmt] "else" stmt
			    stmt -> "if" expr "then" stmt •
			example: "if" expr "then" "if" expr "then" stmt • "else" stmt "else" stmt
			using the shift:
			  stmt -> "if" expr "then" [stmt] "else" stmt
			    stmt -> "if" expr "then" stmt • "else" stmt
		`)))
	})

	It("should find the nonunifying counterexample of a grammar which is not LR(1)", func() {
		Expect(findText(figure3Spec, "ielr1", `"a"`, counterexample.WithTotalTimeLimit(0))).To(Equal(utils.HereDoc(`
			example: "a" • "a"
			using the reduction:
			  s -> [s] [t]
			    s -> [t]
			      t -> [x]
			        x -> "a" •
			    t -> [x]
			      x -> "a"
			example: "a" • "a" "b" t
			using the shift:
			  s -> [s] t
			    s -> [t]
			      t -> [y]
			        y -> "a" • "a" "b"
		`)))
	})

	It("should find a counterexample per shift item", func() {
		Expect(findText(figure7Spec, "ielr1", `"b"`, counterexample.WithTotalTimeLimit(0))).To(Equal(utils.HereDoc(`
			example: "n" "a" • "b"
			using the reduction:
			  s -> [n]
			    n -> "n" [a] "b"
			      a -> "a" •
			example: "n" "a" • "b" "c"
			using the shift:
			  s -> [n]
			    n -> "n" [b]
			      b -> "a" • "b" "c"

			example: "n" "a" • "b"
			using the reduction:
			  s -> [n]
			    n -> "n" [a] "b"
			      a -> "a" •
			example: "n" "a" • "b" "d"
			using the shift:
			  s -> [n]
			    n -> "n" [b]
			      b -> "a" • "b" "d"
		`)))
	})

	It("should require the terminal behind the second reduction", func() {
		Expect(findText(reduceReduceSpec, "ielr1", `"z"`, counterexample.WithTotalTimeLimit(0))).To(Equal(utils.HereDoc(`
			example: "x" • "z"
			using the first reduction:
			  s -> [b] "z"
			    b -> "x" •
			example: "x" • "z"
			using the second reduction:
			  s -> [a] "z"
			    a -> "x" [e]
			      e -> (empty) •
		`)))
	})

	It("should derive the nonterminals behind the dot to the empty string", func() {
		Expect(findText(emptySpec, "ielr1", `"p"`, counterexample.WithTotalTimeLimit(0))).To(Equal(utils.HereDoc(`
			example: • "p" "z"
			using the reduction:
			  s -> [x] [y] "z"
			    x -> (empty) •
			    y -> [x]
			      x -> "p"
			example: • "p" y "z"
			using the shift:
			  s -> [x] y "z"
			    x -> • "p"

			example: "w" • "p" "z"
			using the reduction:
			  s -> "w" [s]
			    s -> [x] [y] "z"
			      x -> (empty) •
			      y -> [x]
			        x -> "p"
			example: "w" • "p" y "z"
			using the shift:
			  s -> "w" [s]
			    s -> [x] y "z"
			      x -> • "p"
		`)))
	})

	It("should search a path of its own for a second reduction which the path of the first one cannot reach", func() {
		text := findText(mergedReduceReduceSpec, "lalr1", `"d"`, counterexample.WithTotalTimeLimit(0))
		Expect(text).To(Equal(utils.HereDoc(`
			example: "a" "c" • "d"
			using the first reduction:
			  s -> "a" [x] "d"
			    x -> "c" •
			example: "b" "c" • "d"
			using the second reduction:
			  s -> "b" [y] "d"
			    y -> "c" •
		`)))
	})

	It("should keep the start production when the conflict terminal is the end of the input", func() {
		Expect(findText(endOfInputSpec, "ielr1", "$end", counterexample.WithTotalTimeLimit(0))).To(Equal(utils.HereDoc(`
			example: "x" • $end
			using the first reduction:
			  $accept -> [s] $end
			    s -> "x" •
			example: "x" • $end
			using the second reduction:
			  $accept -> [s] $end
			    s -> "x" [t]
			      t -> (empty) •
		`)))
	})

	It("should carry the conflict terminal to its shift as the associativity decides", func() {
		text := findText(leftEmptyOperandSpec, "ielr1", `"+"`, counterexample.WithTotalTimeLimit(0))
		Expect(text).To(Equal(utils.HereDoc(`
			example: • "+" s
			using the reduction:
			  s -> [s] "+" s
			    s -> (empty) •
			example: • "+" "x"
			using the shift:
			  s -> • "+" "x"

			example: s "+" • "+" s
			using the reduction:
			  s -> [s] "+" s
			    s -> s "+" [s]
			      s -> (empty) •
			example: s "+" • "+" "x" "+" s
			using the shift:
			  s -> [s] "+" s
			    s -> s "+" [s]
			      s -> • "+" "x"
		`)))
		text = findText(rightEmptyOperandSpec, "ielr1", `"+"`, counterexample.WithTotalTimeLimit(0))
		Expect(text).To(Equal(utils.HereDoc(`
			example: • "+" s
			using the reduction:
			  s -> [s] "+" s
			    s -> (empty) •
			example: • "+" "x"
			using the shift:
			  s -> • "+" "x"

			example: s "+" • "+" s
			using the reduction:
			  s -> s "+" [s]
			    s -> [s] "+" s
			      s -> (empty) •
			example: s "+" • "+" "x"
			using the shift:
			  s -> s "+" [s]
			    s -> • "+" "x"
		`)))
	})

	DescribeTable("should find no counterexample where the declarations never let the terminal follow the reduction",
		func(spec string, coreName string, wantCounterexampleCountByKernelItems map[string]int) {
			_, grammar, err := golrfrontend.GrammarFromString(spec)
			Expect(err).ToNot(HaveOccurred())
			parser, conflicts, _, err := resolvedParsers[coreName](grammar, conflict.DefaultPolicy)
			Expect(err).ToNot(HaveOccurred())
			counterexamplesByConflictIdx := counterexample.Find(parser, conflicts, counterexample.WithTotalTimeLimit(0))
			gotCounterexampleCountByKernelItems := map[string]int{}
			for conflictIdx, c := range conflicts {
				kernelItems := strings.Join(formatKernelItems(parser, c.StateIdx), ", ")
				gotCounterexampleCountByKernelItems[kernelItems] = len(counterexamplesByConflictIdx[conflictIdx])
			}
			Expect(gotCounterexampleCountByKernelItems).To(Equal(wantCounterexampleCountByKernelItems))
		},
		// Behind s "+" s, the non-associative "+" rejects "+", so the empty operand behind s "+" is never followed by it.
		Entry("for a terminal rejected behind the reduction", nonassociativeEmptyOperandSpec, "ielr1", map[string]int{
			`$accept -> • s $end`: 1,
			`s -> s "+" • s`:      0,
		}),
		// LALR(1) merges the state behind the first "a" with the one behind a nested "a", where the non-associative "a"
		// rejects the shift of "a". IELR(1) keeps them apart, and the rule of last resort keeps the shift behind the
		// first "a", so "a" "a" s • "a" is a counterexample there.
		Entry("for a lookahead only a rejected shift brings", nestedNonassociativeSpec, "lalr1", map[string]int{
			`s -> "a" •, s -> "a" • s s`: 1,
			`s -> "a" s • s`:             0,
		}),
	)

	It("should find the same counterexamples on tables whose conflicts are left unresolved", func() {
		_, grammar, err := golrfrontend.GrammarFromString(figure1Spec)
		Expect(err).ToNot(HaveOccurred())
		resolvedParser, resolvedConflicts, _, err := lalr1golr.GrammarToParser(grammar, conflict.DefaultPolicy)
		Expect(err).ToNot(HaveOccurred())
		unresolvedParser, unresolvedConflicts, _, err := lalr1golr.GrammarToParser(grammar, conflict.SelectPolicy(true, true))
		Expect(err).To(HaveOccurred())

		resolved := counterexample.Find(resolvedParser, resolvedConflicts)
		unresolved := counterexample.Find(unresolvedParser, unresolvedConflicts)
		Expect(unresolved).To(Equal(resolved))
	})

	DescribeTable("should find valid counterexamples on the tables of every core",
		func(spec string) {
			_, grammar, err := golrfrontend.GrammarFromString(spec)
			Expect(err).ToNot(HaveOccurred())
			for coreName, grammarToParser := range resolvedParsers {
				parser, conflicts, _, err := grammarToParser(grammar, conflict.DefaultPolicy)
				Expect(err).ToNot(HaveOccurred())
				expectValidCounterexamples(parser, conflicts, coreName)
			}
		},
		Entry("for figure 1", figure1Spec),
		Entry("for figure 3", figure3Spec),
		Entry("for figure 7", figure7Spec),
		Entry("for nonterminals which can vanish", emptySpec),
		Entry("for a reduction to the empty string", reduceReduceSpec),
		Entry("for states which LALR(1) merges", mergedReduceReduceSpec),
		Entry("for the end of the input", endOfInputSpec),
	)

	It("should only take the innermost nonterminal when the conflict terminal can follow the dot", func() {
		_, grammar, err := golrfrontend.GrammarFromString(blockedInnermostSpec)
		Expect(err).ToNot(HaveOccurred())
		for coreName, grammarToParser := range resolvedParsers {
			parser, conflicts, _, err := grammarToParser(grammar, conflict.DefaultPolicy)
			Expect(err).ToNot(HaveOccurred())
			// The searches give up early, so the innermost nonterminal is what they fall back to.
			expectValidCounterexamples(parser, conflicts, coreName, counterexample.WithTimeLimit(20*time.Millisecond))
		}
	})

	Context("well known grammars", func() {
		for _, wellKnownGrammar := range testdata.WellKnownGrammars {
			It("should find valid counterexamples for the conflicts of "+wellKnownGrammar.Title, func() {
				grammar, err := bisonfrontend.ToGrammar(
					bytes.NewBuffer(wellKnownGrammar.Content()),
					wellKnownGrammar.FileName,
				)
				Expect(err).ToNot(HaveOccurred())
				// Canonical LR(1) is left out, as its tables take too long to build for a unit test.
				for _, coreName := range []string{"ielr1", "lalr1"} {
					parser, conflicts, _, err := resolvedParsers[coreName](grammar, conflict.DefaultPolicy)
					Expect(err).ToNot(HaveOccurred())
					// Short time limits keep the test fast, the pairs whose search runs out get nonunifying ones.
					expectValidCounterexamples(parser, conflicts, coreName,
						counterexample.WithTimeLimit(10*time.Millisecond),
						counterexample.WithTotalTimeLimit(time.Second),
					)
				}
			})
		}
	})
})

// endOfInputSpec is an ambiguous grammar with a reduce/reduce conflict on the end of the input.
var endOfInputSpec = utils.HereDoc(`
	@scanner {
	    X: "x";
	}

	@parser {
	    s : "x" | "x" t ;
	    t : @empty ;
	}
`)

// findText returns the rendered counterexamples of the conflicts on the terminal which the core reports for the
// grammar, one line per element and an empty line between two counterexamples.
func findText(spec string, coreName string, terminalName string, options ...counterexample.Option) string {
	_, grammar, err := golrfrontend.GrammarFromString(spec)
	Expect(err).ToNot(HaveOccurred())
	parser, conflicts, _, err := resolvedParsers[coreName](grammar, conflict.DefaultPolicy)
	Expect(err).ToNot(HaveOccurred())
	counterexamplesByConflictIdx := counterexample.Find(parser, conflicts, options...)

	var texts []string
	for conflictIdx, c := range conflicts {
		if parser.Grammar.Terminals[c.TerminalIdx].String() != terminalName {
			continue
		}
		for _, ce := range counterexamplesByConflictIdx[conflictIdx] {
			texts = append(texts, strings.Join(ce.Lines(parser.Grammar), "\n")+"\n")
		}
	}
	return strings.Join(texts, "\n")
}

// expectValidCounterexamples checks that every conflict gets one counterexample per pair of its conflict items, with
// the items of the pair at the dots of the two derivations, and that every counterexample is valid, see
// expectValidCounterexample.
func expectValidCounterexamples(
	parser backend.Parser,
	conflicts []conflict.Conflict,
	coreName string,
	options ...counterexample.Option,
) {
	counterexamplesByConflictIdx := counterexample.Find(parser, conflicts, options...)
	Expect(counterexamplesByConflictIdx).To(HaveLen(len(conflicts)))
	tables := counterexample.NewLookupTables(parser)
	for conflictIdx, c := range conflicts {
		description := fmt.Sprintf("core %s, state %d, terminal %s",
			coreName, c.StateIdx, parser.Grammar.Terminals[c.TerminalIdx])
		var pairs [][2]backend.Core
		for _, ce := range counterexamplesByConflictIdx[conflictIdx] {
			pairs = append(pairs, [2]backend.Core{dotCore(ce.Derivations[0]), dotCore(ce.Derivations[1])})
			expectValidCounterexample(parser, &tables, ce, c, description)
		}
		Expect(pairs).To(ConsistOf(conflictItemPairs(parser, c)), description)
	}
}

// conflictItemPairs returns the pairs of conflict items of the conflict: every reduce item with every item which shifts
// the conflict terminal, and every two reduce items, the earlier production first.
func conflictItemPairs(parser backend.Parser, c conflict.Conflict) [][2]backend.Core {
	var shiftCores []backend.Core
	if c.Undeclared.Contains(conflict.NewShiftContribution()) {
		for _, core := range closure(parser.Grammar, parser.States[c.StateIdx].KernelItems) {
			symbolRefs := parser.Grammar.Productions[core.ProductionIdx()].SymbolRefs
			if core.Position() < len(symbolRefs) && symbolRefs[core.Position()] == frontend.NewTerminalRef(c.TerminalIdx) {
				shiftCores = append(shiftCores, core)
			}
		}
	}
	var reduceCores []backend.Core
	for _, contribution := range c.Undeclared.All() {
		if contribution.IsReduceAction() {
			productionIdx := contribution.ProductionIdx()
			reduceCores = append(reduceCores,
				backend.NewCore(productionIdx, len(parser.Grammar.Productions[productionIdx].SymbolRefs)))
		}
	}

	var result [][2]backend.Core
	for i, reduceCore := range reduceCores {
		for _, shiftCore := range shiftCores {
			result = append(result, [2]backend.Core{reduceCore, shiftCore})
		}
		for _, otherReduceCore := range reduceCores[i+1:] {
			result = append(result, [2]backend.Core{reduceCore, otherReduceCore})
		}
	}
	return result
}

// formatKernelItems returns the kernel items of the state, each as its production with the dot.
func formatKernelItems(parser backend.Parser, stateIdx int) []string {
	var result []string
	for _, core := range parser.States[stateIdx].KernelItems.All() {
		result = append(result, frontend.FormatItem(parser.Grammar, core.ProductionIdx(), core.Position()))
	}
	return result
}

// closure returns the kernel items and the closure items of a state.
func closure(grammar frontend.Grammar, kernelItems backend.CoreSet) []backend.Core {
	var result []backend.Core
	for _, core := range kernelItems.All() {
		result = append(result, core)
	}
	closureNonterminalIdxs := map[int]bool{}
	for i := 0; i < len(result); i++ {
		symbolRefs := grammar.Productions[result[i].ProductionIdx()].SymbolRefs
		if result[i].Position() == len(symbolRefs) || symbolRefs[result[i].Position()].IsTerminal() {
			continue
		}
		nonterminalIdx := symbolRefs[result[i].Position()].Idx()
		if closureNonterminalIdxs[nonterminalIdx] {
			continue
		}
		closureNonterminalIdxs[nonterminalIdx] = true
		for productionIdx, production := range grammar.Productions {
			if production.NonterminalIdx == nonterminalIdx {
				result = append(result, backend.NewCore(productionIdx, 0))
			}
		}
	}
	return result
}

// expectValidCounterexample checks that both derivations of the counterexample are derivations of its nonterminal with
// a single dot, which the parser tables can replay, see derivationReplay. The checks on the leaves follow in
// expectValidUnifying and expectValidNonunifying.
func expectValidCounterexample(
	parser backend.Parser,
	tables *counterexample.LookupTables,
	ce counterexample.Counterexample,
	c conflict.Conflict,
	description string,
) {
	grammar := parser.Grammar
	var leaves [2][]counterexample.Derivation
	var dotIdxs [2]int
	for i, derivation := range ce.Derivations {
		Expect(derivation.Symbol).To(Equal(frontend.NewNonterminalRef(ce.NonterminalIdx)), description)
		expectValidDerivation(grammar, derivation, description)

		leaves[i] = appendLeaves(derivation, nil)
		dotIdxs[i] = -1
		for leafIdx, leaf := range leaves[i] {
			if leaf.Dot {
				Expect(dotIdxs[i]).To(Equal(-1), "%s: more than one dot", description)
				dotIdxs[i] = leafIdx
			}
		}
		Expect(dotIdxs[i]).ToNot(Equal(-1), "%s: no dot", description)
	}
	Expect(replayFromAnyState(len(parser.States), tables, ce, c.StateIdx)).To(Succeed(), description)
	if ce.Unifying {
		expectValidUnifying(ce, leaves, dotIdxs[0], c.TerminalIdx, description)
	} else {
		expectValidNonunifying(grammar, ce, leaves, dotIdxs, c.TerminalIdx, description)
	}
}

// expectValidUnifying checks that the two derivations differ but have the same leaves. When a terminal follows the dot,
// it is the conflict terminal.
func expectValidUnifying(
	ce counterexample.Counterexample,
	leaves [2][]counterexample.Derivation,
	dotIdx int,
	terminalIdx int,
	description string,
) {
	Expect(ce.Derivations[0]).ToNot(Equal(ce.Derivations[1]), description)
	Expect(leaves[1]).To(Equal(leaves[0]), description)
	if dotIdx+1 < len(leaves[0]) && leaves[0][dotIdx+1].Symbol.IsTerminal() {
		Expect(leaves[0][dotIdx+1].Symbol).To(Equal(frontend.NewTerminalRef(terminalIdx)), description)
	}
}

// expectValidNonunifying checks that the conflict terminal directly follows the dot of both derivations, and that they
// share the symbols in front of the dot.
func expectValidNonunifying(
	grammar frontend.Grammar,
	ce counterexample.Counterexample,
	leaves [2][]counterexample.Derivation,
	dotIdxs [2]int,
	terminalIdx int,
	description string,
) {
	var prefixes [2][]frontend.SymbolRef
	for i := range leaves {
		Expect(dotIdxs[i]+1).To(BeNumerically("<", len(leaves[i])), "%s: nothing behind the dot", description)
		Expect(leaves[i][dotIdxs[i]+1].Symbol).To(Equal(frontend.NewTerminalRef(terminalIdx)), description)
		for _, leaf := range leaves[i][:dotIdxs[i]] {
			prefixes[i] = append(prefixes[i], leaf.Symbol)
		}
	}
	otherCore := dotCore(ce.Derivations[1])
	if otherCore.Position() < len(grammar.Productions[otherCore.ProductionIdx()].SymbolRefs) {
		// Only a second reduction can search a path of its own, see the spec for states which LALR(1) merges.
		Expect(prefixes[1]).To(Equal(prefixes[0]), description)
	}
}

// expectValidDerivation checks that every expanded node of the derivation has the symbols of its production as its
// children, besides the dot.
func expectValidDerivation(grammar frontend.Grammar, derivation counterexample.Derivation, description string) {
	if !derivation.IsExpanded() {
		return
	}
	production := grammar.Productions[derivation.ProductionIdx]
	Expect(derivation.Symbol).To(Equal(frontend.NewNonterminalRef(production.NonterminalIdx)), description)
	var symbolRefs []frontend.SymbolRef
	for _, child := range derivation.Children {
		if !child.Dot {
			symbolRefs = append(symbolRefs, child.Symbol)
		}
		expectValidDerivation(grammar, child, description)
	}
	Expect(symbolRefs).To(HaveExactElements(production.SymbolRefs), description)
}

// appendLeaves appends the leaves of the derivation, including the dot.
func appendLeaves(
	derivation counterexample.Derivation,
	leaves []counterexample.Derivation,
) []counterexample.Derivation {
	if !derivation.IsExpanded() {
		return append(leaves, derivation)
	}
	for _, child := range derivation.Children {
		leaves = appendLeaves(child, leaves)
	}
	return leaves
}

// dotCore returns the item at the dot of the derivation: the production the dot is a child of, with the dot at its
// position.
func dotCore(derivation counterexample.Derivation) backend.Core {
	core, found := findDotCore(derivation)
	Expect(found).To(BeTrue(), "no dot")
	return core
}

// findDotCore returns the item at the dot of the derivation, see dotCore. It reports false when the derivation has no
// dot.
func findDotCore(derivation counterexample.Derivation) (backend.Core, bool) {
	for childIdx, child := range derivation.Children {
		if child.Dot {
			return backend.NewCore(derivation.ProductionIdx, childIdx), true
		}
		if core, found := findDotCore(child); found {
			return core, true
		}
	}
	return 0, false
}

// replayFromAnyState returns an error when no state of the parser tables replays both derivations of the
// counterexample, see derivationReplay.
func replayFromAnyState(
	stateCount int,
	tables *counterexample.LookupTables,
	ce counterexample.Counterexample,
	conflictStateIdx int,
) error {
	var err error
	for startStateIdx := range stateCount {
		err = nil
		for _, derivation := range ce.Derivations {
			replay := newDerivationReplay(tables, ce, derivation, conflictStateIdx, startStateIdx)
			if err = replay.Replay(derivation); err != nil {
				break
			}
		}
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("no state replays both derivations, the last state failed with: %w", err)
}

// derivationReplay runs a derivation through the parser tables the way the parser would, from a start state: a leaf
// takes the transition of the item in front of it, the dot has to be in the conflict state, and an expanded node
// reduces its production when its children are done. Every shift, and every reduction which a terminal follows, has to
// be an action the declarations of the grammar leave to the parser. This ties the counterexample to the state of its
// conflict, and makes sure the search never took an action a precedence declaration removed.
//
// The replay of a nonunifying counterexample ends with the conflict terminal. Its derivations only complete the
// productions behind it, which the parser does not have to accept.
type derivationReplay struct {
	tables           *counterexample.LookupTables
	leaves           []counterexample.Derivation
	leafIdx          int
	lastLeafIdx      int
	done             bool
	conflictStateIdx int
	stateIdxs        []int
}

// newDerivationReplay returns a replay of the derivation of the counterexample, beginning in the start state.
func newDerivationReplay(
	tables *counterexample.LookupTables,
	ce counterexample.Counterexample,
	derivation counterexample.Derivation,
	conflictStateIdx int,
	startStateIdx int,
) derivationReplay {
	leaves := appendLeaves(derivation, nil)
	lastLeafIdx := len(leaves) - 1
	if !ce.Unifying {
		lastLeafIdx = slices.IndexFunc(leaves, isDot) + 1
	}
	return derivationReplay{
		tables:           tables,
		leaves:           leaves,
		lastLeafIdx:      lastLeafIdx,
		conflictStateIdx: conflictStateIdx,
		stateIdxs:        []int{startStateIdx},
	}
}

// isDot reports if the node is the dot.
func isDot(derivation counterexample.Derivation) bool {
	return derivation.Dot
}

// Replay replays the expanded node and reduces its production, leaving the state the node began in on top.
func (r *derivationReplay) Replay(derivation counterexample.Derivation) error {
	position := 0
	for _, child := range derivation.Children {
		if r.done {
			return nil
		}
		stateIdx := r.stateIdxs[len(r.stateIdxs)-1]
		if child.Dot {
			r.leafIdx++
			if stateIdx != r.conflictStateIdx {
				return fmt.Errorf("the dot is in state %d", stateIdx)
			}
			continue
		}
		itemIdx, found := r.tables.ItemIdx(stateIdx, backend.NewCore(derivation.ProductionIdx, position))
		if !found {
			return fmt.Errorf("state %d has no item of production %d at %d", stateIdx, derivation.ProductionIdx, position)
		}
		if child.IsExpanded() {
			if err := r.Replay(child); err != nil || r.done {
				return err
			}
		} else {
			r.leafIdx++
			if child.Symbol.IsTerminal() &&
				!r.tables.IsActionAllowed(stateIdx, child.Symbol.Idx(), conflict.NewShiftContribution()) {
				return fmt.Errorf("state %d may not shift terminal %d", stateIdx, child.Symbol.Idx())
			}
		}
		targetItemIdx, found := r.tables.Transition(itemIdx)
		if !found {
			return fmt.Errorf("item %d has no transition", itemIdx)
		}
		r.stateIdxs = append(r.stateIdxs, r.tables.StateIdx(targetItemIdx))
		position++
		r.done = r.leafIdx > r.lastLeafIdx
	}
	if r.done {
		return nil
	}

	stateIdx := r.stateIdxs[len(r.stateIdxs)-1]
	if terminalIdx, found := r.nextTerminal(); found &&
		!r.tables.IsActionAllowed(stateIdx, terminalIdx, conflict.NewReduceContribution(derivation.ProductionIdx)) {
		return fmt.Errorf("state %d may not reduce production %d on terminal %d",
			stateIdx, derivation.ProductionIdx, terminalIdx)
	}
	r.stateIdxs = r.stateIdxs[:len(r.stateIdxs)-position]
	return nil
}

// nextTerminal returns the leaf behind the leaves replayed so far, besides the dot, when it is a terminal.
func (r *derivationReplay) nextTerminal() (int, bool) {
	for _, leaf := range r.leaves[r.leafIdx:] {
		if leaf.Dot {
			continue
		}
		return leaf.Symbol.Idx(), leaf.Symbol.IsTerminal()
	}
	return 0, false
}
