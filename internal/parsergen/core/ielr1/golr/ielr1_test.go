package golr_test

import (
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	ielr1golrcore "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

var _ = Describe("IELR(1)", func() {
	// The random grammar corpus of the split states builder tests only observes the reduction lookahead sets of phase 4
	// through the presence of conflicts, which catches lookahead sets that are too large but not ones that are too
	// small: a lookahead set missing a terminal creates no conflict, it silently makes the parser reject valid input.
	// Pinning the exact parser table of a grammar which requires a split closes that gap for the sharpest hand-picked
	// case: the two isocores of the split "c" state must end up with exactly the mirrored one-terminal lookahead sets
	// their own predecessors generate.
	DescribeTable("should correctly compute the IELR(1) parser table",
		func(grammar frontend.Grammar, wantIELR1Parser backend.Parser) {
			augmentedGrammar := frontend.AugmentGrammar(grammar)
			ielr1Parser, err := ielr1golrcore.GrammarToParser(augmentedGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(ielr1Parser).To(Equal(wantIELR1Parser))
		},
		Entry(
			"the LR(1) but not LALR(1) grammar with a reduce/reduce conflict",
			ielr1golrcore.ReduceReduceConflictTestGrammar,
			ielr1golrcore.ReduceReduceConflictTestGrammarIELR1Parser,
		),
	)

	// Phase 5 of IELR(1) (section 3.7 of the paper) resolves the conflicts which splitting cannot remove, the genuine
	// ones canonical LR(1) has too. It runs at the GrammarToParser interface through conflict.Resolve, not inside the
	// builder, so the raw table BuildParser returns still carries the conflict while the table GrammarToParser returns is
	// free of it. The ambiguous grammar of figure 2 has such a genuine conflict and is the sharpest case for this split of
	// responsibilities.
	It("should resolve the genuine conflict of the ambiguous grammar only at the GrammarToParser interface", func() {
		augmentedGrammar := frontend.AugmentGrammar(ielr1golrcore.AmbiguousTestGrammarFig2)

		rawBuilder := ielr1golrcore.NewIELR1(augmentedGrammar, conflict.NewDefaultPolicy(augmentedGrammar))
		rawParser := rawBuilder.BuildParser()
		Expect(hasConflict(rawParser)).To(BeTrue(), "the raw split table is expected to keep the genuine conflict")

		resolvedParser, err := ielr1golrcore.GrammarToParser(augmentedGrammar)
		Expect(err).ToNot(HaveOccurred())
		Expect(hasConflict(resolvedParser)).To(BeFalse(), "phase 5 is expected to resolve the genuine conflict")
	})

	// The follow kernel items of definition 3.16 of IELR(1) are the kernel items whose lookahead sets a goto follow set
	// depends on. The definition asks for the reflexive transitive closure of the goto follows internal relation, so a
	// goto depends on the kernel items of its own state, and on the kernel items of every goto it reaches through the
	// internal relation. The grammars below pin down both halves of that closure, which are easy to lose when only the
	// edges of the internal relation are looked at.
	//
	// The gotos are keyed by the name of the nonterminal they happen on, which is unique per grammar here, and the
	// expected value is the list of kernel item indexes of the state the goto is coming from.
	DescribeTable("should compute the follow kernel items of definition 3.16",
		func(grammar frontend.Grammar, wantFollowKernelItems map[string][]int) {
			augmentedGrammar := frontend.AugmentGrammar(grammar)
			ielr1 := ielr1golrcore.NewIELR1(augmentedGrammar, conflict.NewDefaultPolicy(augmentedGrammar))
			ielr1.BuildParser()

			gotFollowKernelItems := make(map[string][]int)
			for gotoIdx, gotoRecord := range ielr1.GotoRecords() {
				nonterminalName := augmentedGrammar.Nonterminals[gotoRecord.NonterminalIdx].Name
				Expect(gotFollowKernelItems).ToNot(
					HaveKey(nonterminalName),
					"the test grammar is expected to have a single goto per nonterminal, so that the goto is fully "+
						"identified by the nonterminal it happens on",
				)
				followKernelItems := ielr1.FollowKernelItems()[gotoIdx]
				gotFollowKernelItems[nonterminalName] = slices.Collect(followKernelItems.All())
			}
			Expect(gotFollowKernelItems).To(Equal(wantFollowKernelItems))
		},
		Entry(
			"the reflexive dependency on a kernel item of the goto's own state",
			ielr1golrcore.FollowKernelItemsReflexiveTestGrammar,
			map[string][]int{
				// The kernel item "$accept -> .S $eof" has the end of input marker after S, so nothing which follows S
				// can come from the lookahead set of that kernel item.
				"S": nil,

				// The goto on C comes from the state with the single kernel item "S -> a.C", whose lookahead set follows
				// C. No goto follows internal relation is involved, the goto depends on the kernel item directly.
				"C": {0},
			},
		),
		Entry(
			"the transitive dependency through a chain of goto follows internal relations",
			ielr1golrcore.FollowKernelItemsTransitiveTestGrammar,
			map[string][]int{
				"S": nil,

				// All three gotos come from the state with the single kernel item "S -> a.A". The goto on A depends on
				// that kernel item directly, the goto on B over one internal relation, and the goto on C over two.
				"A": {0},
				"B": {0},
				"C": {0},
			},
		),
	)
})
