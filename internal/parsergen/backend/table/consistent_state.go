package table

// NewConsistentStates returns, for every state, whether it is consistent: it reduces by the same production whatever
// the lookahead is. Such a state needs no lookahead to decide, which is what lets the lookahead correction of a
// generated parser skip its check there, see section 3.5.2 "Parser" of "PSLR(1): Pseudo-Scannerless Minimal LR(1) for
// the Deterministic Parsing of Composite Languages" by Joel E. Denny.
//
// A state is consistent when its default action is a reduce and its row of the action table has no entry which
// differs from it. The accept is not a reduce here, because the check only ever runs before a reduction.
func NewConsistentStates(actionRows [][]int, defaultActionByStateIdx []Action) []bool {
	result := make([]bool, len(actionRows))
	for stateIdx, row := range actionRows {
		result[stateIdx] = isConsistent(row, defaultActionByStateIdx[stateIdx])
	}
	return result
}

// isConsistent returns whether a single state with the given row and default action reduces by the same production on
// every lookahead.
func isConsistent(row []int, defaultAction Action) bool {
	if defaultAction == NoAction || defaultAction.Kind() != ActionKindReduce {
		return false
	}
	for _, action := range row {
		if action != int(NoAction) && action != int(defaultAction) {
			return false
		}
	}
	return true
}
