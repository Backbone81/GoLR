package conflict_test

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/conflict"
)

var _ = Describe("Policies", func() {
	// The mapping of the two switches to the four compositions only exists in SelectPolicy, so it is pinned down here.
	// The factories are functions, which Go cannot compare, so they are compared by the code they point to.
	DescribeTable("should map the switches of SelectPolicy to the matching policy",
		func(failOnShiftReduceConflicts bool, failOnReduceReduceConflicts bool, want conflict.PolicyFactory) {
			got := conflict.SelectPolicy(failOnShiftReduceConflicts, failOnReduceReduceConflicts)
			Expect(reflect.ValueOf(got).Pointer()).To(Equal(reflect.ValueOf(want).Pointer()))
		},
		Entry("neither fails", false, false, conflict.PolicyFactory(conflict.DefaultPolicy)),
		Entry("shift/reduce conflicts fail", true, false,
			conflict.PolicyFactory(conflict.PrecedenceAndEarliestProductionPolicy)),
		Entry("reduce/reduce conflicts fail", false, true,
			conflict.PolicyFactory(conflict.PrecedenceAndShiftOverReducePolicy)),
		Entry("both fail", true, true, conflict.PolicyFactory(conflict.PrecedencePolicy)),
	)

	// The dominant contribution function of definition 2.19 of IELR(1) decides which action wins a conflict. The
	// grammar declares one operator per case a policy has to tell apart, see PrecedenceTestGrammar.
	DescribeTable("should compute the dominant contribution with the default policy",
		func(terminalIdx int, contributions conflict.ContributionSet, wantDecision conflict.Decision) {
			policy := conflict.DefaultPolicy(conflict.PrecedenceTestGrammar)

			decision, _ := conflict.DominantContribution(policy, terminalIdx, contributions)

			Expect(decision).To(
				Equal(wantDecision),
				"expected the dominant contribution to be %s, but it was %s",
				wantDecision,
				decision,
			)
		},
		Entry(
			"no contribution at all leaves the dominant contribution undefined",
			conflict.PrecedenceTestGrammarTerminalIdxTimes,
			conflict.NewContributionSet(),
			conflict.NewUndefinedDecision(),
		),
		Entry(
			"a single contribution is the dominant contribution, as there is nothing it competes with",
			conflict.PrecedenceTestGrammarTerminalIdxTimes,
			conflict.NewContributionSet(conflict.NewShiftContribution()),
			conflict.NewDominantDecision(conflict.NewShiftContribution()),
		),

		// The precedence declarations decide a shift/reduce conflict. The conflicted terminal is the one which would be
		// shifted, and the production of the reduction inherits the precedence of its operator.
		Entry(
			"the shift wins when the conflicted terminal binds tighter than the production",
			conflict.PrecedenceTestGrammarTerminalIdxTimes,
			conflict.NewContributionSet(
				conflict.NewShiftContribution(),
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxPlus),
			),
			conflict.NewDominantDecision(conflict.NewShiftContribution()),
		),
		Entry(
			"the reduction wins when the production binds tighter than the conflicted terminal",
			conflict.PrecedenceTestGrammarTerminalIdxPlus,
			conflict.NewContributionSet(
				conflict.NewShiftContribution(),
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxTimes),
			),
			conflict.NewDominantDecision(
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxTimes),
			),
		),
		Entry(
			"the reduction wins on equal precedence when the conflicted terminal is left associative",
			conflict.PrecedenceTestGrammarTerminalIdxPlus,
			conflict.NewContributionSet(
				conflict.NewShiftContribution(),
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxPlus),
			),
			conflict.NewDominantDecision(
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxPlus),
			),
		),
		Entry(
			"the shift wins on equal precedence when the conflicted terminal is right associative",
			conflict.PrecedenceTestGrammarTerminalIdxPower,
			conflict.NewContributionSet(
				conflict.NewShiftContribution(),
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxPower),
			),
			conflict.NewDominantDecision(conflict.NewShiftContribution()),
		),
		Entry(
			"the conflicted terminal is rejected on equal precedence when it does not associate at all",
			conflict.PrecedenceTestGrammarTerminalIdxCompare,
			conflict.NewContributionSet(
				conflict.NewShiftContribution(),
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxCompare),
			),
			conflict.NewErrorDecision(),
		),

		// Without precedence declarations to go by, the shift beats the reduction and the production which was declared
		// first beats the ones declared after it.
		Entry(
			"the shift wins when the production has no precedence declared",
			conflict.PrecedenceTestGrammarTerminalIdxTimes,
			conflict.NewContributionSet(
				conflict.NewShiftContribution(),
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxIdentity),
			),
			conflict.NewDominantDecision(conflict.NewShiftContribution()),
		),
		Entry(
			"the shift wins when the conflicted terminal has no precedence declared",
			conflict.PrecedenceTestGrammarTerminalIdxIdentity,
			conflict.NewContributionSet(
				conflict.NewShiftContribution(),
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxTimes),
			),
			conflict.NewDominantDecision(conflict.NewShiftContribution()),
		),
		Entry(
			"the production which was declared first wins a conflict between two reductions",
			conflict.PrecedenceTestGrammarTerminalIdxTimes,
			conflict.NewContributionSet(
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxPower),
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxPlus),
			),
			conflict.NewDominantDecision(
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxPlus),
			),
		),

		// A conflict with more than one reduction is decided reduction by reduction against the shift. The reductions
		// which survive that are left to the policies behind the precedence policy.
		Entry(
			"a reduction which loses against the shift is gone before the reductions are compared with each other",
			conflict.PrecedenceTestGrammarTerminalIdxTimes,
			conflict.NewContributionSet(
				conflict.NewShiftContribution(),
				// This reduction binds looser than "*", so it loses against the shift, even though it is the production
				// which was declared first.
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxPlus),
				// This reduction binds tighter than "*", so it beats the shift and stays.
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxPower),
			),
			conflict.NewDominantDecision(
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxPower),
			),
		),
		Entry(
			"the conflicted terminal is rejected as soon as a single reduction asks for it",
			conflict.PrecedenceTestGrammarTerminalIdxCompare,
			conflict.NewContributionSet(
				conflict.NewShiftContribution(),
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxCompare),
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxPower),
			),
			conflict.NewErrorDecision(),
		),
	)

	// A policy which is not part of the compound policy is not applied, which is the whole point of composing the
	// conflict resolution instead of hard coding it.
	Describe("the compound policy", func() {
		// A policy is free to leave a conflict unresolved, but the default policy never does: the shift beats a
		// reduction and the earliest production beats the ones declared after it whenever precedence has nothing to
		// say.
		It("should be total for the default policy, so that it decides every conflict", func() {
			policy := conflict.DefaultPolicy(conflict.PrecedenceTestGrammar)

			for terminalIdx := range conflict.PrecedenceTestGrammar.Terminals {
				for _, contributions := range allPossibleContributionSets() {
					remaining, _ := policy.Resolve(terminalIdx, contributions)

					Expect(remaining.Length()).To(
						BeNumerically("<=", 1),
						"the policy left %s to decide between on terminal %s, so it is not total",
						remaining.String(),
						conflict.PrecedenceTestGrammar.Terminals[terminalIdx].String(),
					)
				}
			}
		})

		It("should resolve a shift/reduce conflict in favor of the reduction without the precedence policy", func() {
			// The precedence declarations of the grammar say that the reduction of "E -> E * E" wins over the shift of
			// "+". Without the precedence policy, the shift over reduce policy decides instead, and the shift wins.
			policy := conflict.CompoundPolicy(
				conflict.ShiftOverReducePolicy,
				conflict.EarliestProductionPolicy,
			)(conflict.PrecedenceTestGrammar)
			contributions := conflict.NewContributionSet(
				conflict.NewShiftContribution(),
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxTimes),
			)

			decision, _ := conflict.DominantContribution(
				policy,
				conflict.PrecedenceTestGrammarTerminalIdxPlus,
				contributions,
			)

			Expect(decision).To(Equal(conflict.NewDominantDecision(conflict.NewShiftContribution())))
		})

		It("should leave a shift/reduce conflict unresolved without the shift over reduce policy", func() {
			// The identity production has no precedence to compare the conflicted terminal against, and no policy
			// behind the precedence policy decides between a shift and a reduction, so the conflict stands. This is
			// how a grammar author insists on deciding every shift/reduce conflict with explicit precedence
			// declarations.
			policy := conflict.CompoundPolicy(
				conflict.PrecedencePolicy,
				conflict.EarliestProductionPolicy,
			)(conflict.PrecedenceTestGrammar)
			contributions := conflict.NewContributionSet(
				conflict.NewShiftContribution(),
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxIdentity),
			)

			decision, _ := conflict.DominantContribution(
				policy,
				conflict.PrecedenceTestGrammarTerminalIdxTimes,
				contributions,
			)

			Expect(decision).To(Equal(conflict.NewUnresolvedDecision(contributions)))
		})

		It("should rule out the reductions which lost even when the conflict stays unresolved", func() {
			// The reduction of "E -> E + E" binds looser than "*", so it loses against the shift and is ruled out. The
			// identity production has no precedence, so the shift and its reduction stay undecided.
			policy := conflict.CompoundPolicy(
				conflict.PrecedencePolicy,
			)(conflict.PrecedenceTestGrammar)
			contributions := conflict.NewContributionSet(
				conflict.NewShiftContribution(),
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxPlus),
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxIdentity),
			)

			decision, _ := conflict.DominantContribution(
				policy,
				conflict.PrecedenceTestGrammarTerminalIdxTimes,
				contributions,
			)

			Expect(decision).To(Equal(conflict.NewUnresolvedDecision(conflict.NewContributionSet(
				conflict.NewShiftContribution(),
				conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxIdentity),
			))))
		})
	})

	// The undeclared contributions are what the grammar author is told about, so they have to be exactly what the
	// declarations left competing, and nothing when the declarations decided the conflict.
	Describe("the undeclared contributions", func() {
		DescribeTable("should hold what precedence left for a rule of last resort",
			func(terminalIdx int, contributions conflict.ContributionSet, wantUndeclared conflict.ContributionSet) {
				policy := conflict.DefaultPolicy(conflict.PrecedenceTestGrammar)

				_, undeclared := conflict.DominantContribution(policy, terminalIdx, contributions)

				Expect(undeclared).To(Equal(wantUndeclared))
			},
			Entry("nothing, when precedence decides the conflict",
				conflict.PrecedenceTestGrammarTerminalIdxPlus,
				conflict.NewContributionSet(
					conflict.NewShiftContribution(),
					conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxTimes),
				),
				conflict.ContributionSet{},
			),
			// "E -> E * E" beats the shift of "+", which leaves a reduce/reduce conflict with the identity production.
			Entry("the reductions, when precedence removes the shift",
				conflict.PrecedenceTestGrammarTerminalIdxPlus,
				conflict.NewContributionSet(
					conflict.NewShiftContribution(),
					conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxTimes),
					conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxIdentity),
				),
				conflict.NewContributionSet(
					conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxTimes),
					conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxIdentity),
				),
			),
			// "E -> E + E" loses against the shift of "*", which leaves a shift/reduce conflict with the identity
			// production.
			Entry("the shift and a reduction, when precedence removes the other reduction",
				conflict.PrecedenceTestGrammarTerminalIdxTimes,
				conflict.NewContributionSet(
					conflict.NewShiftContribution(),
					conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxPlus),
					conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxIdentity),
				),
				conflict.NewContributionSet(
					conflict.NewShiftContribution(),
					conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxIdentity),
				),
			),
		)
	})
})
