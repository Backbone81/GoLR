package table

import (
	"fmt"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/utils"
)

// NoGoto is the entry of the goto table for a state which has no goto on a nonterminal. State indexes are never
// negative, so this can not be confused with one.
const NoGoto = -1

// NewGotoTable lays the gotos of the given parser out as one row per state, indexed by nonterminal. This dense table is
// the form the row displacement compression consumes. An entry is the index of the state the goto leads to.
//
// A goto is what the parser reads after a reduction, when it has popped the right hand side of a production off its
// stack and asks where the nonterminal on the left hand side takes it. Only the states which such a reduction can
// uncover have a goto on a given nonterminal, so the rows are sparse, which is what the row displacement packs. Sparser
// still once the default goto of every nonterminal has been taken out of them, see ApplyDefaultGotos.
//
// Unlike the action table, the goto table needs no entry which stands for an error. A reduction only ever asks for a
// nonterminal in a state whose goto the LR construction created together with the production. An entry which is missing
// there is a defect of the parser tables and not a syntax error of the input. The one nonterminal without a single goto
// is `$accept` of the augmented grammar, because it never appears on the right hand side of a production; its column
// stays empty everywhere and costs nothing once the rows are packed.
func NewGotoTable(parser backend.Parser) [][]int {
	rows := make([][]int, len(parser.States))
	for stateIdx := range parser.States {
		row := make([]int, len(parser.Grammar.Nonterminals))
		for nonterminalIdx := range row {
			row[nonterminalIdx] = NoGoto
		}
		for _, transitionAction := range parser.States[stateIdx].TransitionActions.All() {
			if transitionAction.SymbolRef().IsTerminal() {
				// A transition on a terminal is a shift, which lives in a table of its own, see NewActionTable.
				continue
			}
			setGoto(row, stateIdx, transitionAction.SymbolRef().Idx(), transitionAction.StateIdx())
		}
		rows[stateIdx] = row
	}
	return rows
}

// setGoto writes a single goto into a row of the goto table.
//
// Two gotos on the same nonterminal would leave the parser with a choice to make where the LR construction guarantees
// it has none, so this is a defect of the table construction rather than something a grammar can provoke, which is what
// makes it a debug assertion.
func setGoto(row []int, stateIdx int, nonterminalIdx int, targetStateIdx int) {
	utils.DebugAssert(func() error {
		if row[nonterminalIdx] != NoGoto && row[nonterminalIdx] != targetStateIdx {
			return fmt.Errorf(
				"state %d gotos to state %d and to state %d on nonterminal %d",
				stateIdx, row[nonterminalIdx], targetStateIdx, nonterminalIdx,
			)
		}
		return nil
	})
	row[nonterminalIdx] = targetStateIdx
}
