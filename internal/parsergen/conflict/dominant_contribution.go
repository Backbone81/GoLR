package conflict

// DominantContribution returns what the policy decided about the conflict on the terminal. This is the dominant
// contribution function of definition 2.19 of IELR(1), which the paper calls delta, with the rejection of the terminal
// and the unresolved conflict as additional outcomes, see Decision.
//
// The contributions are the actions which compete for the terminal. Note that the state those actions come from is
// deliberately not part of the input: phase 3 of IELR(1) evaluates the dominant contribution on hypothetical sets of
// contributions, which are the contributions a state would make if its kernel items had different lookahead sets. Such
// a set of contributions does not exist as the actions of any state.
//
// Definition 2.5 of the paper requires the conflict resolution to select a unique action from every conflict, which a
// policy is free not to do: when it narrows the conflict down to more than one contribution, the conflict stands, and
// the decision reports it as unresolved together with the contributions which are left. Phase 3 of IELR(1) only ever
// compares decisions with each other, and an unresolved decision takes part in that comparison like any other: it is
// equal exactly to an unresolved decision whose conflict was left with the same contributions.
func DominantContribution(policy Policy, terminalIdx int, contributions ContributionSet) Decision {
	if contributions.IsEmpty() {
		// There is no contribution to decide about, so there is no dominant contribution.
		return NewUndefinedDecision()
	}

	remaining := policy.Resolve(terminalIdx, contributions)
	if remaining.IsEmpty() {
		// A policy removed every action on purpose, so the parser rejects the terminal in this state.
		return NewErrorDecision()
	}
	if remaining.Length() > 1 {
		// The policy could not narrow the conflict down to a single contribution, so the conflict stands.
		return NewUnresolvedDecision(remaining)
	}
	return NewDominantDecision(remaining.GetByIndex(0))
}
