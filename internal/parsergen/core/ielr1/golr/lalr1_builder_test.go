package golr_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
	golr2 "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

var _ = Describe("LALR(1) Builder", func() {
	DescribeTable("should correctly compute the LALR(1) parser table",
		func(grammar frontend.Grammar, wantLALR1Parser backend.Parser) {
			augmentedGrammar := frontend.AugmentGrammar(grammar)
			lalr1Builder := golr2.NewLALR1Builder(augmentedGrammar)
			lalr1Builder.Build()
			Expect(lalr1Builder.Parser()).To(Equal(wantLALR1Parser))
		},
		Entry(
			"the unambiguous test grammar for Fig. 1",
			golr2.UnambiguousTestGrammarFig1,
			golr2.UnambiguousTestGrammarFig1LALR1Parser,
		),
		Entry(
			"the ambiguous test grammar for Fig. 2",
			golr2.AmbiguousTestGrammarFig2,
			golr2.AmbiguousTestGrammarFig2LALR1Parser,
		),
		Entry(
			"the goto follows test grammar for Fig. 5",
			golr2.GotoFollowsTestGrammarFig5,
			golr2.GotoFollowsTestGrammarFig5LALR1Parser,
		),
		Entry(
			"the goto follows caveats test grammar for Fig. 6",
			golr2.GotoFollowsCaveatsTestGrammarFig6,
			golr2.GotoFollowsCaveatsTestGrammarFig6LALRParser,
		),
		Entry(
			"the LR(1) but not LALR(1) grammar with a reduce/reduce conflict",
			golr2.ReduceReduceConflictTestGrammar,
			golr2.ReduceReduceConflictTestGrammarLALR1Parser,
		),
	)
})

func BenchmarkComputeLALR1ParserTables(b *testing.B) {
	benchmarks := []struct {
		description string
		grammar     frontend.Grammar
	}{
		{"Goto Follows Caveats Test Grammar Fig6", golr2.GotoFollowsCaveatsTestGrammarFig6},
		{"Unambiguous Test Grammar Fig1", golr2.UnambiguousTestGrammarFig1},
		{"Ambiguous Test Grammar Fig2", golr2.AmbiguousTestGrammarFig2},
		{"Goto Follows Test Grammar Fig5", golr2.GotoFollowsTestGrammarFig5},
		{"Reduce/Reduce Conflict Test Grammar", golr2.ReduceReduceConflictTestGrammar},
	}
	for _, benchmark := range benchmarks {
		b.Run(benchmark.description, func(b *testing.B) {
			augmentedGrammar := frontend.AugmentGrammar(benchmark.grammar)
			for range b.N {
				lalr1Builder := golr2.NewLALR1Builder(augmentedGrammar)
				lalr1Builder.Build()
			}
		})
	}
}
