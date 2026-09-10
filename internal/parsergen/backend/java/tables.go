package java

import (
	"slices"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/backend/table"
	"github.com/backbone81/golr/internal/utils"
)

// Tables holds the lookup tables of a table driven parser in the form the template writes them out. It is the shared
// table data with every numeric table split into chunks a single method can hold, because the Java virtual machine
// limits a method to 64 KB of bytecode.
type Tables struct {
	table.Tables

	// TerminalColumnType is the type of an entry of the table TerminalColumnByToken is built into. It is the entry
	// type of the check tables, so the driver can compare the two without a conversion.
	TerminalColumnType string

	// ActionBase maps a state to the displacement of its row within ActionNext.
	ActionBase utils.JavaTable

	// ActionNext holds the actions of all states packed into a single table.
	ActionNext utils.JavaTable

	// ActionCheck holds the column every cell of ActionNext belongs to, or NoColumn for a cell which no state
	// occupies.
	ActionCheck utils.JavaTable

	// DefaultActionByState holds, for every state, the action it takes for a terminal ActionNext has no entry for.
	DefaultActionByState utils.JavaTable

	// GotoBase maps a state to the displacement of its row within GotoNext.
	GotoBase utils.JavaTable

	// GotoNext holds the gotos of all states packed into a single table.
	GotoNext utils.JavaTable

	// GotoCheck holds the nonterminal every cell of GotoNext belongs to, or NoNonterminal for a cell which no state
	// occupies.
	GotoCheck utils.JavaTable

	// DefaultGotoByNonterminal holds, for every nonterminal, the state a goto on it leads to when GotoNext has no
	// entry for it.
	DefaultGotoByNonterminal utils.JavaTable

	// PopCountByProduction holds, for every production, the number of symbols a reduction by it takes off the stacks.
	PopCountByProduction utils.JavaTable

	// NonterminalByProduction holds, for every production, the nonterminal on its left hand side.
	NonterminalByProduction utils.JavaTable

	// ConcatTypes are the types the generated parser joins the chunks of a table for, one entry per type in use.
	ConcatTypes []string
}

// NewTables compresses the given parser into the lookup tables the generated parser reads at runtime.
func NewTables(parser backend.Parser) Tables {
	shared := table.NewTables(parser, table.TablesOptions{
		NewIntArray:  utils.NewJavaIntArray,
		UintType:     utils.JavaIntType,
		TerminalName: terminalName,
	})

	tables := Tables{
		Tables:             shared,
		TerminalColumnType: shared.ActionCheck.Type,

		ActionBase:               utils.NewJavaTable("actionBase", shared.ActionBase),
		ActionNext:               utils.NewJavaTable("actionNext", shared.ActionNext),
		ActionCheck:              utils.NewJavaTable("actionCheck", shared.ActionCheck),
		DefaultActionByState:     utils.NewJavaTable("defaultActionByState", shared.DefaultActionByState),
		GotoBase:                 utils.NewJavaTable("gotoBase", shared.GotoBase),
		GotoNext:                 utils.NewJavaTable("gotoNext", shared.GotoNext),
		GotoCheck:                utils.NewJavaTable("gotoCheck", shared.GotoCheck),
		DefaultGotoByNonterminal: utils.NewJavaTable("defaultGotoByNonterminal", shared.DefaultGotoByNonterminal),
		PopCountByProduction:     utils.NewJavaTable("popCountByProduction", shared.PopCountByProduction),
		NonterminalByProduction:  utils.NewJavaTable("nonterminalByProduction", shared.NonterminalByProduction),
	}
	tables.ConcatTypes = concatTypes(tables)
	return tables
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
