package counterexample_test

import (
	"bytes"
	"fmt"
	"strings"

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
	It("should find the nonunifying counterexample of the challenging conflict", func() {
		Expect(findText(figure1Spec, "ielr1", "DIGIT")).To(Equal(utils.HereDoc(`
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
		Expect(findText(figure1Spec, "ielr1", `"else"`)).To(Equal(utils.HereDoc(`
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
		Expect(findText(figure3Spec, "ielr1", `"a"`)).To(Equal(utils.HereDoc(`
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
		Expect(findText(figure7Spec, "ielr1", `"b"`)).To(Equal(utils.HereDoc(`
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
		Expect(findText(reduceReduceSpec, "ielr1", `"z"`)).To(Equal(utils.HereDoc(`
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
		Expect(findText(emptySpec, "ielr1", `"p"`)).To(Equal(utils.HereDoc(`
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
		Expect(findText(mergedReduceReduceSpec, "lalr1", `"d"`)).To(Equal(utils.HereDoc(`
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
		Expect(findText(endOfInputSpec, "ielr1", "$end")).To(Equal(utils.HereDoc(`
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

	It("should find the same counterexamples on tables whose conflicts are left unresolved", func() {
		_, grammar, err := golrfrontend.GrammarFromString(figure1Spec)
		Expect(err).ToNot(HaveOccurred())
		resolvedParser, resolvedConflicts, _, err := lalr1golr.GrammarToParser(grammar, conflict.DefaultPolicy)
		Expect(err).ToNot(HaveOccurred())
		failingPolicy := conflict.SelectPolicy(true, true)
		unresolvedParser, unresolvedConflicts, _, err := lalr1golr.GrammarToParser(grammar, failingPolicy)
		Expect(err).To(HaveOccurred())

		resolved := counterexample.Find(resolvedParser, resolvedConflicts, conflict.DefaultPolicy(resolvedParser.Grammar))
		unresolved := counterexample.Find(unresolvedParser, unresolvedConflicts, failingPolicy(unresolvedParser.Grammar))
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
					expectValidCounterexamples(parser, conflicts, coreName)
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
func findText(spec string, coreName string, terminalName string) string {
	_, grammar, err := golrfrontend.GrammarFromString(spec)
	Expect(err).ToNot(HaveOccurred())
	parser, conflicts, _, err := resolvedParsers[coreName](grammar, conflict.DefaultPolicy)
	Expect(err).ToNot(HaveOccurred())
	counterexamplesByConflictIdx := counterexample.Find(parser, conflicts, conflict.DefaultPolicy(parser.Grammar))

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

// expectValidCounterexamples checks that every conflict gets one counterexample per pair of conflict items, and that
// every counterexample is valid, see expectValidCounterexample.
func expectValidCounterexamples(parser backend.Parser, conflicts []conflict.Conflict, coreName string) {
	counterexamplesByConflictIdx := counterexample.Find(parser, conflicts, conflict.DefaultPolicy(parser.Grammar))
	Expect(counterexamplesByConflictIdx).To(HaveLen(len(conflicts)))
	for conflictIdx, c := range conflicts {
		description := fmt.Sprintf("core %s, state %d, terminal %s",
			coreName, c.StateIdx, parser.Grammar.Terminals[c.TerminalIdx])
		Expect(counterexamplesByConflictIdx[conflictIdx]).To(HaveLen(pairCount(parser, c)), description)
		for _, ce := range counterexamplesByConflictIdx[conflictIdx] {
			expectValidCounterexample(parser.Grammar, ce, c.TerminalIdx, description)
		}
	}
}

// pairCount returns the number of pairs of conflict items of the conflict.
func pairCount(parser backend.Parser, c conflict.Conflict) int {
	shiftItemCount := 0
	if c.Undeclared.Contains(conflict.NewShiftContribution()) {
		for _, core := range closure(parser.Grammar, parser.States[c.StateIdx].KernelItems) {
			symbolRefs := parser.Grammar.Productions[core.ProductionIdx()].SymbolRefs
			if core.Position() < len(symbolRefs) && symbolRefs[core.Position()] == frontend.NewTerminalRef(c.TerminalIdx) {
				shiftItemCount++
			}
		}
	}
	reduceCount := 0
	for _, contribution := range c.Undeclared.All() {
		if contribution.IsReduceAction() {
			reduceCount++
		}
	}
	return reduceCount*shiftItemCount + reduceCount*(reduceCount-1)/2
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

// expectValidCounterexample checks that both derivations of the nonunifying counterexample are derivations of its
// nonterminal with a single dot, which the terminal follows directly, and that they share the symbols in front of the
// dot. The first derivation uses a reduction.
func expectValidCounterexample(
	grammar frontend.Grammar,
	ce counterexample.Counterexample,
	terminalIdx int,
	description string,
) {
	Expect(ce.Unifying).To(BeFalse(), description)
	var prefixes [2][]frontend.SymbolRef
	for i, derivation := range ce.Derivations {
		Expect(derivation.Symbol).To(Equal(frontend.NewNonterminalRef(ce.NonterminalIdx)), description)
		expectValidDerivation(grammar, derivation, description)

		leaves := appendLeaves(derivation, nil)
		dotIdx := -1
		for leafIdx, leaf := range leaves {
			if leaf.Dot {
				Expect(dotIdx).To(Equal(-1), "%s: more than one dot", description)
				dotIdx = leafIdx
			}
		}
		Expect(dotIdx).ToNot(Equal(-1), "%s: no dot", description)
		Expect(dotIdx+1).To(BeNumerically("<", len(leaves)), "%s: nothing behind the dot", description)
		Expect(leaves[dotIdx+1].Symbol).To(Equal(frontend.NewTerminalRef(terminalIdx)), description)
		for _, leaf := range leaves[:dotIdx] {
			prefixes[i] = append(prefixes[i], leaf.Symbol)
		}
	}
	Expect(firstProductionEndsWithDot(ce.Derivations[0])).To(BeTrue(), description)
	if !firstProductionEndsWithDot(ce.Derivations[1]) {
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
	Expect(symbolRefs).To(Equal(production.SymbolRefs), description)
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

// firstProductionEndsWithDot reports if the production the dot is a child of has the dot at its end.
func firstProductionEndsWithDot(derivation counterexample.Derivation) bool {
	for childIdx, child := range derivation.Children {
		if child.Dot {
			return childIdx == len(derivation.Children)-1
		}
		if len(appendDots(child, nil)) > 0 {
			return firstProductionEndsWithDot(child)
		}
	}
	return false
}

// appendDots appends the dots of the derivation.
func appendDots(derivation counterexample.Derivation, dots []counterexample.Derivation) []counterexample.Derivation {
	if derivation.Dot {
		return append(dots, derivation)
	}
	for _, child := range derivation.Children {
		dots = appendDots(child, dots)
	}
	return dots
}
