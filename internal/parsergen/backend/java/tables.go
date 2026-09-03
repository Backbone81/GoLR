package java

import (
	"slices"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/backend/table"
	"github.com/backbone81/golr/internal/utils"
)

// Tables holds the lookup tables of a table driven parser in the form the template writes them out. It is the same
// data the Go table driven backend emits, differing only in the type a table is given, which is a Java one here, and
// in every table being split into chunks a single method can hold.
type Tables struct {
	// TerminalColumnByToken translates a token the scanner delivers into the column of the action table which holds
	// the decisions for it. A token which is not a terminal of this grammar is not in here and gets
	// NoTerminalColumn, see TokenColumn.
	TerminalColumnByToken []TokenColumn

	// TerminalColumnType is the type of an entry of the table TerminalColumnByToken is built into.
	TerminalColumnType string

	// ActionBase maps a state to the displacement of its row within ActionNext.
	ActionBase utils.JavaTable

	// ActionNext holds the actions of all states packed into a single table. The action a state has for a terminal
	// lives at ActionBase[state] + column, if ActionCheck confirms that the cell belongs to that column.
	ActionNext utils.JavaTable

	// ActionCheck holds the column every cell of ActionNext belongs to, or NoColumn for a cell which no state
	// occupies. It has the same entry type as TerminalColumnByToken, so that the driver can compare the two without
	// a conversion.
	ActionCheck utils.JavaTable

	// NoColumn is the entry ActionCheck uses for a cell which no state occupies. It is one past the highest column
	// in use, so no lookup can ever ask for it.
	NoColumn int

	// NoTerminalColumn is the column a token gets which is no terminal of this grammar. It holds no entry in any
	// state, so a lookup falls through to the default action of the state.
	NoTerminalColumn int

	// DefaultActionByState holds, for every state, the action it takes for a terminal ActionNext has no entry for.
	// A state which has no default action carries the error action here, which makes such a terminal a syntax
	// error.
	DefaultActionByState utils.JavaTable

	// GotoBase maps a state to the displacement of its row within GotoNext.
	GotoBase utils.JavaTable

	// GotoNext holds the gotos of all states packed into a single table. The goto a state has for a nonterminal
	// lives at GotoBase[state] + nonterminal, if GotoCheck confirms that the cell belongs to that nonterminal.
	GotoNext utils.JavaTable

	// GotoCheck holds the nonterminal every cell of GotoNext belongs to, or NoNonterminal for a cell which no state
	// occupies.
	GotoCheck utils.JavaTable

	// NoNonterminal is the entry GotoCheck uses for a cell which no state occupies. It is one past the highest
	// nonterminal index, so no lookup can ever ask for it.
	NoNonterminal int

	// DefaultGotoByNonterminal holds, for every nonterminal, the state a goto on it leads to when GotoNext has no
	// entry for it. A nonterminal which no state has a goto on carries a state index which is never read.
	DefaultGotoByNonterminal utils.JavaTable

	// PopCountByProduction holds, for every production, the number of symbols a reduction by it takes off the
	// stacks. This is the length of the right hand side of the production and needs no compression, so it is read
	// off the grammar rather than built by the table package.
	PopCountByProduction utils.JavaTable

	// NonterminalByProduction holds, for every production, the nonterminal on its left hand side. A reduction looks
	// up its goto with it and labels the node it pushes with it.
	NonterminalByProduction utils.JavaTable

	// ProductionNames holds, for every production, its name. Index 0 ($accept) is empty, since that production is
	// never reduced. See backend.ProductionNames.
	ProductionNames []string

	// ErrorTerminalColumn is the column of the action table which holds the shifts of the error symbol, which is
	// where the error recovery reads the state to resume in.
	ErrorTerminalColumn int

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

	// ConcatTypes are the types the generated parser joins the chunks of a table for, one entry per type in use.
	ConcatTypes []string
}

// NewTables compresses the given parser into the lookup tables the generated parser reads at runtime.
//
//nolint:funlen // It is one literal with a field per table, which reads better whole than split up.
func NewTables(parser backend.Parser) Tables {
	compressed := table.NewCompressedParser(parser)

	// The check entries hold a column or the one value past them which means the cell is unused, so the column table
	// and the check table are typed by that value and can be compared against each other directly.
	maxColumn := table.TerminalColumn(len(parser.Grammar.Terminals) - 1)
	noColumn := maxColumn + 1
	columnType := utils.JavaIntType(noColumn)

	nonterminalCount := len(parser.Grammar.Nonterminals)

	errorTerminalColumn := table.NoTerminalColumn
	if compressed.ErrorTerminalIdx != table.NoTerminal {
		errorTerminalColumn = table.TerminalColumn(compressed.ErrorTerminalIdx)
	}

	tables := Tables{
		TerminalColumnByToken: newTerminalColumnByToken(parser),
		TerminalColumnType:    columnType,

		ActionBase: utils.NewJavaTable("actionBase", utils.NewJavaIntArray(compressed.Actions.Base)),
		ActionNext: utils.NewJavaTable("actionNext", utils.NewJavaIntArray(table.FillHoles(compressed.Actions.Next))),
		ActionCheck: utils.NewJavaTable(
			"actionCheck",
			utils.NewTypedIntArray(columnType, table.FillChecks(compressed.Actions.Check, noColumn)),
		),

		NoColumn:         noColumn,
		NoTerminalColumn: table.NoTerminalColumn,

		DefaultActionByState: utils.NewJavaTable(
			"defaultActionByState",
			utils.NewJavaIntArray(table.DefaultActions(compressed)),
		),

		GotoBase: utils.NewJavaTable("gotoBase", utils.NewJavaIntArray(compressed.Gotos.Base)),
		GotoNext: utils.NewJavaTable("gotoNext", utils.NewJavaIntArray(table.FillHoles(compressed.Gotos.Next))),
		GotoCheck: utils.NewJavaTable(
			"gotoCheck",
			utils.NewTypedIntArray(
				utils.JavaIntType(nonterminalCount),
				table.FillChecks(compressed.Gotos.Check, nonterminalCount),
			),
		),

		NoNonterminal: nonterminalCount,
		DefaultGotoByNonterminal: utils.NewJavaTable(
			"defaultGotoByNonterminal",
			utils.NewJavaIntArray(table.FillHoles(compressed.DefaultGotoByNonterminalIdx)),
		),

		PopCountByProduction: utils.NewJavaTable(
			"popCountByProduction",
			utils.NewJavaIntArray(table.PopCounts(parser)),
		),
		NonterminalByProduction: utils.NewJavaTable(
			"nonterminalByProduction",
			utils.NewJavaIntArray(table.Nonterminals(parser)),
		),
		ProductionNames: backend.ProductionNames(parser.Grammar),

		ErrorTerminalColumn: errorTerminalColumn,

		ActionKindBits:   table.ActionKindBits,
		ActionKindMask:   table.ActionKindMask,
		ActionKindShift:  int(table.ActionKindShift),
		ActionKindReduce: int(table.ActionKindReduce),
		ActionKindAccept: int(table.ActionKindAccept),
		ActionKindError:  int(table.ActionKindError),
	}
	tables.ConcatTypes = concatTypes(tables)
	return tables
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

// concatTypes returns the types the generated parser needs a way of joining the chunks of a table for, which is one
// entry per type in use rather than one per table.
func concatTypes(tables Tables) []string {
	var result []string
	for _, typeName := range []string{
		tables.ActionBase.Type,
		tables.ActionNext.Type,
		tables.ActionCheck.Type,
		tables.DefaultActionByState.Type,
		tables.GotoBase.Type,
		tables.GotoNext.Type,
		tables.GotoCheck.Type,
		tables.DefaultGotoByNonterminal.Type,
		tables.PopCountByProduction.Type,
		tables.NonterminalByProduction.Type,
	} {
		if !slices.Contains(result, typeName) {
			result = append(result, typeName)
		}
	}
	return result
}
