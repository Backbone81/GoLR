package kotlin

import (
	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/backend/table"
	"github.com/backbone81/golr/internal/utils"
)

// Tables holds the lookup tables of a table driven parser in the form the template writes them out. It is the shared
// table data with every numeric table split into chunks a single function can hold, because Kotlin compiles to the
// same virtual machine Java does and a method may hold only 64 KB of bytecode.
type Tables struct {
	table.Tables

	// TerminalColumn is the array the entries of TerminalColumnByToken are built into. It holds a column, so it has
	// the entry type of the check tables and the driver can compare the two without a conversion of its own.
	TerminalColumn utils.KotlinArray

	// ActionBase maps a state to the displacement of its row within ActionNext.
	ActionBase utils.KotlinTable

	// ActionNext holds the actions of all states packed into a single table.
	ActionNext utils.KotlinTable

	// ActionCheck holds the column every cell of ActionNext belongs to, or NoColumn for a cell which no state
	// occupies.
	ActionCheck utils.KotlinTable

	// DefaultActionByState holds, for every state, the action it takes for a terminal ActionNext has no entry for.
	DefaultActionByState utils.KotlinTable

	// GotoBase maps a state to the displacement of its row within GotoNext.
	GotoBase utils.KotlinTable

	// GotoNext holds the gotos of all states packed into a single table.
	GotoNext utils.KotlinTable

	// GotoCheck holds the nonterminal every cell of GotoNext belongs to, or NoNonterminal for a cell which no state
	// occupies.
	GotoCheck utils.KotlinTable

	// DefaultGotoByNonterminal holds, for every nonterminal, the state a goto on it leads to when GotoNext has no
	// entry for it.
	DefaultGotoByNonterminal utils.KotlinTable

	// PopCountByProduction holds, for every production, the number of symbols a reduction by it takes off the stacks.
	PopCountByProduction utils.KotlinTable

	// NonterminalByProduction holds, for every production, the nonterminal on its left hand side.
	NonterminalByProduction utils.KotlinTable
}

// NewTables compresses the given parser into the lookup tables the generated parser reads at runtime.
func NewTables(parser backend.Parser) Tables {
	shared := table.NewTables(parser, table.TablesOptions{
		NewIntArray:  utils.NewKotlinIntArray,
		UintType:     utils.KotlinIntType,
		TerminalName: terminalName,
	})

	return Tables{
		Tables:         shared,
		TerminalColumn: utils.NewKotlinArray(shared.ActionCheck.Type),

		ActionBase:               utils.NewKotlinTable("actionBase", shared.ActionBase),
		ActionNext:               utils.NewKotlinTable("actionNext", shared.ActionNext),
		ActionCheck:              utils.NewKotlinTable("actionCheck", shared.ActionCheck),
		DefaultActionByState:     utils.NewKotlinTable("defaultActionByState", shared.DefaultActionByState),
		GotoBase:                 utils.NewKotlinTable("gotoBase", shared.GotoBase),
		GotoNext:                 utils.NewKotlinTable("gotoNext", shared.GotoNext),
		GotoCheck:                utils.NewKotlinTable("gotoCheck", shared.GotoCheck),
		DefaultGotoByNonterminal: utils.NewKotlinTable("defaultGotoByNonterminal", shared.DefaultGotoByNonterminal),
		PopCountByProduction:     utils.NewKotlinTable("popCountByProduction", shared.PopCountByProduction),
		NonterminalByProduction:  utils.NewKotlinTable("nonterminalByProduction", shared.NonterminalByProduction),
	}
}
