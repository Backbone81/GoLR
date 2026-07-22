package conflict

import "github.com/backbone81/golr/internal/parsergen/frontend"

// NullPolicy returns the policy which decides nothing, so that every conflict of the grammar stands as it is. It is
// what an empty CompoundPolicy amounts to, said plainly, and the grammar is only taken to meet PolicyFactory.
//
// This is not a policy to generate a parser with: Resolve reports every conflict as unresolved, so a core fails on any
// grammar which has one. It is the policy for looking at the conflicts a grammar really has, which is what the oracle
// and differential testing work compares the cores on, and what the conflict-preserving tests of phase 2 of IELR(1)
// rely on.
//
// Handing it to the IELR(1) core is more than leaving phase 5 undecided, because phase 3 splits the states with the
// policy too: a policy which narrows nothing leaves every potential contribution in the split stability bookkeeping, so
// no annotation is ever discarded and phase 3 splits wherever the lookaheads differ at all. That is the conservative
// end of the algorithm - correct, but with tables which grow towards canonical LR(1) instead of hugging LALR(1).
//
//nolint:ireturn // Returning the interface is what makes this usable as a PolicyFactory.
func NullPolicy(augmentedGrammar frontend.Grammar) Policy {
	return &nullPolicy{}
}

// NullPolicy is a PolicyFactory.
var _ PolicyFactory = NullPolicy

// nullPolicy leaves every conflict undecided. See NullPolicy for what that is good for.
type nullPolicy struct{}

// nullPolicy implements Policy.
var _ Policy = (*nullPolicy)(nil)

// Resolve returns the candidates unchanged, so no contribution ever wins a conflict.
func (p *nullPolicy) Resolve(terminalIdx int, candidates ContributionSet) ContributionSet {
	return candidates
}

// ContributeSplitStability leaves the bookkeeping alone, because a policy which narrows nothing cannot narrow anything
// in a way which depends on a potential contribution being present. The general case of definition 3.35 of IELR(1)
// collapses to observation 3.33 then: the dominant contribution is split-stable exactly when the conflict has no
// potential contribution, which IsSplitStable reports from the untouched bookkeeping on its own.
func (p *nullPolicy) ContributeSplitStability(terminalIdx int, splitStability *SplitStability) {
}
