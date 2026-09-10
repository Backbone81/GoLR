package python

import (
	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/backend/table"
	"github.com/backbone81/golr/internal/utils"
)

// Tables holds the lookup tables of a table driven parser in the form the template writes them out. It is the same data
// the Go table driven backend emits, with the entry types left off, because Python has a single integer type and
// nothing to narrow.
type Tables = table.Tables

// TokenColumn is one entry of the lookup which translates a token into the column of the action table holding the
// decisions for it.
type TokenColumn = table.TokenColumn

// NewTables compresses the given parser into the lookup tables the generated parser reads at runtime. Python has one
// integer type and no fixed width, so the tables carry no type name and the check tables are not narrowed.
func NewTables(parser backend.Parser) Tables {
	return table.NewTables(parser, table.TablesOptions{
		NewIntArray:  utils.NewPythonIntArray,
		TerminalName: terminalName,
	})
}
