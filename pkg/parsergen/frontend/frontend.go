package frontend

import intfrontend "github.com/backbone81/golr/internal/parsergen/frontend"

type (
	// Grammar is a context free grammar.
	Grammar = intfrontend.Grammar

	// Symbol is the textual representation of either a terminal or a nonterminal.
	Symbol = intfrontend.Symbol

	// SymbolRef is either a terminal index or a nonterminal index. The most significant bit is used to signal a
	// nonterminal. The maximum terminal or nonterminal index which can be stored is 32767.
	SymbolRef = intfrontend.SymbolRef

	// Production is a production of a context-free grammar. The Nonterminal is the left hand side of the production and
	// the Symbols are the right hand side of the production.
	Production = intfrontend.Production

	// Associativity describes how terminals associate with each other.
	Associativity = intfrontend.Associativity
)

const (
	// AssociativityUndeclared means that no associativity is declared. This is the default for every terminal.
	AssociativityUndeclared = intfrontend.AssociativityUndeclared

	// AssociativityLeft introduces a left associativity.
	AssociativityLeft = intfrontend.AssociativityLeft

	// AssociativityRight introduces a right associativity.
	AssociativityRight = intfrontend.AssociativityRight

	// AssociativityNone describes that the terminal should not associate at all and should trigger an error if some
	// association is needed.
	AssociativityNone = intfrontend.AssociativityNone
)

var (
	// NewTerminalRef creates a new SymbolRef for a terminal index.
	NewTerminalRef = intfrontend.NewTerminalRef

	// NewNonterminalRef creates a new SymbolRef for a nonterminal index.
	NewNonterminalRef = intfrontend.NewNonterminalRef
)
