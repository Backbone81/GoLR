package golr_test

import (
	"fmt"
	"math/rand"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
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
			lalr1Parser, _, err := lalr1golrcore.GrammarToUnresolvedParser(grammar, conflict.DefaultPolicy)
			Expect(err).ToNot(HaveOccurred())
			ielr1Parser, _, err := ielr1golrcore.GrammarToUnresolvedParser(grammar, conflict.DefaultPolicy)
			Expect(err).ToNot(HaveOccurred())

			lr1Parser, _, err := lr1golrcore.GrammarToUnresolvedParser(grammar, conflict.DefaultPolicy)
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

		lalr1Parser, _, err := lalr1golrcore.GrammarToUnresolvedParser(grammar, conflict.DefaultPolicy)
		Expect(err).ToNot(HaveOccurred())
		ielr1Parser, _, err := ielr1golrcore.GrammarToUnresolvedParser(grammar, conflict.DefaultPolicy)
		Expect(err).ToNot(HaveOccurred())

		Expect(conflict.HasConflict(lalr1Parser)).To(
			BeTrue(),
			"the LALR(1) parser tables are expected to have the mysterious conflict",
		)
		Expect(conflict.HasConflict(ielr1Parser)).To(
			BeFalse(),
			"phase 3 is expected to remove the mysterious conflict",
		)
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

			lr1Parser, _, err := lr1golrcore.GrammarToUnresolvedParser(grammar, conflict.DefaultPolicy)
			if err != nil {
				// A grammar whose canonical LR(1) automaton exceeds the addressable state limit cannot be the oracle. It
				// is skipped, not a failure of the builder under test.
				Expect(err).To(MatchError(backend.ErrStateLimitExceeded), "grammar seed %d:\n%s", grammarSeed, grammar.String())
				continue
			}

			// Both are bounded above by the canonical LR(1) automaton which just fit, so neither can reach the state
			// limit here.
			lalr1Parser, _, err := lalr1golrcore.GrammarToUnresolvedParser(grammar, conflict.DefaultPolicy)
			Expect(err).ToNot(HaveOccurred(), "grammar seed %d:\n%s", grammarSeed, grammar.String())
			ielr1Parser, _, err := ielr1golrcore.GrammarToUnresolvedParser(grammar, conflict.DefaultPolicy)
			Expect(err).ToNot(HaveOccurred(), "grammar seed %d:\n%s", grammarSeed, grammar.String())

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

	// A policy without a rule of last resort leaves some conflicts unresolved, and phase 3 has to keep two isocores apart
	// when their conflicts are left unresolved between different contributions. The compatibility test of definition
	// 3.43 compares the decisions, and two unresolved decisions are only equal when they were left with the same
	// contributions. If phase 3 merged such isocores anyway, the merged state would carry the contributions of both,
	// which is a conflict canonical LR(1) does not have, and the conflict report would describe a state the grammar
	// author cannot find in the canonical LR(1) automaton.
	//
	// The grammar has two isocores which canonical LR(1) leaves with the shift of "c" against A -> e and against B -> e.
	// Under a policy which decides every shift/reduce conflict in favor of the shift, both isocores decide the same, so
	// phase 3 does not split them. Under a policy which leaves the shift/reduce conflicts unresolved, it has to.
	DescribeTable("should keep isocores apart whose conflicts are left unresolved between different contributions",
		func(policyFactory conflict.PolicyFactory, expectSplit bool) {
			grammar := ielr1golrcore.UnresolvedShiftReduceIsocoresTestGrammar

			lalr1Parser, _, err := lalr1golrcore.GrammarToUnresolvedParser(grammar, policyFactory)
			Expect(err).ToNot(HaveOccurred())
			ielr1Parser, _, err := ielr1golrcore.GrammarToUnresolvedParser(grammar, policyFactory)
			Expect(err).ToNot(HaveOccurred())
			if expectSplit {
				Expect(len(ielr1Parser.States)).To(
					BeNumerically(">", len(lalr1Parser.States)),
					"phase 3 is expected to split the isocores apart",
				)
			} else {
				Expect(ielr1Parser.States).To(
					HaveLen(len(lalr1Parser.States)),
					"phase 3 is not expected to split isocores which decide the same",
				)
			}

			// The state indexes differ between the two automatons, so the unresolved conflicts are compared by the
			// conflicted terminal and the contributions they were left with.
			_, ielr1Conflicts, _, _ := ielr1golrcore.GrammarToParser(grammar, policyFactory)
			_, lr1Conflicts, _, _ := lr1golrcore.GrammarToParser(grammar, policyFactory)
			if expectSplit {
				// One unresolved shift/reduce conflict in each of the two isocores, so that the comparison below is not
				// vacuous.
				Expect(unresolvedConflictDescriptions(lr1Conflicts)).To(HaveLen(2))
			} else {
				Expect(unresolvedConflictDescriptions(lr1Conflicts)).To(BeEmpty())
			}
			Expect(unresolvedConflictDescriptions(ielr1Conflicts)).To(
				ConsistOf(unresolvedConflictDescriptions(lr1Conflicts)),
			)
		},
		Entry("the default policy",
			conflict.DefaultPolicy,
			false,
		),
		Entry("precedence and shift over reduce",
			conflict.PrecedenceAndShiftOverReducePolicy,
			false,
		),
		Entry("precedence and earliest production",
			conflict.PrecedenceAndEarliestProductionPolicy,
			true,
		),
		Entry("precedence alone",
			conflict.PrecedencePolicy,
			true,
		),
	)
})

// unresolvedConflictDescriptions describes each unresolved conflict by its conflicted terminal and the contributions it
// was left with, which is how two automatons with different state indexes can be compared on their unresolved
// conflicts.
func unresolvedConflictDescriptions(conflicts []conflict.Conflict) []string {
	var result []string
	for _, c := range conflicts {
		if c.Decision.Kind != conflict.DecisionUnresolved {
			continue
		}
		result = append(result, fmt.Sprintf("terminal %d: %s", c.TerminalIdx, c.Decision.String()))
	}
	return result
}
