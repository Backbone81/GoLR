package conflict

import (
	"errors"
	"fmt"

	"github.com/backbone81/golr/internal/utils"
)

// DecisionKind describes what the dominant contribution function decided about a set of contributions.
//
// The first three values are what the dominant contribution function of definition 3.42 of IELR(1) can take, and they
// match the three shapes of the ContributionIndex which GNU Bison computes: a contribution, no contribution, and the
// error action. The fourth value is our extension for a policy which does not decide every conflict, which GNU Bison
// never needs because its policy always decides.
type DecisionKind int

const (
	// DecisionUndefined means that there was no contribution to decide about. This is point 2 of definition 3.42 of
	// IELR(1), where the dominant contribution is undefined because the state makes no contribution to the conflict at
	// all.
	//
	// It is not the same as DecisionError, and the state compatibility test of definition 3.43 relies on the difference:
	// a state which makes no contribution says nothing about the terminal and is compatible with anything, while a state
	// whose actions were all removed decided to reject the terminal and only stays compatible with a state which decided
	// the same.
	DecisionUndefined DecisionKind = iota

	// DecisionDominant means that a single contribution won, which is the dominant contribution. This is point 1 of
	// definition 3.42 of IELR(1).
	DecisionDominant

	// DecisionError means that a policy removed every action for the terminal on purpose, so the parser rejects the
	// terminal in this state. This is what a terminal declared as nonassociative asks for.
	//
	// The paper has no such value, because the dominant contribution function of definition 2.19 always returns one of
	// the contributions it is given. It models the rejection outside of that function instead, by rewriting the action
	// as the empty action of point 4 of definition 2.9. A parser which has no action for a terminal reports a syntax
	// error, see definition 2.11.
	DecisionError

	// DecisionUnresolved means that the policy could not narrow the conflict down to a single contribution, so the
	// conflict stands. The contributions the conflict was left with are part of the decision, see Survivors, and their
	// actions all stay in the parser tables.
	//
	// The paper has no such value either, because definition 2.5 requires the conflict resolution to select a unique
	// action from every conflict. It is what a policy without a rule of last resort produces, which is how a grammar
	// author insists on deciding every conflict with explicit precedence declarations: a conflict which stays
	// unresolved is reported instead of being decided by a default like shift over reduce.
	DecisionUnresolved
)

// Decision is what the dominant contribution function decided about a set of contributions. This is the value of the
// dominant contribution function of definition 3.42 of IELR(1), extended by the rejection of a terminal and by the
// unresolved conflict, which the paper does not express as values of that function, see the decision kinds.
//
// Two decisions are compared with Equal, which is how the state compatibility test of definition 3.43 of IELR(1)
// compares what two states decide.
type Decision struct {
	// Kind describes which of the four possible decisions this is.
	Kind DecisionKind

	// Dominant is the contribution which won. It is only meaningful when Kind is DecisionDominant, and it is the zero
	// value otherwise, so that two decisions of the same kind compare as equal.
	Dominant Contribution

	// Unresolved holds the contributions an unresolved conflict was left with. It is only meaningful when Kind is
	// DecisionUnresolved, and it is the empty set otherwise, so that two decisions of the same kind compare as equal.
	Unresolved ContributionSet
}

// NewUndefinedDecision creates the decision for a state which makes no contribution to the conflict.
func NewUndefinedDecision() Decision {
	return Decision{
		Kind: DecisionUndefined,
	}
}

// NewDominantDecision creates the decision for a conflict which the policies resolved in favor of a single
// contribution.
func NewDominantDecision(contribution Contribution) Decision {
	return Decision{
		Kind:     DecisionDominant,
		Dominant: contribution,
	}
}

// NewErrorDecision creates the decision for a conflict which the policies resolved by removing every action.
func NewErrorDecision() Decision {
	return Decision{
		Kind: DecisionError,
	}
}

// NewUnresolvedDecision creates the decision for a conflict which the policies could not narrow down to a single
// contribution. The remaining contributions are the ones the conflict was left with, which is at least two of them: a
// single remaining contribution is a dominant decision, and no remaining contribution at all is an error decision.
func NewUnresolvedDecision(remaining ContributionSet) Decision {
	utils.DebugAssert(func() error {
		if remaining.Length() < 2 {
			return errors.New("an unresolved decision needs at least two contributions to be undecided between")
		}
		return nil
	})
	return Decision{
		Kind: DecisionUnresolved,
		// The set is cloned because it can share its storage with the contributions of the conflict, and the decision
		// has to stay what it was even when those are modified later.
		Unresolved: remaining.Clone(),
	}
}

// Equal reports if both decisions decided the same. Two unresolved decisions only decided the same when their
// conflicts were left with the same contributions, because a state which is undecided between different actions
// behaves differently.
func (d Decision) Equal(other Decision) bool {
	return d.Kind == other.Kind &&
		d.Dominant == other.Dominant &&
		d.Unresolved.Equal(&other.Unresolved)
}

// Survivors returns the contributions which keep their actions under this decision: the contribution which won when
// there is a dominant one, the contributions the conflict was left with when it is unresolved, and no contribution at
// all when the terminal is rejected or when there was nothing to decide about.
func (d Decision) Survivors() ContributionSet {
	switch d.Kind {
	case DecisionDominant:
		// The contribution which won is the only one whose action stays.
		return NewContributionSet(d.Dominant)
	case DecisionUnresolved:
		// The conflict stands, so the actions of every contribution it was left undecided between stay.
		return d.Unresolved
	case DecisionError:
		// The terminal is rejected, so no contribution survives and every action for the terminal is removed.
		return ContributionSet{}
	case DecisionUndefined:
		// There was no contribution to decide about, so there is none which could survive either.
		return ContributionSet{}
	}
	return ContributionSet{}
}

// Decision implements fmt.Stringer.
var _ fmt.Stringer = (*Decision)(nil)

// String returns a string representation.
func (d Decision) String() string {
	switch d.Kind {
	case DecisionUndefined:
		return "undefined"
	case DecisionDominant:
		return d.Dominant.String()
	case DecisionError:
		return "error"
	case DecisionUnresolved:
		return "unresolved between " + d.Unresolved.String()
	}
	return "unknown"
}
