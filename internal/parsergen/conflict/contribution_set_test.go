package conflict_test

import (
	"github.com/backbone81/golr/internal/parsergen/conflict"
)

// allPossibleContributionSets returns every set of contributions which could ever compete for a terminal in
// PrecedenceTestGrammar, so that a test can check a property of a policy on all of them instead of arguing about it.
//
// The actions which can compete for a terminal are the shift of that terminal and the reduction of any of the
// productions, and any combination of them is a conceivable conflict. The empty set is left out, because it is not a
// conflict: the dominant contribution is undefined when there is nothing to decide about.
func allPossibleContributionSets() []conflict.ContributionSet {
	return nonEmptySubsets(allPossibleContributions())
}

// allPossibleContributions returns all the actions which could ever compete for a terminal in PrecedenceTestGrammar:
// the shift of the terminal, and the reduction of any of its productions.
func allPossibleContributions() []conflict.Contribution {
	return []conflict.Contribution{
		conflict.NewShiftContribution(),
		conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxPlus),
		conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxTimes),
		conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxCompare),
		conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxPower),
		conflict.NewReduceContribution(conflict.PrecedenceTestGrammarProductionIdxIdentity),
	}
}

// nonEmptySubsets returns every non-empty subset of the contributions. The bits of the counter select which of the
// contributions are part of the subset, so counting from one to the last bit pattern enumerates all of them.
func nonEmptySubsets(contributions []conflict.Contribution) []conflict.ContributionSet {
	var result []conflict.ContributionSet
	for bits := 1; bits < 1<<len(contributions); bits++ {
		var subset conflict.ContributionSet
		for i, contribution := range contributions {
			if bits&(1<<i) != 0 {
				subset.Add(contribution)
			}
		}
		result = append(result, subset)
	}
	return result
}
