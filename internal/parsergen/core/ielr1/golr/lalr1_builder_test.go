package golr_test

import (
	"bytes"
	"maps"
	"math/rand"
	"slices"
	"testing"

	"github.com/backbone81/golr/internal/parsergen/core/ielr1/golr/oracle"
	lr1golrcore "github.com/backbone81/golr/internal/parsergen/core/lr1/golr"
	bisonfrontend "github.com/backbone81/golr/internal/parsergen/frontend/bison"
	lalr1bisoncore "github.com/backbone81/golr/pkg/parsergen/core/lalr1/bison"
	"github.com/backbone81/golr/testdata"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
	ielr1golrcore "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

var _ = Describe("LALR(1) Builder", func() {
	DescribeTable("should correctly compute the LALR(1) parser table",
		func(grammar frontend.Grammar, wantLALR1Parser backend.Parser) {
			augmentedGrammar := frontend.AugmentGrammar(grammar)
			lalr1Builder := ielr1golrcore.NewLALR1Builder(augmentedGrammar)
			lalr1Builder.Build()
			Expect(lalr1Builder.Parser()).To(Equal(wantLALR1Parser))
		},
		Entry(
			"the unambiguous test grammar for Fig. 1",
			ielr1golrcore.UnambiguousTestGrammarFig1,
			ielr1golrcore.UnambiguousTestGrammarFig1LALR1Parser,
		),
		Entry(
			"the ambiguous test grammar for Fig. 2",
			ielr1golrcore.AmbiguousTestGrammarFig2,
			ielr1golrcore.AmbiguousTestGrammarFig2LALR1Parser,
		),
		Entry(
			"the goto follows test grammar for Fig. 5",
			ielr1golrcore.GotoFollowsTestGrammarFig5,
			ielr1golrcore.GotoFollowsTestGrammarFig5LALR1Parser,
		),
		Entry(
			"the goto follows caveats test grammar for Fig. 6",
			ielr1golrcore.GotoFollowsCaveatsTestGrammarFig6,
			ielr1golrcore.GotoFollowsCaveatsTestGrammarFig6LALRParser,
		),
		Entry(
			"the LR(1) but not LALR(1) grammar with a reduce/reduce conflict",
			ielr1golrcore.ReduceReduceConflictTestGrammar,
			ielr1golrcore.ReduceReduceConflictTestGrammarLALR1Parser,
		),
	)

	// We run the LALR(1) builder implementation against an alternative implementation which is our oracle. If both
	// implementations disagree on the outcome, either one is wrong.
	// The LALR(1) builder derives its reduction lookaheads by propagating goto follows along the digraph relations of
	// DeRemer and Pennello. The oracle arrives at the same parser table on a completely different route: it builds the
	// canonical LR(1) automaton, which needs nothing but item closures and first sets, and then merges the states which
	// agree on their cores.
	Context("Oracle comparison tests", func() {
		DescribeTable(
			"should agree with the LALR(1) parser table obtained by merging isocores of canonical LR(1)",
			func(grammar frontend.Grammar) {
				augmentedGrammar := frontend.AugmentGrammar(grammar)

				lalr1Builder := ielr1golrcore.NewLALR1Builder(augmentedGrammar)
				lalr1Builder.Build()
				gotParser := lalr1Builder.Parser()

				lr1Builder := lr1golrcore.NewLR1Builder(augmentedGrammar)
				Expect(lr1Builder.Build()).To(Succeed())
				wantParser, err := oracle.LALR1FromLR1(lr1Builder.Parser())
				Expect(err).ToNot(HaveOccurred())

				// The two automatons have to agree before it is worth comparing any lookahead set. A difference here
				// means the states themselves are wrong, and every lookahead difference which follows from it would
				// only be noise.
				Expect(oracle.DiffLALR1ParserKernelItems(wantParser, gotParser)).To(BeEmpty())
				Expect(oracle.DiffLALR1ParserStates(wantParser, gotParser)).To(BeEmpty())
			},
			Entry("the unambiguous test grammar for Fig. 1", ielr1golrcore.UnambiguousTestGrammarFig1),
			Entry("the ambiguous test grammar for Fig. 2", ielr1golrcore.AmbiguousTestGrammarFig2),
			Entry("the goto follows test grammar for Fig. 5", ielr1golrcore.GotoFollowsTestGrammarFig5),
			Entry("the goto follows caveats test grammar for Fig. 6", ielr1golrcore.GotoFollowsCaveatsTestGrammarFig6),
			Entry("the LR(1) but not LALR(1) grammar", ielr1golrcore.ReduceReduceConflictTestGrammar),
		)
	})

	// We feed a large corpus of random grammars through both the LALR(1) builder and the oracle and assert they agree
	// on every one. Where the hand-picked grammars above pin down specific figures of the papers, the random corpus is
	// there to surprise us: it reaches grammar shapes we did not think to write down. Each run draws a fresh corpus from
	// the Ginkgo random seed, so a rare edge case surfaces over repeated runs rather than being hidden by a fixed corpus;
	// a failing run replays with `ginkgo --seed=...`, and the reported per-grammar seed reconstructs the single failing
	// grammar on its own.
	Context("Random grammar oracle comparison", func() {
		It("should agree with the oracle on a corpus of random grammars", func() {
			// grammarCount trades test time for how much of the grammar space is explored. The corpus runs under -race,
			// so keep it to a size which still finishes in a couple of seconds; bump it when hunting a suspected bug.
			const grammarCount = 2000

			var compared, withNullable, withMerging, withConflict int

			// scenarioGrammarCounts records in how many grammars each scenario fired at least once, and scenarioFireCounts
			// the total number of productions it built across the corpus. A fresh generator's ObservedScenarioCounts holds
			// only the current grammar's scenarios, so every key it has after Generate is a scenario that grammar used.
			// These aggregates are reported below to make the scenario distribution visible when the weights need tuning.
			scenarioGrammarCounts := map[oracle.ProductionScenario]int{}
			scenarioFireCounts := map[oracle.ProductionScenario]int{}

			// The corpus is seeded from the Ginkgo random seed, so every run explores a different set of grammars and a
			// rare edge case surfaces eventually rather than being masked by a fixed corpus, while a failing run stays
			// reproducible with `ginkgo --seed=...`. A master RNG derives a distinct seed for every grammar, so the corpus
			// is grammarCount different grammars instead of the same grammar repeated. The derived seed is reported with a
			// failure so a single failing grammar can be reconstructed directly, which is what shrinking it into a
			// regression fixture needs.
			masterRng := rand.New(rand.NewSource(GinkgoRandomSeed()))
			for range grammarCount {
				grammarSeed := masterRng.Int63()
				generator := oracle.DefaultGrammarGenerator(rand.New(rand.NewSource(grammarSeed)))
				grammar := generator.Generate()
				for scenario, count := range generator.ObservedScenarioCounts {
					scenarioGrammarCounts[scenario]++
					scenarioFireCounts[scenario] += count
				}
				augmentedGrammar := frontend.AugmentGrammar(grammar)

				lr1Builder := lr1golrcore.NewLR1Builder(augmentedGrammar)
				if err := lr1Builder.Build(); err != nil {
					// A grammar whose canonical LR(1) automaton exceeds the addressable state limit cannot be handled by
					// the oracle. It is skipped, not a failure of the builder under test.
					Expect(err).To(MatchError(lr1golrcore.ErrStateLimitExceeded), "grammar seed %d:\n%s", grammarSeed, grammar.String())
					continue
				}
				lr1Parser := lr1Builder.Parser()

				wantParser, err := oracle.LALR1FromLR1(lr1Parser)
				Expect(err).ToNot(HaveOccurred(), "grammar seed %d:\n%s", grammarSeed, grammar.String())

				lalr1Builder := ielr1golrcore.NewLALR1Builder(augmentedGrammar)
				lalr1Builder.Build()
				gotParser := lalr1Builder.Parser()

				Expect(oracle.DiffLALR1ParserKernelItems(wantParser, gotParser)).To(
					BeEmpty(), "grammar seed %d:\n%s", grammarSeed, grammar.String(),
				)
				Expect(oracle.DiffLALR1ParserStates(wantParser, gotParser)).To(
					BeEmpty(), "grammar seed %d:\n%s", grammarSeed, grammar.String(),
				)

				compared++
				if hasEmptyProduction(grammar) {
					withNullable++
				}
				if len(lr1Parser.States) > len(gotParser.States) {
					// The canonical LR(1) automaton has more states than the LALR(1) one, so merging isocores actually
					// collapsed states. These are the grammars which exercise the lookahead merging the oracle and the
					// builder can disagree on, so we track how many of them the corpus reaches.
					withMerging++
				}
				if hasConflict(gotParser) {
					// A conflict is where a wrong lookahead set actually changes the parser, so grammars whose LALR(1)
					// table has a conflict are the sharpest test of the lookahead computation. We track them to make
					// sure the corpus keeps reaching them.
					withConflict++
				}
			}

			GinkgoWriter.Printf(
				"random grammar corpus: %d compared, %d with nullable nonterminals, %d with isocore merging, %d with a conflict\n",
				compared, withNullable, withMerging, withConflict,
			)

			// Report the scenario distribution so a future weight adjustment can be checked against the fraction of
			// grammars each scenario actually reaches, rather than against the raw weights: because a cluster scenario is
			// dropped once it no longer fits the nonterminal budget, the reached fraction can differ markedly from the
			// weight.
			GinkgoWriter.Printf("scenario distribution over %d generated grammars (grammars using the scenario, total fires):\n", grammarCount)
			for _, scenario := range slices.Sorted(maps.Keys(scenarioGrammarCounts)) {
				grammars := scenarioGrammarCounts[scenario]
				GinkgoWriter.Printf(
					"  %-32s %5.1f%% (%d grammars, %d fires)\n",
					scenario, 100*float64(grammars)/float64(grammarCount), grammars, scenarioFireCounts[scenario],
				)
			}

			// Guard the discriminating power of the corpus: if the generator ever degrades into trivial grammars, these
			// expectations fail even though every comparison still passes, which is the failure we care about the most.
			Expect(compared).To(BeNumerically(">", grammarCount/2))
			Expect(withNullable).To(BeNumerically(">", 100))
			Expect(withMerging).To(BeNumerically(">", 100))
			Expect(withConflict).To(BeNumerically(">", 100))
		})
	})

	Context("well known grammars", func() {
		for _, wellKnownGrammar := range testdata.WellKnownGrammars {
			It("should correctly build the "+wellKnownGrammar.Title+" parser", func() {
				grammar, err := bisonfrontend.ToGrammar(
					bytes.NewBuffer(wellKnownGrammar.Content),
					wellKnownGrammar.FileName,
				)
				Expect(err).ToNot(HaveOccurred())

				bisonParser, _, err := lalr1bisoncore.GrammarToParser(grammar)
				Expect(err).ToNot(HaveOccurred())

				augmentedGrammar := frontend.AugmentGrammar(grammar)
				lalr1Builder := ielr1golrcore.NewLALR1Builder(augmentedGrammar)
				lalr1Builder.Build()
				golrParser := lalr1Builder.Parser()

				Expect(len(golrParser.States)).To(Equal(len(bisonParser.States)))
			})
		}
	})
})

// hasEmptyProduction reports if the grammar has a production with an empty right hand side, which is what makes a
// nonterminal nullable.
func hasEmptyProduction(grammar frontend.Grammar) bool {
	for _, production := range grammar.Productions {
		if len(production.SymbolRefs) == 0 {
			return true
		}
	}
	return false
}

// hasConflict reports if any state of the parser table has a shift/reduce or reduce/reduce conflict: a reduce action
// whose lookahead set contains a terminal the state also shifts on, or which a reduce action seen earlier in the same
// state also reduces on.
func hasConflict(parser backend.Parser) bool {
	for stateIdx := range parser.States {
		state := &parser.States[stateIdx]

		shiftTerminals := make(map[int]bool)
		for _, transitionAction := range state.TransitionActions.All() {
			if transitionAction.SymbolRef().IsTerminal() {
				shiftTerminals[transitionAction.SymbolRef().Idx()] = true
			}
		}

		var claimedTerminals backend.LookaheadSet
		for _, reduceAction := range state.ReduceActions.All() {
			for terminalIdx := range reduceAction.LookaheadSet.All() {
				if shiftTerminals[terminalIdx] || claimedTerminals.Contains(terminalIdx) {
					return true
				}
			}
			claimedTerminals.Merge(&reduceAction.LookaheadSet)
		}
	}
	return false
}

func BenchmarkComputeLALR1ParserTables(b *testing.B) {
	benchmarks := []struct {
		description string
		grammar     frontend.Grammar
	}{
		{"Goto Follows Caveats Test Grammar Fig6", ielr1golrcore.GotoFollowsCaveatsTestGrammarFig6},
		{"Unambiguous Test Grammar Fig1", ielr1golrcore.UnambiguousTestGrammarFig1},
		{"Ambiguous Test Grammar Fig2", ielr1golrcore.AmbiguousTestGrammarFig2},
		{"Goto Follows Test Grammar Fig5", ielr1golrcore.GotoFollowsTestGrammarFig5},
		{"Reduce/Reduce Conflict Test Grammar", ielr1golrcore.ReduceReduceConflictTestGrammar},
	}
	for _, benchmark := range benchmarks {
		b.Run(benchmark.description, func(b *testing.B) {
			augmentedGrammar := frontend.AugmentGrammar(benchmark.grammar)
			for range b.N {
				lalr1Builder := ielr1golrcore.NewLALR1Builder(augmentedGrammar)
				lalr1Builder.Build()
			}
		})
	}
}

// BenchmarkGenerateRandomGrammar measures the cost of generating a single random grammar. The grammar corpus of the
// oracle comparison test builds on this, so it is worth knowing how much of the corpus run time is generation and how
// much is the comparison which follows. The seed is varied per iteration so the benchmark covers a spread of grammars
// rather than repeating one.
func BenchmarkGenerateRandomGrammar(b *testing.B) {
	for i := range b.N {
		oracle.DefaultGrammarGenerator(rand.New(rand.NewSource(int64(i)))).Generate()
	}
}

// BenchmarkRandomGrammarOracleComparison measures the cost of one full step of the oracle comparison test: generate a
// random grammar, build both the LALR(1) table and the canonical LR(1) table it is merged from, and diff the two. This
// is the benchmark to watch when tuning either the builder or the generator, as it reflects the actual corpus run time
// per grammar. Grammars the oracle cannot handle are skipped, the same way the test skips them.
func BenchmarkRandomGrammarOracleComparison(b *testing.B) {
	for i := range b.N {
		grammar := oracle.DefaultGrammarGenerator(rand.New(rand.NewSource(int64(i)))).Generate()
		augmentedGrammar := frontend.AugmentGrammar(grammar)

		lr1Builder := lr1golrcore.NewLR1Builder(augmentedGrammar)
		if err := lr1Builder.Build(); err != nil {
			continue
		}
		wantParser, err := oracle.LALR1FromLR1(lr1Builder.Parser())
		if err != nil {
			continue
		}

		lalr1Builder := ielr1golrcore.NewLALR1Builder(augmentedGrammar)
		lalr1Builder.Build()
		gotParser := lalr1Builder.Parser()

		oracle.DiffLALR1ParserKernelItems(wantParser, gotParser)
		oracle.DiffLALR1ParserStates(wantParser, gotParser)
	}
}
