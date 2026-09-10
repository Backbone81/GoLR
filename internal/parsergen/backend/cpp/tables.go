package cpp

import (
	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/backend/table"
	"github.com/backbone81/golr/internal/utils"
)

// Tables holds the lookup tables of a table driven parser in the form the template writes them out. It is the same data
// the Go table driven backend emits, differing only in the type a table is given, which is a C++ one here, plus the
// underlying types of the two enumerations the generated parser declares.
type Tables struct {
	table.Tables

	// NonterminalType is the underlying type of the Nonterminal enumeration. A scoped enumeration whose underlying
	// type is left open is as wide as an int, while a nonterminal is one alternative of a ParseSymbol and is held
	// once per node of the parse tree, so the type is what it costs there.
	NonterminalType string

	// ProductionType is the underlying type of the Production enumeration, sized to the highest production index.
	ProductionType string
}

// TokenColumn is one entry of the lookup which translates a token into the column of the action table holding the
// decisions for it.
type TokenColumn = table.TokenColumn

// NewTables compresses the given parser into the lookup tables the generated parser reads at runtime.
func NewTables(parser backend.Parser) Tables {
	return Tables{
		Tables: table.NewTables(parser, table.TablesOptions{
			NewIntArray:  utils.NewCppIntArray,
			UintType:     utils.CppUintType,
			TerminalName: terminalName,
		}),

		// The enumerators are one per nonterminal numbered from zero, so the last of them is the value the type has
		// to hold.
		NonterminalType: utils.CppUintType(len(parser.Grammar.Nonterminals) - 1),

		// The enumerators are one per production numbered from one, production 0 ($accept) having none, so the number
		// of productions other than $accept is the value the type has to hold.
		ProductionType: utils.CppUintType(len(parser.Grammar.Productions) - 1),
	}
}
