package counterexample_test

import (
	"bytes"
	"fmt"
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
	policy := conflict.DefaultPolicy(parser.Grammar)
	counterexamplesByConflictIdx := counterexample.Find(parser, conflicts, policy, options...)

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
func expectValidCounterexamples(
	parser backend.Parser,
	conflicts []conflict.Conflict,
	coreName string,
	options ...counterexample.Option,
) {
	policy := conflict.DefaultPolicy(parser.Grammar)
	counterexamplesByConflictIdx := counterexample.Find(parser, conflicts, policy, options...)
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

// expectValidCounterexample checks that both derivations of the counterexample are derivations of its nonterminal with
// a single dot, and that the first one uses a reduction, see expectValidUnifying and expectValidNonunifying.
func expectValidCounterexample(
	grammar frontend.Grammar,
	ce counterexample.Counterexample,
	terminalIdx int,
	description string,
) {
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
	Expect(firstProductionEndsWithDot(ce.Derivations[0])).To(BeTrue(), description)
	if ce.Unifying {
		expectValidUnifying(ce, leaves, dotIdxs[0], terminalIdx, description)
	} else {
		expectValidNonunifying(ce, leaves, dotIdxs, terminalIdx, description)
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
