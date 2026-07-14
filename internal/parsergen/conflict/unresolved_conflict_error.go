package conflict

import (
	"fmt"
)

// UnresolvedConflictError reports a conflict which the policies did not decide, so the state is left with more than one
// action for the conflicted terminal. A parser cannot be generated from such a state, because it would not know which
// of the actions to take, which is why an unresolved conflict is an error and not just something to report.
//
// A conflict is left unresolved when the policies have nothing to say about it. That happens when the grammar is
// ambiguous in a way the precedence declarations do not cover, and it happens when the grammar author composed a policy
// on purpose which has no rule of last resort, so that every conflict has to be decided by an explicit precedence
// declaration instead of by a default like shift over reduce.
type UnresolvedConflictError struct {
	// Conflict is the conflict which was left unresolved, including the decision which says which contributions it was
	// left undecided between.
	Conflict Conflict

	// TerminalName is the name of the conflicted terminal, so that the error can be read without having the grammar at
	// hand.
	TerminalName string
}

// UnresolvedConflictError implements error.
var _ error = (*UnresolvedConflictError)(nil)

// Error returns the error message.
func (e UnresolvedConflictError) Error() string {
	return fmt.Sprintf(
		"state %d is undecided between %s on terminal %s",
		e.Conflict.StateIdx,
		e.Conflict.Decision.Unresolved.String(),
		e.TerminalName,
	)
}
