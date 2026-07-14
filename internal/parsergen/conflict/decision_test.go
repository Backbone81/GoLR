package conflict_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/conflict"
)

var _ = Describe("Decision", func() {
	// Phase 3 of IELR(1) merges two states only when they decide a conflict the same way, see definition 3.43. An
	// unresolved decision takes part in that comparison like any other, so two unresolved decisions may only be equal
	// when their conflicts were left with the same contributions.
	Describe("Equal", func() {
		It("should compare unresolved decisions by the contributions they were left with", func() {
			first := conflict.NewUnresolvedDecision(conflict.NewContributionSet(
				conflict.NewShiftContribution(),
				conflict.NewReduceContribution(0),
			))
			same := conflict.NewUnresolvedDecision(conflict.NewContributionSet(
				conflict.NewReduceContribution(0),
				conflict.NewShiftContribution(),
			))
			other := conflict.NewUnresolvedDecision(conflict.NewContributionSet(
				conflict.NewShiftContribution(),
				conflict.NewReduceContribution(1),
			))

			Expect(first.Equal(same)).To(
				BeTrue(),
				"both conflicts were left with the same contributions, so the decisions are the same",
			)
			Expect(first.Equal(other)).To(
				BeFalse(),
				"the conflicts are undecided between different contributions, so the decisions differ",
			)
		})

		It("should tell the decision kinds apart", func() {
			decisions := []conflict.Decision{
				conflict.NewUndefinedDecision(),
				conflict.NewDominantDecision(conflict.NewShiftContribution()),
				conflict.NewErrorDecision(),
				conflict.NewUnresolvedDecision(conflict.NewContributionSet(
					conflict.NewShiftContribution(),
					conflict.NewReduceContribution(0),
				)),
			}

			for i, first := range decisions {
				for j, second := range decisions {
					Expect(first.Equal(second)).To(
						Equal(i == j),
						"the decision %s is expected to be equal to %s exactly when it is the same decision",
						first.String(),
						second.String(),
					)
				}
			}
		})
	})

	Describe("Survivors", func() {
		It("should keep only the contribution which won", func() {
			dominant := conflict.NewReduceContribution(1)
			decision := conflict.NewDominantDecision(dominant)

			survivors := decision.Survivors()
			Expect(survivors.Length()).To(Equal(1))
			Expect(survivors.Contains(dominant)).To(BeTrue())
		})

		It("should keep every contribution an unresolved conflict was left with", func() {
			remaining := conflict.NewContributionSet(
				conflict.NewShiftContribution(),
				conflict.NewReduceContribution(0),
			)
			decision := conflict.NewUnresolvedDecision(remaining)

			survivors := decision.Survivors()
			Expect(survivors.Equal(&remaining)).To(BeTrue())
		})

		It("should keep nothing when the terminal is rejected", func() {
			decision := conflict.NewErrorDecision()

			survivors := decision.Survivors()
			Expect(survivors.IsEmpty()).To(BeTrue())
		})
	})
})
