package golr_test

import (
	"testing"

	"github.com/backbone81/golr/internal/parsergen/core/ielr1/golr/oracle"
	lr1golrcore "github.com/backbone81/golr/internal/parsergen/core/lr1/golr"
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
})

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
