package table

import (
	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

// Tables holds the lookup tables of a table driven parser in the form a template writes them out. Every table driven
// backend needs exactly this data; only the integer type a table is given, and for the JVM backends the split into
// chunks, are language specific and stay in the backend.
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
	// occupies. It has the entry type of the terminal column lookup, so that the driver can compare the two without a
	// conversion.
	ActionCheck utils.IntArray

	// NoColumn is the entry ActionCheck uses for a cell which no state occupies. It is one past the highest column in
	// use, so no lookup can ever ask for it.
	NoColumn int

	// NoTerminalColumn is the column a token gets which is no terminal of this grammar. It holds no entry in any
	// state, so a lookup falls through to the default action of the state.
	NoTerminalColumn int

	// DefaultActionByState holds, for every state, the action it takes for a terminal ActionNext has no entry for. A
	// state which has no default action carries the error action here, which makes such a terminal a syntax error.
	DefaultActionByState utils.IntArray

	// GotoBase maps a state to the displacement of its row within GotoNext.
	GotoBase utils.IntArray

	// GotoNext holds the gotos of all states packed into a single table. The goto a state has for a nonterminal lives
	// at GotoBase[state] + nonterminal, if GotoCheck confirms that the cell belongs to that nonterminal.
	GotoNext utils.IntArray

	// GotoCheck holds the nonterminal every cell of GotoNext belongs to, or NoNonterminal for a cell which no state
	// occupies.
	GotoCheck utils.IntArray

	// NoNonterminal is the entry GotoCheck uses for a cell which no state occupies. It is one past the highest
	// nonterminal index, so no lookup can ever ask for it.
	NoNonterminal int

	// DefaultGotoByNonterminal holds, for every nonterminal, the state a goto on it leads to when GotoNext has no
	// entry for it. A nonterminal which no state has a goto on carries a state index which is never read.
	DefaultGotoByNonterminal utils.IntArray

	// PopCountByProduction holds, for every production, the number of symbols a reduction by it takes off the stacks.
	// This is the length of the right hand side of the production and needs no compression, so it is read off the
	// grammar rather than built by the table package.
	PopCountByProduction utils.IntArray

	// NonterminalByProduction holds, for every production, the nonterminal on its left hand side. A reduction looks up
	// its goto with it and labels the node it pushes with it.
	NonterminalByProduction utils.IntArray

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
}

// TokenColumn is one entry of the lookup which translates a token the scanner delivers into the column of the action
// table which holds the decisions for it.
//
// It carries the name of the token constant rather than its value, because the value belongs to the scanner and a
// parser backend never learns it. The generated lookup is written in terms of the token constants themselves, which is
// what keeps the two generators free of a numbering they would both have to agree on, see NoTerminalColumn.
type TokenColumn struct {
	// Name is the name of the token constant which stands for the terminal.
	Name string

	// Column is the column of the action table which holds the decisions for the token.
	Column int
}

// TablesOptions carries the language specific choices NewTables makes while it builds the shared tables.
type TablesOptions struct {
	// NewIntArray wraps a slice of table entries into the array form the backend emits, typed the way that language
	// narrows an integer. It is one of the utils.NewXxxIntArray constructors.
	NewIntArray func(values []int) utils.IntArray

	// UintType returns the narrowest type name which holds every value from zero up to the given one. It types the
	// two check tables, which are compared against a column at runtime and therefore need a shared type of their own
	// rather than the one their own values would pick. Leave it nil for a language which does not narrow, in which
	// case the check tables are typed like every other table.
	UintType func(maxValue int) string

	// TerminalName returns the name of the token constant which stands for the given terminal.
	TerminalName func(symbol frontend.Symbol) string
}

// NewTables compresses the given parser into the lookup tables the generated parser reads at runtime, in the form the
// given options spell them for the target language.
func NewTables(parser backend.Parser, opts TablesOptions) Tables {
	compressed := NewCompressedParser(parser)

	// One past the highest column in use, which is what a cell no state occupies carries.
	noColumn := TerminalColumn(len(parser.Grammar.Terminals)-1) + 1

	nonterminalCount := len(parser.Grammar.Nonterminals)

	errorTerminalColumn := NoTerminalColumn
	if compressed.ErrorTerminalIdx != NoTerminal {
		errorTerminalColumn = TerminalColumn(compressed.ErrorTerminalIdx)
	}

	return Tables{
		TerminalColumnByToken: terminalColumnByToken(parser, opts.TerminalName),

		ActionBase:  opts.NewIntArray(compressed.Actions.Base),
		ActionNext:  opts.NewIntArray(FillHoles(compressed.Actions.Next)),
		ActionCheck: opts.checkArray(FillChecks(compressed.Actions.Check, noColumn), noColumn),

		NoColumn:         noColumn,
		NoTerminalColumn: NoTerminalColumn,

		DefaultActionByState: opts.NewIntArray(DefaultActions(compressed)),

		GotoBase:  opts.NewIntArray(compressed.Gotos.Base),
		GotoNext:  opts.NewIntArray(FillHoles(compressed.Gotos.Next)),
		GotoCheck: opts.checkArray(FillChecks(compressed.Gotos.Check, nonterminalCount), nonterminalCount),

		NoNonterminal:            nonterminalCount,
		DefaultGotoByNonterminal: opts.NewIntArray(FillHoles(compressed.DefaultGotoByNonterminalIdx)),

		PopCountByProduction:    opts.NewIntArray(PopCounts(parser)),
		NonterminalByProduction: opts.NewIntArray(Nonterminals(parser)),
		ProductionNames:         backend.ProductionNames(parser.Grammar),

		ErrorTerminalColumn: errorTerminalColumn,

		ActionKindBits:   ActionKindBits,
		ActionKindMask:   ActionKindMask,
		ActionKindShift:  int(ActionKindShift),
		ActionKindReduce: int(ActionKindReduce),
		ActionKindAccept: int(ActionKindAccept),
		ActionKindError:  int(ActionKindError),
	}
}

// checkArray types a check table by the value one past the highest entry it can hold, or like any other table when the
// language does not narrow an integer.
func (o TablesOptions) checkArray(values []int, bound int) utils.IntArray {
	if o.UintType == nil {
		return o.NewIntArray(values)
	}
	return utils.NewTypedIntArray(o.UintType(bound), values)
}

// terminalColumnByToken returns one entry per terminal of the grammar, naming the token constant which stands for it
// and the column of the action table which holds its decisions.
func terminalColumnByToken(parser backend.Parser, terminalName func(symbol frontend.Symbol) string) []TokenColumn {
	result := make([]TokenColumn, 0, len(parser.Grammar.Terminals))
	for terminalIdx, terminal := range parser.Grammar.Terminals {
		result = append(result, TokenColumn{
			Name:   terminalName(terminal),
			Column: TerminalColumn(terminalIdx),
		})
	}
	return result
}
