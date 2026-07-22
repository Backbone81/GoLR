package conflict_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/conflict"
)

// Merge stability is definition 3.44 of IELR(1), and it is the one property phase 3 needs from a conflict resolution
// policy. Phase 3 merges two states when they have the same dominant contribution for an inadequacy. The merged state
// makes the contributions of both of them, so it makes the dominant contribution of neither of them if the dominant
// contribution of the union of their contributions is a different one. A policy for which that can happen would force
// phase 3 to compare the merged state instead of the two states it merges, which section 3.5.3 of the paper spells out
// in detail.
//
// Written as a property, a policy is merge stable when for every conflicted terminal and for every two sets of
// contributions with the same dominant contribution, the union of those two sets has that same dominant contribution
// again.
//
// The paper says verifying merge stability for the policy of GNU Bison was non-trivial. Composing the policy from
// several smaller policies makes it worse, because the composition can be changed. So we verify the property instead of
// arguing about it: the contributions of a conflict are few, so we can simply enumerate every subset of them, see
// allPossibleContributionSets. The two candidate states which phase 3 would merge are modelled by their contributions
// alone, because the contributions are all the dominant contribution function looks at.
var _ = Describe("Merge stability", func() {
	It("should hold for the default policy", func() {
		policy := conflict.DefaultPolicy(conflict.PrecedenceTestGrammar)
		contributionSets := allPossibleContributionSets()

		for terminalIdx := range conflict.PrecedenceTestGrammar.Terminals {
			for _, contributionsOfFirstState := range contributionSets {
				decisionOfFirstState := conflict.DominantContribution(policy, terminalIdx, contributionsOfFirstState)

				for _, contributionsOfSecondState := range contributionSets {
					decisionOfSecondState := conflict.DominantContribution(
						policy,
						terminalIdx,
						contributionsOfSecondState,
					)

					// Merge stability says nothing about two states which do not decide the conflicted terminal the same
					// way, because phase 3 does not merge those states in the first place. Only the pairs which survive
					// this filter are the ones the property is about.
					if !decisionOfFirstState.Equal(decisionOfSecondState) {
						continue
					}

					// Merging the two states gives a state which makes every contribution either of them makes, so its
					// contributions are the union of their contributions.
					contributionsOfMergedState := contributionsOfFirstState.Clone()
					contributionsOfMergedState.Merge(&contributionsOfSecondState)

					decisionOfMergedState := conflict.DominantContribution(
						policy,
						terminalIdx,
						contributionsOfMergedState,
					)

					// The property: the merged state has to decide the conflicted terminal the same way the two states
					// it was merged from decided it. Any other decision would make the merge change the behavior of the
					// parser, which is exactly what phase 3 relies on not happening.
					Expect(decisionOfMergedState).To(
						Equal(decisionOfFirstState),
						"the policy is not merge stable on terminal %s: the dominant contribution of %s is %s and the "+
							"dominant contribution of %s is %s, but the dominant contribution of their union %s is %s",
						conflict.PrecedenceTestGrammar.Terminals[terminalIdx].String(),
						contributionsOfFirstState.String(), decisionOfFirstState.String(),
						contributionsOfSecondState.String(), decisionOfSecondState.String(),
						contributionsOfMergedState.String(), decisionOfMergedState.String(),
					)
				}
			}
		}
	})
})
