package golr_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	ielr1golr "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	lalr1golr "github.com/backbone81/golr/internal/parsergen/core/lalr1/golr"
	lr1golr "github.com/backbone81/golr/internal/parsergen/core/lr1/golr"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

var _ = Describe("LR(1) Builder", func() {
	DescribeTable("should correctly compute the canonical LR(1) parser table",
		func(grammar frontend.Grammar, wantLR1Parser backend.Parser) {
			lr1Builder := lr1golr.NewLR1Builder(frontend.AugmentGrammar(grammar))
			Expect(lr1Builder.Build()).To(Succeed())
			Expect(lr1Builder.Parser()).To(Equal(wantLR1Parser))
		},
		Entry(
			"the unambiguous test grammar for Fig. 1",
			ielr1golr.UnambiguousTestGrammarFig1,
			ielr1golr.UnambiguousTestGrammarFig1LR1Parser,
		),
	)

	DescribeTable("should produce the correct number of LR(1) states",
		func(grammar frontend.Grammar, wantStateCount int) {
			lr1Builder := lr1golr.NewLR1Builder(frontend.AugmentGrammar(grammar))
			Expect(lr1Builder.Build()).To(Succeed())

			Expect(lr1Builder.Parser().States).To(HaveLen(wantStateCount))
		},
		// The state counts were cross-checked against GNU Bison 3.8.2 (--define=lr.type=canonical-lr).
		Entry("the unambiguous test grammar for Fig. 1", ielr1golr.UnambiguousTestGrammarFig1, 13),
		Entry("the goto follows test grammar for Fig. 5", ielr1golr.GotoFollowsTestGrammarFig5, 27),
		Entry("the goto follows caveats test grammar for Fig. 6", ielr1golr.GotoFollowsCaveatsTestGrammarFig6, 18),
		Entry("the LR(1) but not LALR(1) grammar", ielr1golr.ReduceReduceConflictTestGrammar, 15),
	)

	// The two tests below are about which conflicts a table is left with, so they go through
	// GrammarToUnresolvedParser: the conflicts are what GrammarToParser would have resolved away.
	It("should not report a conflict for a grammar which is LR(1) but not LALR(1)", func() {
		// This is the whole point of canonical LR(1) as an oracle: the reduce/reduce conflict which LALR(1) reports for
		// this grammar is an artifact of merging the two "c" states, not a property of the grammar.
		lalr1Parser, err := lalr1golr.GrammarToUnresolvedParser(
			ielr1golr.ReduceReduceConflictTestGrammar,
			conflict.DefaultPolicy,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(conflict.HasConflict(lalr1Parser)).To(BeTrue())

		lr1Parser, err := lr1golr.GrammarToUnresolvedParser(
			ielr1golr.ReduceReduceConflictTestGrammar,
			conflict.DefaultPolicy,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(conflict.HasConflict(lr1Parser)).To(BeFalse())
	})

	It("should report a conflict for an ambiguous grammar", func() {
		lr1Parser, err := lr1golr.GrammarToUnresolvedParser(ielr1golr.AmbiguousTestGrammarFig2, conflict.DefaultPolicy)
		Expect(err).ToNot(HaveOccurred())
		Expect(conflict.HasConflict(lr1Parser)).To(BeTrue())
	})
})

func BenchmarkComputeLR1ParserTables(b *testing.B) {
	benchmarks := []struct {
		description string
		grammar     frontend.Grammar
	}{
		{"Goto Follows Caveats Test Grammar Fig6", ielr1golr.GotoFollowsCaveatsTestGrammarFig6},
		{"Unambiguous Test Grammar Fig1", ielr1golr.UnambiguousTestGrammarFig1},
		{"Ambiguous Test Grammar Fig2", ielr1golr.AmbiguousTestGrammarFig2},
		{"Goto Follows Test Grammar Fig5", ielr1golr.GotoFollowsTestGrammarFig5},
		{"Reduce/Reduce Conflict Test Grammar", ielr1golr.ReduceReduceConflictTestGrammar},
	}
	for _, benchmark := range benchmarks {
		b.Run(benchmark.description, func(b *testing.B) {
			augmentedGrammar := frontend.AugmentGrammar(benchmark.grammar)
			for b.Loop() {
				lr1Builder := lr1golr.NewLR1Builder(augmentedGrammar)
				if err := lr1Builder.Build(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
