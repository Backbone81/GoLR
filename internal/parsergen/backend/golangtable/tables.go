package golangtable

import (
	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/backend/table"
	"github.com/backbone81/golr/internal/utils"
)

// Tables holds the lookup tables of a table driven parser in the form the template writes them out.
type Tables struct {
	// TerminalColumnByToken translates a token the scanner delivers into the column of the action table which holds
	// the decisions for it. A token which is not a terminal of this grammar is not in here and gets
	// NoTerminalColumn, see TokenColumn.
	TerminalColumnByToken []TokenColumn

	// ActionBase maps a state to the displacement of its row within ActionNext.
	ActionBase utils.IntArray

	// ActionNext holds the actions of all states packed into a single table. The action a state has for a terminal
	// lives at ActionBase[state] + column, if ActionCheck confirms that the cell belongs to that column.
	ActionNext utils.IntArray

	// ActionCheck holds the column every cell of ActionNext belongs to, or NoColumn for a cell which no state
	// occupies. It has the same entry type as TerminalColumnByToken, so that the driver can compare the two without
	// a conversion.
	ActionCheck utils.IntArray

	// NoColumn is the entry ActionCheck uses for a cell which no state occupies. It is one past the highest column
	// in use, so no lookup can ever ask for it.
	NoColumn int

	// NoTerminalColumn is the column a token gets which is no terminal of this grammar. It holds no entry in any
	// state, so a lookup falls through to the default action of the state.
	NoTerminalColumn int

	// DefaultActionByState holds, for every state, the action it takes for a terminal ActionNext has no entry for.
	// A state which has no default action carries the error action here, which makes such a terminal a syntax
	// error.
	DefaultActionByState utils.IntArray

	// GotoBase maps a state to the displacement of its row within GotoNext.
	GotoBase utils.IntArray

	// GotoNext holds the gotos of all states packed into a single table. The goto a state has for a nonterminal
	// lives at GotoBase[state] + nonterminal, if GotoCheck confirms that the cell belongs to that nonterminal.
	GotoNext utils.IntArray

	// GotoCheck holds the nonterminal every cell of GotoNext belongs to, or NoNonterminal for a cell which no state
	// occupies.
	GotoCheck utils.IntArray

	// NoNonterminal is the entry GotoCheck uses for a cell which no state occupies. It is one past the highest
	// nonterminal index, so no lookup can ever ask for it.
	NoNonterminal int

	// PopCountByProduction holds, for every production, the number of symbols a reduction by it takes off the
	// stacks. This is the length of the right hand side of the production and needs no compression, so it is read
	// off the grammar rather than built by the table package.
	PopCountByProduction utils.IntArray

	// NonterminalByProduction holds, for every production, the nonterminal on its left hand side. A reduction looks
	// up its goto with it and labels the node it pushes with it.
	NonterminalByProduction utils.IntArray

	// ErrorTerminalColumn is the column of the action table which holds the shifts of the error symbol, which is
	// where the error recovery reads the state to resume in.
	ErrorTerminalColumn int

	// HasErrorRecovery reports whether any state can shift the error symbol. The generated parser leaves out the
	// parts of the recovery which cost something on the hot path when no state can.
	HasErrorRecovery bool

	// ActionKindBits is the number of low bits of an action which hold what it does.
	ActionKindBits int

	// ActionKindMask selects out of an action what it does.
	ActionKindMask int

	// ActionKindShift is the action which shifts the terminal and continues in the state the action carries.
	ActionKindShift int

	// ActionKindReduce is the action which reduces by the production the action carries.
	ActionKindReduce int

	// ActionKindAccept is the action which ends the parse successfully.
	ActionKindAccept int

	// ActionKindError is the action which rejects the terminal as a syntax error.
	ActionKindError int
}

// NewTables compresses the given parser into the lookup tables the generated parser reads at runtime.
func NewTables(parser backend.Parser) Tables {
	compressed := table.NewCompressedParser(parser)

	// The check entries hold a column or the one value past them which means the cell is unused, so the column
	// table and the check table are typed by that value and can be compared against each other directly.
	maxColumn := table.TerminalColumn(len(parser.Grammar.Terminals) - 1)
	noColumn := maxColumn + 1
	columnType := utils.GoUintType(noColumn)

	nonterminalCount := len(parser.Grammar.Nonterminals)

	errorTerminalColumn := table.NoTerminalColumn
	if compressed.ErrorTerminalIdx != table.NoTerminal {
		errorTerminalColumn = table.TerminalColumn(compressed.ErrorTerminalIdx)
	}

	return Tables{
		TerminalColumnByToken: newTerminalColumnByToken(parser),

		ActionBase:  utils.NewIntArray(compressed.Actions.Base),
		ActionNext:  utils.NewIntArray(fillHoles(compressed.Actions.Next)),
		ActionCheck: utils.NewTypedIntArray(columnType, fillChecks(compressed.Actions.Check, noColumn)),

		NoColumn:         noColumn,
		NoTerminalColumn: table.NoTerminalColumn,

		DefaultActionByState: utils.NewIntArray(newDefaultActions(compressed)),

		GotoBase: utils.NewIntArray(compressed.Gotos.Base),
		GotoNext: utils.NewIntArray(fillHoles(compressed.Gotos.Next)),
		GotoCheck: utils.NewTypedIntArray(
			utils.GoUintType(nonterminalCount),
			fillChecks(compressed.Gotos.Check, nonterminalCount),
		),

		NoNonterminal: nonterminalCount,

		PopCountByProduction:    utils.NewIntArray(newPopCounts(parser)),
		NonterminalByProduction: utils.NewIntArray(newNonterminals(parser)),

		ErrorTerminalColumn: errorTerminalColumn,
		HasErrorRecovery:    compressed.HasErrorRecovery(),

		ActionKindBits:   table.ActionKindBits,
		ActionKindMask:   table.ActionKindMask,
		ActionKindShift:  int(table.ActionKindShift),
		ActionKindReduce: int(table.ActionKindReduce),
		ActionKindAccept: int(table.ActionKindAccept),
		ActionKindError:  int(table.ActionKindError),
	}
}

// newTerminalColumnByToken returns one entry per terminal of the grammar, naming the token constant which stands for it
// and the column of the action table which holds its decisions.
func newTerminalColumnByToken(parser backend.Parser) []TokenColumn {
	result := make([]TokenColumn, 0, len(parser.Grammar.Terminals))
	for terminalIdx, terminal := range parser.Grammar.Terminals {
		result = append(result, TokenColumn{
			Name:   terminalName(terminal),
			Column: table.TerminalColumn(terminalIdx),
		})
	}
	return result
}

// newDefaultActions returns the action every state takes for a terminal its row has no entry for.
//
// A state without a default action carries the error action instead of the absent entry the table package reports. Both
// make such a terminal a syntax error, but the error action is a value like any other, which keeps the emitted table
// free of a negative entry and therefore lets it use an unsigned type.
func newDefaultActions(compressed table.CompressedParser) []int {
	result := make([]int, len(compressed.DefaultActionByStateIdx))
	for stateIdx, action := range compressed.DefaultActionByStateIdx {
		if action == table.NoAction {
			action = table.NewErrorAction()
		}
		result[stateIdx] = int(action)
	}
	return result
}

// newPopCounts returns, for every production, the number of symbols a reduction by it takes off the stacks.
func newPopCounts(parser backend.Parser) []int {
	result := make([]int, len(parser.Grammar.Productions))
	for productionIdx, production := range parser.Grammar.Productions {
		result[productionIdx] = len(production.SymbolRefs)
	}
	return result
}

// newNonterminals returns, for every production, the nonterminal on its left hand side.
func newNonterminals(parser backend.Parser) []int {
	result := make([]int, len(parser.Grammar.Productions))
	for productionIdx, production := range parser.Grammar.Productions {
		result[productionIdx] = production.NonterminalIdx
	}
	return result
}

// fillHoles replaces the holes of a packed table with zeros. A hole is never read, because the check table keeps the
// driver from using it, so the value only has to keep the entry type free of a negative value.
func fillHoles(values []int) []int {
	result := make([]int, len(values))
	for cellIdx, value := range values {
		result[cellIdx] = max(value, 0)
	}
	return result
}

// fillChecks replaces the holes of a check table with the given value, which is one past the highest column in use and
// therefore matches no lookup.
func fillChecks(values []int, noColumn int) []int {
	result := make([]int, len(values))
	for cellIdx, value := range values {
		result[cellIdx] = value
		if value == utils.NoColumn {
			result[cellIdx] = noColumn
		}
	}
	return result
}
