package ielr1go_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/core/ielr1go"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

var _ = Describe("LALR(1) Builder", func() {
	DescribeTable("should correctly compute the LALR(1) parser table",
		func(grammar frontend.Grammar, wantLALR1Parser backend.Parser) {
			augmentedGrammar := frontend.AugmentGrammar(grammar)
			lalr1Builder := ielr1go.NewLALR1Builder(augmentedGrammar)
			lalr1Builder.Build()
			Expect(lalr1Builder.Parser()).To(Equal(wantLALR1Parser))
		},
		Entry(
			"the unambiguous test grammar for Fig. 1",
			ielr1go.UnambiguousTestGrammarFig1,
			ielr1go.UnambiguousTestGrammarFig1LALR1Parser,
		),
		Entry(
			"the ambiguous test grammar for Fig. 2",
			ielr1go.AmbiguousTestGrammarFig2,
			ielr1go.AmbiguousTestGrammarFig2LALR1Parser,
		),
		Entry(
			"the goto follows test grammar for Fig. 5",
			ielr1go.GotoFollowsTestGrammarFig5,
			ielr1go.GotoFollowsTestGrammarFig5LALR1Parser,
		),
		Entry(
			"the goto follows caveats test grammar for Fig. 6",
			ielr1go.GotoFollowsCaveatsTestGrammarFig6,
			ielr1go.GotoFollowsCaveatsTestGrammarFig6LALRParser,
		),
		Entry(
			"the LR(1) but not LALR(1) grammar with a reduce/reduce conflict",
			ielr1go.ReduceReduceConflictTestGrammar,
			ielr1go.ReduceReduceConflictTestGrammarLALR1Parser,
		),
	)
})

func BenchmarkComputeLALR1ParserTables(b *testing.B) {
	benchmarks := []struct {
		description string
		grammar     frontend.Grammar
	}{
		{"Goto Follows Caveats Test Grammar Fig6", ielr1go.GotoFollowsCaveatsTestGrammarFig6},
		{"Unambiguous Test Grammar Fig1", ielr1go.UnambiguousTestGrammarFig1},
		{"Ambiguous Test Grammar Fig2", ielr1go.AmbiguousTestGrammarFig2},
		{"Goto Follows Test Grammar Fig5", ielr1go.GotoFollowsTestGrammarFig5},
		{"Reduce/Reduce Conflict Test Grammar", ielr1go.ReduceReduceConflictTestGrammar},
	}
	for _, benchmark := range benchmarks {
		b.Run(benchmark.description, func(b *testing.B) {
			augmentedGrammar := frontend.AugmentGrammar(benchmark.grammar)
			for range b.N {
				lalr1Builder := ielr1go.NewLALR1Builder(augmentedGrammar)
				lalr1Builder.Build()
			}
		})
	}
}
