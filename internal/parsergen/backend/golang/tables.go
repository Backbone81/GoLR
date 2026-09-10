package golang

import (
	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/backend/table"
	"github.com/backbone81/golr/internal/utils"
)

// Tables holds the lookup tables of a table driven parser in the form the template writes them out.
type Tables = table.Tables

// TokenColumn is one entry of the table which translates a token into the column of the action table holding the
// decisions for it.
type TokenColumn = table.TokenColumn

// NewTables compresses the given parser into the lookup tables the generated parser reads at runtime. The tables use
// Go integer types, picked per grammar so the entries of a small grammar fit into a single byte each.
func NewTables(parser backend.Parser) Tables {
	return table.NewTables(parser, table.TablesOptions{
		NewIntArray:  utils.NewIntArray,
		UintType:     utils.GoUintType,
		TerminalName: terminalName,
	})
}
