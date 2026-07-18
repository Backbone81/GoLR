// Package conflict provides the public API re-exports for the conflicts a parser core reports.
package conflict

import intconflict "github.com/backbone81/golr/internal/parsergen/conflict"

type (
	// Conflict describes a single conflicted terminal of a state, together with what the policy decided about it. This
	// is what a parser generator reports to the user as a shift/reduce or a reduce/reduce conflict.
	Conflict = intconflict.Conflict

	// Contribution is a single action which a state can take on a terminal. When a state has more than one contribution
	// for the same terminal, those contributions are in conflict with each other.
	Contribution = intconflict.Contribution

	// Decision is what the policy decided about a set of contributions.
	Decision = intconflict.Decision
)

const (
	// DecisionUndefined means that there was no contribution to decide about.
	DecisionUndefined = intconflict.DecisionUndefined

	// DecisionDominant means that a single contribution won, which is the dominant contribution.
	DecisionDominant = intconflict.DecisionDominant

	// DecisionError means that a policy removed every action for the terminal on purpose, so the parser rejects the
	// terminal in this state.
	DecisionError = intconflict.DecisionError

	// DecisionUnresolved means that the policy could not narrow the conflict down to a single contribution, so the
	// conflict stands.
	DecisionUnresolved = intconflict.DecisionUnresolved
)
