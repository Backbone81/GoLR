package conflict_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/conflict"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

var _ = Describe("SplitStability", func() {
	// The policies compute split stability analytically in a single pass, which must agree with the brute force
	// definition of definition 3.35: the dominant contribution is split-stable exactly when it stays the same for every
	// subset of the potential contributions. The enumeration below is that brute force definition, so it is the ground
	// truth to check the analytic bookkeeping against.
	//
	// Every way the contributions of a conflict can be split into always, potential, and never contributions is checked,
	// on every terminal of the entry's grammar, so that the precedence classes the grammar declares are all exercised.
	// The policies are checked one by one and composed, because split stability of a compound policy is what the shared
	// bookkeeping is really about. The grammar-reading policies run on both test grammars, because the two grammars
	// declare disjoint precedence cases: PrecedenceTestGrammar covers the associativities and the relative binding
	// strengths, and MultiRejecterTestGrammar covers several rejecting reductions on the same terminal and a precedence
	// without an associativity.
	DescribeTable("should compute the same split stability as the brute force enumeration of definition 3.35",
		func(grammar frontend.Grammar, policy conflict.Policy) {
			for terminalIdx := range grammar.Terminals {
				for _, partition := range allAlwaysPotentialPartitions(allContributionsOfGrammar(grammar)) {
					if partition.potentials.Contains(conflict.NewShiftContribution()) {
						// A shift is an always contribution in every conflict: point 1 of definition 3.30 of IELR(1) makes
						// its contribution matrix row undefined, which is an always contribution by point 2(a) of
						// definition 3.28, because splitting a state keeps its transitions so every isocore makes the shift.
						// A potential shift is therefore not a valid input, and ContributeSplitStability relies on that, see
						// the precedence policy, so it is not tested with one.
						continue
					}
					splitStability := conflict.NewSplitStability(partition.always, partition.potentials)
					policy.ContributeSplitStability(terminalIdx, &splitStability)
					analytic := splitStability.IsSplitStable()

					enumerated := isSplitStableByEnumeration(
						policy,
						terminalIdx,
						partition.always,
						partition.potentials,
					)

					Expect(analytic).To(
						Equal(enumerated),
						"the analytic split stability disagreed with the enumeration on terminal %s for always "+
							"contributions %s and potential contributions %s",
						grammar.Terminals[terminalIdx].String(),
						partition.always.String(),
						partition.potentials.String(),
					)
				}
			}
		},
		Entry("the default policy of GNU Bison",
			conflict.PrecedenceTestGrammar, conflict.DefaultPolicy(conflict.PrecedenceTestGrammar)),
		Entry("the shift over reduce policy alone",
			conflict.PrecedenceTestGrammar, conflict.ShiftOverReducePolicy(conflict.PrecedenceTestGrammar)),
		Entry("the earliest production policy alone",
			conflict.PrecedenceTestGrammar, conflict.EarliestProductionPolicy(conflict.PrecedenceTestGrammar)),
		Entry("the precedence policy alone",
			conflict.PrecedenceTestGrammar, conflict.PrecedencePolicy(conflict.PrecedenceTestGrammar)),
		Entry("the conflict-preserving empty compound policy",
			conflict.PrecedenceTestGrammar, conflict.CompoundPolicy()(conflict.PrecedenceTestGrammar)),
		Entry("the conflict-preserving null policy",
			conflict.PrecedenceTestGrammar, conflict.NullPolicy(conflict.PrecedenceTestGrammar)),
		Entry("the default policy of GNU Bison on the multi rejecter grammar",
			conflict.MultiRejecterTestGrammar, conflict.DefaultPolicy(conflict.MultiRejecterTestGrammar)),
		Entry("the precedence policy alone on the multi rejecter grammar",
			conflict.MultiRejecterTestGrammar, conflict.PrecedencePolicy(conflict.MultiRejecterTestGrammar)),
	)

	// The null policy resolves nothing, so the general case of definition 3.35 collapses to observation 3.33: the
	// dominant contribution is split-stable exactly when there is no potential contribution, because then every isocore
	// makes the same contributions. This is the behavior phase 2 relies on for its conflict-preserving tests, so it is
	// pinned down here.
	It("should reduce to observation 3.33 for the null policy", func() {
		shift := conflict.NewShiftContribution()
		reduce := conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxPlus)

		onlyAlways := conflict.NewSplitStability(conflict.NewContributionSet(shift, reduce), conflict.ContributionSet{})
		conflict.NullPolicy(conflict.PrecedenceTestGrammar).
			ContributeSplitStability(conflict.PrecedenceTestGrammarTerminalIdxTimes, &onlyAlways)
		Expect(onlyAlways.IsSplitStable()).To(BeTrue())

		withPotential := conflict.NewSplitStability(
			conflict.NewContributionSet(shift),
			conflict.NewContributionSet(reduce),
		)
		conflict.NullPolicy(conflict.PrecedenceTestGrammar).
			ContributeSplitStability(conflict.PrecedenceTestGrammarTerminalIdxTimes, &withPotential)
		Expect(withPotential.IsSplitStable()).To(BeFalse())
	})

	// This is the paper's own example on page 24: shift over reduce makes a shift/reduce conflict split-stable no matter
	// which reductions are potential contributions, because the shift always dominates. So phase 2 can discard the
	// annotation of such a conflict, which is the whole payoff of the general case over observation 3.33.
	It("should find a shift/reduce conflict split-stable under shift over reduce", func() {
		splitStability := conflict.NewSplitStability(
			conflict.NewContributionSet(conflict.NewShiftContribution()),
			conflict.NewContributionSet(conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxIdentity)),
		)

		// The identity production has no precedence, so precedence leaves the conflict to shift over reduce, which lets
		// the shift win regardless of whether the reduction is made.
		conflict.DefaultPolicy(conflict.PrecedenceTestGrammar).
			ContributeSplitStability(conflict.PrecedenceTestGrammarTerminalIdxIdentity, &splitStability)

		Expect(splitStability.IsSplitStable()).To(BeTrue())
	})

	// A conflict holds several rejecting reductions when several productions carry the precedence of the same
	// nonassociative terminal. A single rejecter already turns the conflict into an error action, so the error holds in
	// every isocore as soon as one rejecter is an always contribution, no matter which of the other rejecters are
	// potential - and it holds in no isocore reliably when every rejecter is potential, because the isocores which make
	// none of them resolve the conflict some other way.
	It("should decide split stability from the always rejecter among several rejecting reductions", func() {
		lessReduce := conflict.NewReduceContribution(conflict.MultiRejecterTestGrammarProductionIdxLess)
		lessEqualReduce := conflict.NewReduceContribution(conflict.MultiRejecterTestGrammarProductionIdxLessEqual)
		policy := conflict.DefaultPolicy(conflict.MultiRejecterTestGrammar)

		// Both reductions reject the nonassociative "<", and the one on "E -> E < E" is an always contribution, so
		// every isocore makes it and errors out.
		withAlwaysRejecter := conflict.NewSplitStability(
			conflict.NewContributionSet(conflict.NewShiftContribution(), lessReduce),
			conflict.NewContributionSet(lessEqualReduce),
		)
		policy.ContributeSplitStability(conflict.MultiRejecterTestGrammarTerminalIdxLess, &withAlwaysRejecter)
		Expect(withAlwaysRejecter.IsSplitStable()).To(BeTrue())

		// With both rejecters potential, an isocore which makes neither shifts instead of erroring out.
		onlyPotentialRejecters := conflict.NewSplitStability(
			conflict.NewContributionSet(conflict.NewShiftContribution()),
			conflict.NewContributionSet(lessReduce, lessEqualReduce),
		)
		policy.ContributeSplitStability(conflict.MultiRejecterTestGrammarTerminalIdxLess, &onlyPotentialRejecters)
		Expect(onlyPotentialRejecters.IsSplitStable()).To(BeFalse())
	})

	// A terminal declared with a precedence but no associativity, like %precedence in GNU Bison, makes the precedence
	// comparison against its own production end in a tie which no associativity decides. The precedence policy then
	// leaves the conflict untouched, so what the policies behind it do determines the split stability.
	It("should leave a conflict on a terminal without associativity to the policies behind precedence", func() {
		tildeReduce := conflict.NewReduceContribution(conflict.MultiRejecterTestGrammarProductionIdxTilde)

		// The default policy resolves the leftover conflict with shift over reduce, which is anchored on the always
		// shift, so the shift dominates whether or not an isocore makes the reduction.
		underDefaultPolicy := conflict.NewSplitStability(
			conflict.NewContributionSet(conflict.NewShiftContribution()),
			conflict.NewContributionSet(tildeReduce),
		)
		conflict.DefaultPolicy(conflict.MultiRejecterTestGrammar).
			ContributeSplitStability(conflict.MultiRejecterTestGrammarTerminalIdxTilde, &underDefaultPolicy)
		Expect(underDefaultPolicy.IsSplitStable()).To(BeTrue())

		// The precedence policy alone leaves the potential reduction in the conflict, so splitting the state still
		// changes what the conflict is decided between.
		underPrecedenceAlone := conflict.NewSplitStability(
			conflict.NewContributionSet(conflict.NewShiftContribution()),
			conflict.NewContributionSet(tildeReduce),
		)
		conflict.PrecedencePolicy(conflict.MultiRejecterTestGrammar).
			ContributeSplitStability(conflict.MultiRejecterTestGrammarTerminalIdxTilde, &underPrecedenceAlone)
		Expect(underPrecedenceAlone.IsSplitStable()).To(BeFalse())
	})
})

// allContributionsOfGrammar returns all the actions which could ever compete for a terminal in the grammar: the shift
// of the terminal, and the reduction of any of the productions.
func allContributionsOfGrammar(grammar frontend.Grammar) []conflict.Contribution {
	result := make([]conflict.Contribution, 0, 1+len(grammar.Productions))
	result = append(result, conflict.NewShiftContribution())
	for productionIdx := range grammar.Productions {
		result = append(result, conflict.NewReduceContribution(productionIdx))
	}
	return result
}

// alwaysPotentialPartition is one way to split the contributions of a conflict into always contributions and potential
// contributions. A contribution which is in neither set is a never contribution, which no isocore makes.
type alwaysPotentialPartition struct {
	always     conflict.ContributionSet
	potentials conflict.ContributionSet
}

// allAlwaysPotentialPartitions returns every way the contributions can be split into always contributions, potential
// contributions, and never contributions. Each contribution independently is one of the three, so this enumerates all
// 3^n assignments by counting in base three.
func allAlwaysPotentialPartitions(contributions []conflict.Contribution) []alwaysPotentialPartition {
	total := 1
	for range contributions {
		total *= 3
	}

	result := make([]alwaysPotentialPartition, 0, total)
	for code := range total {
		var partition alwaysPotentialPartition
		remaining := code
		for _, contribution := range contributions {
			switch remaining % 3 {
			case 1:
				partition.always.Add(contribution)
			case 2:
				partition.potentials.Add(contribution)
			default:
				// A never contribution is left out of both sets.
			}
			remaining /= 3
		}
		result = append(result, partition)
	}
	return result
}

// isSplitStableByEnumeration decides split stability the brute force way of definition 3.35: the dominant contribution
// of the always contributions together with every subset of the potential contributions must be the same. This is the
// exponential computation the analytic bookkeeping is meant to avoid, which makes it the ground truth to check that
// bookkeeping against.
func isSplitStableByEnumeration(
	policy conflict.Policy,
	terminalIdx int,
	always conflict.ContributionSet,
	potentials conflict.ContributionSet,
) bool {
	var potentialList []conflict.Contribution
	for _, contribution := range potentials.All() {
		potentialList = append(potentialList, contribution)
	}

	reference := dominantContributionOfSubset(policy, terminalIdx, always, potentialList, (1<<len(potentialList))-1)
	for subset := range 1 << len(potentialList) {
		if !dominantContributionOfSubset(policy, terminalIdx, always, potentialList, subset).Equal(reference) {
			return false
		}
	}
	return true
}

// dominantContributionOfSubset computes the dominant contribution the policy decides for the always contributions
// together with the potential contributions selected by the subset bitmask.
func dominantContributionOfSubset(
	policy conflict.Policy,
	terminalIdx int,
	always conflict.ContributionSet,
	potentialList []conflict.Contribution,
	subset int,
) conflict.Decision {
	candidates := always.Clone()
	for i, contribution := range potentialList {
		if subset&(1<<i) != 0 {
			candidates.Add(contribution)
		}
	}
	return conflict.DominantContribution(policy, terminalIdx, candidates)
}
