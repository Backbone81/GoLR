package conflict

import (
	"errors"

	"github.com/backbone81/golr/internal/utils"
)

// CompoundPolicy resolves a conflict by applying its policies in order until a single contribution is left. This is how
// the rules by which conflicts are resolved are composed: a rule which is not part of the compound policy is not
// applied. The order matters, because an earlier policy decides a conflict which a later policy would have decided
// differently.
//
// A compound policy does not have to decide every conflict. A composition which narrows a conflict down to more than
// one contribution leaves the conflict unresolved, which DominantContribution reports as such. That is how a grammar
// author insists on explicit precedence declarations: without NewShiftOverReducePolicy in the composition, a
// shift/reduce conflict which no declaration decides is reported instead of silently going to the shift. Ending the
// compound policy with a policy which always decides, like EarliestProductionPolicy, is what guarantees that every
// reduce/reduce conflict is decided.
type CompoundPolicy []Policy

// CompoundPolicy implements Policy.
var _ Policy = (CompoundPolicy)(nil)

// Resolve applies the policies of the compound policy in order, and stops as soon as a single contribution is left or
// every contribution was removed.
func (p CompoundPolicy) Resolve(terminalIdx int, candidates ContributionSet) ContributionSet {
	result := candidates
	for _, policy := range p {
		if result.Length() <= 1 {
			// The conflict is decided, so there is nothing left for the remaining policies to decide.
			return result
		}
		result = policy.Resolve(terminalIdx, result)
		utils.DebugAssert(func() error {
			if result.Length() > candidates.Length() {
				return errors.New("a policy is expected to narrow down the candidates, not to add to them")
			}
			return nil
		})
	}
	return result
}
