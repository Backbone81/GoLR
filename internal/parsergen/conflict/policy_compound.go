package conflict

import (
	"errors"

	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

// CompoundPolicy composes the policies of the factories into a single policy, which resolves a conflict by applying
// them in order until a single contribution is left. This is how the rules by which conflicts are resolved are put
// together: a rule whose factory is not part of the composition is not applied. The order matters, because an earlier
// policy decides a conflict which a later policy would have decided differently.
//
// A compound policy does not have to decide every conflict. A composition which narrows a conflict down to more than
// one contribution leaves the conflict unresolved, which DominantContribution reports as such. That is how a grammar
// author insists on explicit precedence declarations: without ShiftOverReducePolicy in the composition, a shift/reduce
// conflict which no declaration decides is reported instead of silently going to the shift. Ending the composition with
// a policy which always decides, like EarliestProductionPolicy, is what guarantees that every reduce/reduce conflict is
// decided. A composition of no factories at all decides nothing, which NullPolicy says more plainly.
//
// The composition is itself a PolicyFactory, so it can be handed to a core as it is, and it stays grammar independent
// until it is applied. Every application makes its own policies, which matters because a policy may keep state across
// the calls of a single build, see PolicyFactory.
func CompoundPolicy(policyFactories ...PolicyFactory) PolicyFactory {
	return func(augmentedGrammar frontend.Grammar) Policy {
		result := make(compoundPolicy, 0, len(policyFactories))
		for _, policyFactory := range policyFactories {
			result = append(result, policyFactory(augmentedGrammar))
		}
		return result
	}
}

// compoundPolicy is the applied form of a composition: the policies of CompoundPolicy, in the order they are applied
// in, all made from the same augmented grammar. See CompoundPolicy for what composing them means.
type compoundPolicy []Policy

// compoundPolicy implements Policy.
var _ Policy = (compoundPolicy)(nil)

// Resolve applies the policies of the compound policy in order, and stops as soon as a single contribution is left or
// every contribution was removed.
func (p compoundPolicy) Resolve(terminalIdx int, candidates ContributionSet) ContributionSet {
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

// ContributeSplitStability lets each policy of the compound narrow the same bookkeeping in order. That shared
// bookkeeping is what makes the split stability of a compound policy decidable at all: a later policy narrows what the
// earlier policies left in remaining, so the policies together account for interactions which none of them sees on its
// own, like a reduction which precedence removes never reaching the reduce/reduce comparison behind it.
//
// Unlike Resolve, this does not stop once a single contribution is left. A policy which decides between reductions
// still has to weigh in on whether the surviving contribution is an always contribution, and it does so by inspecting
// the remaining contribution even when it is the only one left. Replaying the policies on candidates which the early
// stopping Resolve would never hand them stays faithful to Resolve because the Policy contract demands that a Resolve
// returns one or no candidate unchanged, see Policy.
func (p compoundPolicy) ContributeSplitStability(terminalIdx int, splitStability *SplitStability) {
	for _, policy := range p {
		policy.ContributeSplitStability(terminalIdx, splitStability)
	}
}
