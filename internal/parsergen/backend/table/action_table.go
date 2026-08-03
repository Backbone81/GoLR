package table

import (
	"fmt"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/utils"
)

// NewActionTable lays the actions of the given parser out as one row per state, indexed by terminal. This dense table
// is the form the row displacement compression consumes. An entry is an Action.
//
// A terminal which the state has no action for is left as NoAction, so that the row displacement has holes to place the
// other rows in. An empty entry does not mean the terminal is an error: the parser takes the default action of the
// state for it, which is what makes the rows sparse enough to be worth packing in the first place, see
// NewDefaultActions.
//
// A terminal which the state rejects on purpose is the one syntax error which does get an entry of its own. It has to,
// because it has to beat the default action of the state, and an absent entry would fall through to it.
func NewActionTable(parser backend.Parser) [][]int {
	errorIdx := errorTerminalIdx(parser.Grammar)

	rows := make([][]int, len(parser.States))
	for stateIdx := range parser.States {
		row := make([]int, len(parser.Grammar.Terminals))
		for terminalIdx := range row {
			row[terminalIdx] = int(NoAction)
		}
		fillActionRow(row, stateIdx, &parser.States[stateIdx], errorIdx)
		rows[stateIdx] = row
	}
	return rows
}

// fillActionRow writes the actions of a single state into its row of the action table.
func fillActionRow(row []int, stateIdx int, state *backend.State, errorIdx int) {
	for terminalIdx := range state.RejectedTerminals.All() {
		if terminalIdx == errorIdx {
			// No scanner ever delivers the error symbol, the parser shifts it itself while recovering, so an entry
			// which rejects it could never be read. Its column carries the shift the recovery looks up instead, see
			// CompressedParser.ErrorShiftStateIdx. A grammar reaches this by declaring the error symbol
			// nonassociative, which is what makes leaving the entry out a decision rather than an impossible case.
			assertNoErrorShift(stateIdx, state, errorIdx)
			continue
		}
		setAction(row, stateIdx, terminalIdx, NewErrorAction())
	}

	for _, reduceAction := range state.ReduceActions.All() {
		for terminalIdx := range reduceAction.LookaheadSet.All() {
			setAction(row, stateIdx, terminalIdx, NewReduceAction(reduceAction.ProductionIdx))
		}
	}

	for _, transitionAction := range state.TransitionActions.All() {
		if transitionAction.SymbolRef().IsNonterminal() {
			// A transition on a nonterminal is a goto, which lives in a table of its own, see NewGotoTable.
			continue
		}
		setAction(row, stateIdx, transitionAction.SymbolRef().Idx(), NewShiftAction(transitionAction.StateIdx()))
	}
}

// assertNoErrorShift checks that a state which rejects the error symbol does not shift it as well.
//
// This is the invariant which lets fillActionRow leave the rejection out of the row. Rejecting a terminal means every
// action for it was removed, see conflict.resolveState, so the shift is gone along with the rest. Were both present,
// dropping the rejection would leave the shift standing in the error column and the error recovery would resume in a
// state which the conflict resolution ruled out, instead of setAction reporting the two actions as the contradiction
// they are.
func assertNoErrorShift(stateIdx int, state *backend.State, errorIdx int) {
	utils.DebugAssert(func() error {
		for _, transitionAction := range state.TransitionActions.All() {
			symbolRef := transitionAction.SymbolRef()
			if symbolRef.IsTerminal() && symbolRef.Idx() == errorIdx {
				return fmt.Errorf(
					"state %d rejects the error symbol and shifts it to state %d",
					stateIdx, transitionAction.StateIdx(),
				)
			}
		}
		return nil
	})
}

// setAction writes a single action into a row of the action table.
//
// Two different actions for one terminal are a conflict which was never decided. Every conflict is decided by the time
// a parser reaches a backend, see the GrammarToParser interface of the cores, so this is a defect of the table
// construction rather than something a grammar can provoke, which is what makes it a debug assertion. Writing the same
// action twice is allowed, because a state may carry a reduce action explicitly which its default action covers as
// well.
func setAction(row []int, stateIdx int, terminalIdx int, action Action) {
	utils.DebugAssert(func() error {
		if row[terminalIdx] != int(NoAction) && row[terminalIdx] != int(action) {
			return fmt.Errorf(
				"state %d has the actions %s and %s for terminal %d",
				stateIdx, Action(row[terminalIdx]), action, terminalIdx,
			)
		}
		return nil
	})
	row[terminalIdx] = int(action)
}
