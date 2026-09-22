// Package conflict provides the public API re-exports for the conflicts a parser core reports.
package conflict

import (
	intconflict "github.com/backbone81/golr/internal/parsergen/conflict"
)

type (
	// Conflict describes a single conflicted terminal of a state, together with what the policy decided about it. This
	// is what a parser generator reports to the user as a shift/reduce or a reduce/reduce conflict.
	Conflict = intconflict.Conflict

	// Contribution is a single action which a state can take on a terminal. When a state has more than one contribution
	// for the same terminal, those contributions are in conflict with each other.
	Contribution = intconflict.Contribution
	// ContributionSet is the set of contributions which compete for the same terminal, which is what a conflict is made
	// of.
	ContributionSet = intconflict.ContributionSet

	// Decision is what the policy decided about a set of contributions.
	Decision = intconflict.Decision

	// ReportConfig controls what WriteConflictReport and ConflictReport.Write write.
	ReportConfig = intconflict.ReportConfig

	// ConflictReport is the report of the conflicts of a single state, rendered with the names of the grammar symbols
	// so it can be written without the grammar at hand.
	ConflictReport = intconflict.ConflictReport

	// ConflictReportEntry is the report of a single conflicted terminal of a state.
	ConflictReportEntry = intconflict.ConflictReportEntry

	// ConflictReportReduction is a reduction of a conflicted state together with the lookaheads which tell the state
	// apart from the other states with the same kernel items.
	ConflictReportReduction = intconflict.ConflictReportReduction

	// UnresolvedConflictError reports a conflict which the policies did not decide. A core joins one of them per
	// unresolved conflict into the error it returns.
	UnresolvedConflictError = intconflict.UnresolvedConflictError
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

// UnresolvedConflictErrors returns every UnresolvedConflictError in the error tree, in the order they were joined. A
// core joins one of them per unresolved conflict into the error it returns.
var UnresolvedConflictErrors = intconflict.UnresolvedConflictErrors

// WriteConflictReport writes a report of the given conflicts to w. Conflicts the policy resolved on its own are
// summarized and listed in full only with ReportConfig.Verbose. Conflicts decided by precedence declarations are not
// reported, and conflicts the policy could not decide are reported by WriteUnresolvedConflictReport.
var WriteConflictReport = intconflict.WriteConflictReport

// WriteUnresolvedConflictReport writes the report of a core which failed on unresolved conflicts: the summary of the
// conflicts it returned, followed by the report of every UnresolvedConflictError in the error.
var WriteUnresolvedConflictReport = intconflict.WriteUnresolvedConflictReport
