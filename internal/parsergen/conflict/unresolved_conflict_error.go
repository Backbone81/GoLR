package conflict

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

	// message names the conflicted terminal and state, see Error.
	message string
}

// UnresolvedConflictError implements error.
var _ error = (*UnresolvedConflictError)(nil)

// Error returns a single line naming the conflicted terminal and state. The full report of the conflict is written by
// the report package.
func (e UnresolvedConflictError) Error() string {
	return e.message
}

// UnresolvedConflictErrors returns every UnresolvedConflictError in the error tree, in the order they were joined.
// errors.As is no help here, because it stops at the first match, while Resolve joins one error per unresolved
// conflict.
func UnresolvedConflictErrors(err error) []UnresolvedConflictError {
	//nolint:errorlint // errors.As stops at the first match, but every error of a join is needed. The recursion unwraps.
	switch typedErr := err.(type) {
	case UnresolvedConflictError:
		return []UnresolvedConflictError{typedErr}
	case interface{ Unwrap() []error }:
		var result []UnresolvedConflictError
		for _, wrappedErr := range typedErr.Unwrap() {
			result = append(result, UnresolvedConflictErrors(wrappedErr)...)
		}
		return result
	case interface{ Unwrap() error }:
		return UnresolvedConflictErrors(typedErr.Unwrap())
	}
	return nil
}
