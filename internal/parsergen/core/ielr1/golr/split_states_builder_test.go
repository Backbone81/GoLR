package golr_test

import (
	"math/rand"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/conflict"
	ielr1golrcore "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	"github.com/backbone81/golr/internal/parsergen/core/ielr1/golr/oracle"
	lalr1golrcore "github.com/backbone81/golr/internal/parsergen/core/lalr1/golr"
	lr1golrcore "github.com/backbone81/golr/internal/parsergen/core/lr1/golr"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

var _ = Describe("Split States Builder", func() {
	// Phase 3 splits the LALR(1) states into the isocores of the minimal LR(1) parser tables. We verify it behaviorally
	// against canonical LR(1), which is the defining property of IELR(1): under a conflict-preserving policy it removes
	// exactly the conflicts of the LALR(1) parser tables which canonical LR(1) does not have, and nothing else. Two
	// invariants capture that without comparing the tables structurally, which two correct generators are free to differ
	// on:
	//
	//  1. The state count is bounded by the two extremes, |LALR(1)| <= |IELR(1)| <= |canonical LR(1)|. IELR(1) only ever
	//     splits states, so it never drops below LALR(1), and it never splits further than canonical LR(1).
	//  2. IELR(1) has a conflict exactly when canonical LR(1) has one. A conflict of the LALR(1) parser tables which
	//     canonical LR(1) does not have is a mysterious conflict which phase 3 removes by splitting; a conflict canonical
	//     LR(1) has too is genuine and survives.
	//
	// The conflict invariant is about the raw automaton, before phase 5 resolves anything, so we compare the tables
	// GrammarToUnresolvedParser returns, not the conflict-free ones GrammarToParser produces. Resolving the conflicts
	// with the default policy would leave conflict.Detect with nothing to report and defeat the comparison.
	DescribeTable("should agree with canonical LR(1) on the state count bounds and the conflicts",
		func(grammar frontend.Grammar) {
			lalr1Parser := lalr1golrcore.GrammarToUnresolvedParser(grammar, conflict.DefaultPolicy)
			ielr1Parser := ielr1golrcore.GrammarToUnresolvedParser(grammar, conflict.DefaultPolicy)

			lr1Parser, err := lr1golrcore.GrammarToUnresolvedParser(grammar, conflict.DefaultPolicy)
			Expect(err).ToNot(HaveOccurred())

			Expect(len(ielr1Parser.States)).To(BeNumerically(">=", len(lalr1Parser.States)))
			Expect(len(ielr1Parser.States)).To(BeNumerically("<=", len(lr1Parser.States)))

			// The two automatons have different states, so they can differ in how many conflicts they report; what has to
			// agree is whether they are left with a conflict at all.
			Expect(conflict.HasConflict(ielr1Parser)).To(Equal(conflict.HasConflict(lr1Parser)))
		},
		Entry("the unambiguous test grammar for Fig. 1", ielr1golrcore.UnambiguousTestGrammarFig1),
		Entry("the ambiguous test grammar for Fig. 2", ielr1golrcore.AmbiguousTestGrammarFig2),
		Entry("the goto follows test grammar for Fig. 5", ielr1golrcore.GotoFollowsTestGrammarFig5),
		Entry("the goto follows caveats test grammar for Fig. 6", ielr1golrcore.GotoFollowsCaveatsTestGrammarFig6),
		Entry("the LR(1) but not LALR(1) grammar", ielr1golrcore.ReduceReduceConflictTestGrammar),
	)

	// The reduce/reduce grammar is LR(1) but not LALR(1): its LALR(1) parser tables have a reduce/reduce conflict which
	// canonical LR(1) does not have. It is the sharpest hand-picked case for phase 3, because getting the conflict to
	// disappear requires actually splitting a state. A phase 3 which never split would leave the conflict in place and
	// silently degrade IELR(1) into LALR(1), which the state count check pins down alongside the conflict check.
	It("should split a state to remove the mysterious conflict of the reduce/reduce grammar", func() {
		grammar := ielr1golrcore.ReduceReduceConflictTestGrammar

		lalr1Parser := lalr1golrcore.GrammarToUnresolvedParser(grammar, conflict.DefaultPolicy)
		ielr1Parser := ielr1golrcore.GrammarToUnresolvedParser(grammar, conflict.DefaultPolicy)

		Expect(conflict.HasConflict(lalr1Parser)).To(BeTrue(), "the LALR(1) parser tables are expected to have the mysterious conflict")
		Expect(conflict.HasConflict(ielr1Parser)).To(BeFalse(), "phase 3 is expected to remove the mysterious conflict")
		Expect(len(ielr1Parser.States)).To(
			BeNumerically(">", len(lalr1Parser.States)),
			"phase 3 is expected to have split at least one state",
		)
	})

	// We feed a large corpus of random grammars through the LALR(1) builder, the IELR(1) builder and canonical LR(1) and
	// assert the two invariants on every one. The hand-picked grammars above pin down specific figures, the random corpus
	// is there to surprise us with grammar shapes we did not think to write down. Each run draws a fresh corpus from the
	// Ginkgo random seed, so a rare edge case surfaces over repeated runs; a failing run replays with `ginkgo --seed=...`
	// and the reported per-grammar seed reconstructs the single failing grammar on its own.
	It("should agree with canonical LR(1) on a corpus of random grammars", func() {
		// grammarCount trades test time for how much of the grammar space is explored. The corpus builds three automatons
		// per grammar under -race, so keep it to a size which still finishes in a few seconds; bump it when hunting a bug.
		const grammarCount = 1000

		var compared, mysteriousConflictRemoved int

		masterRng := rand.New(rand.NewSource(GinkgoRandomSeed()))
		for range grammarCount {
			grammarSeed := masterRng.Int63()
			grammar := oracle.DefaultGrammarGenerator(rand.New(rand.NewSource(grammarSeed))).Generate()

			lr1Parser, err := lr1golrcore.GrammarToUnresolvedParser(grammar, conflict.DefaultPolicy)
			if err != nil {
				// A grammar whose canonical LR(1) automaton exceeds the addressable state limit cannot be the oracle. It
				// is skipped, not a failure of the builder under test.
				Expect(err).To(MatchError(lr1golrcore.ErrStateLimitExceeded), "grammar seed %d:\n%s", grammarSeed, grammar.String())
				continue
			}

			lalr1Parser := lalr1golrcore.GrammarToUnresolvedParser(grammar, conflict.DefaultPolicy)
			ielr1Parser := ielr1golrcore.GrammarToUnresolvedParser(grammar, conflict.DefaultPolicy)

			Expect(len(ielr1Parser.States)).To(
				BeNumerically(">=", len(lalr1Parser.States)),
				"IELR(1) dropped below the LALR(1) state count, grammar seed %d:\n%s", grammarSeed, grammar.String(),
			)
			Expect(len(ielr1Parser.States)).To(
				BeNumerically("<=", len(lr1Parser.States)),
				"IELR(1) split further than canonical LR(1), grammar seed %d:\n%s", grammarSeed, grammar.String(),
			)
			Expect(conflict.HasConflict(ielr1Parser)).To(
				Equal(conflict.HasConflict(lr1Parser)),
				"IELR(1) and canonical LR(1) disagree on the conflicts, grammar seed %d:\n%s", grammarSeed, grammar.String(),
			)

			compared++
			if conflict.HasConflict(lalr1Parser) && !conflict.HasConflict(lr1Parser) {
				// The LALR(1) parser tables have a mysterious conflict which canonical LR(1) does not have, and IELR(1)
				// removed it, as the conflict check above just confirmed. These are the grammars which actually exercise
				// the state splitting, so we track how many the corpus reaches to guard its discriminating power.
				mysteriousConflictRemoved++
			}
		}

		GinkgoWriter.Printf(
			"random grammar corpus: %d compared, %d with a mysterious conflict IELR(1) removed\n",
			compared, mysteriousConflictRemoved,
		)

		// Guard the discriminating power of the corpus: passing thousands of grammars which never trigger a split proves
		// little, so fail if the corpus stops reaching the grammars phase 3 exists for. A healthy corpus removes a
		// mysterious conflict on the order of two dozen grammars per thousand, so a comfortable margin below that still
		// catches a generator which degraded into trivial grammars, where the count would collapse towards zero.
		Expect(compared).To(BeNumerically(">", grammarCount/2))
		Expect(mysteriousConflictRemoved).To(BeNumerically(">", 10))
	})
})
