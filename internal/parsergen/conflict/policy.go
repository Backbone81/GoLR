package conflict

import (
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// Policy is a single rule by which a conflict is resolved. A policy narrows the contributions which compete for the
// conflicted terminal down to those which survive the rule, and it returns the candidates unchanged when the rule does
// not apply to them. Narrowing the candidates down to a single contribution makes that contribution the dominant one.
//
// A policy must not modify the candidates it is given, because the caller keeps them around: the set of contributions
// which are in conflict is what a conflict is reported with. It returns a new set instead.
//
// Removing every candidate is how a policy asks for the conflicted terminal to be rejected in this state, which is what
// a terminal declared as nonassociative needs. A policy which does not want that must never return an empty set for a
// non-empty input.
//
// A conflict which is already decided must be left alone: Resolve must return candidates of one or no contribution
// unchanged. A compound policy stops calling Resolve as soon as so few candidates are left, but its split stability
// bookkeeping keeps replaying every policy on whatever remains, see CompoundPolicy. A Resolve which narrowed a decided
// conflict further would make the bookkeeping diverge from what Resolve actually decides, and the split stability
// verdict would be about a decision the policies never make.
type Policy interface {
	Resolve(terminalIdx int, candidates ContributionSet) ContributionSet

	// ContributeSplitStability narrows the split stability bookkeeping the same way Resolve narrows its candidates, and
	// records in the bookkeeping whether any narrowing it made depended on a potential contribution being present. It is
	// how a policy takes part in deciding the general case of definition 3.35 of IELR(1), so that phase 2 can discard the
	// annotations whose dominant contribution splitting a state cannot change. See SplitStability for the whole picture.
	ContributeSplitStability(terminalIdx int, splitStability *SplitStability)
}

// PolicyFactory makes the policy for a grammar. A policy which reads the grammar has to be bound to one, and the
// grammar it has to be bound to is the augmented grammar the parser tables are built from, not the grammar a frontend
// produced: AugmentGrammar inserts the end of input terminal and the new start nonterminal at index 0 and shifts every
// terminal, nonterminal and production index of the original grammar by one. A policy bound to the unaugmented grammar
// would read the precedence of the neighbouring symbol for every index it looks up, which is wrong without looking
// wrong.
//
// Augmenting the grammar is the business of the cores, so the caller is not in a position to bind the policy itself. It
// hands over a factory instead, which the core calls once on the augmented grammar. That keeps a policy composition
// grammar independent until it is applied, and it gives every build its own policy instances, which matters because a
// policy may keep state across the calls of a single build, see PrecedencePolicy.
//
// A factory is called once per build, so it is also the place for a policy to precompute whatever it wants to derive
// from the grammar.
type PolicyFactory func(augmentedGrammar frontend.Grammar) Policy

// DefaultPolicy returns the compound policy which resolves conflicts the way GNU Bison and Yacc do: precedence and
// associativity decide first, a shift beats a reduction when precedence has nothing to say, and the production which
// was declared first wins a conflict between two reductions.
//
// This is a PolicyFactory, so it can be handed to a core as it is, and the grammar MUST be an augmented grammar.
//
//nolint:ireturn // Returning the interface is what makes this usable as a PolicyFactory.
func DefaultPolicy(augmentedGrammar frontend.Grammar) Policy {
	return CompoundPolicy(
		PrecedencePolicy,
		ShiftOverReducePolicy,
		EarliestProductionPolicy,
	)(augmentedGrammar)
}

// DefaultPolicy is a PolicyFactory.
var _ PolicyFactory = DefaultPolicy
