package frontend

import (
	"fmt"
)

// Symbol is the textual representation of either a terminal or a nonterminal.
type Symbol struct {
	// Name is the technical name for that symbol.
	Name string `json:"name" yaml:"name"`

	// Alias is an alternative name for that symbol which might be less technical. For example while the technical name
	// for a terminal might be OP_PLUS the alias might be "+" to make it easier to read.
	Alias string `json:"alias,omitempty" yaml:"alias,omitempty"`

	// Associativity describes how the terminal is associating with other terminals.
	Associativity Associativity `json:"associativity,omitempty" yaml:"associativity,omitempty"`

	// Precedence describes the precedence of the terminal in relation to other terminals.
	Precedence int `json:"precedence,omitempty" yaml:"precedence,omitempty"`
}

// Symbol implements fmt.Stringer.
var _ fmt.Stringer = (*Symbol)(nil)

// String returns the alias for the symbol if an alias is set. If no alias is set the name is returned.
func (s Symbol) String() string {
	if s.Alias == "" {
		return s.Name
	}
	return s.Alias
}

// SymbolEOF is the end of file symbol which marks the end of the parse.
var SymbolEOF = Symbol{
	Name: "$end",
}

// SymbolError is the error symbol which marks the places in the grammar where a parser can resume after a syntax
// error. It is the special terminal symbol of the error recovery productions of section 9 "Error Recovery" of "LR
// Parsing" by Aho and Johnson: a terminal for the purpose of constructing the parse tables, so that states carrying an
// error recovery production get a shift action on it, but one which no scanner ever produces. The parser shifts it
// itself while recovering.
//
// The name is what identifies the symbol, at whatever terminal index a grammar happens to carry it. Nothing relies on
// it sitting at a fixed position: the few places which need it resolve it once per grammar while the tables are built,
// so a fixed index would buy nothing and would force every one of them to know whether it is looking at a grammar
// before or after AugmentGrammar.
//
// The name is spelled with the same leading dollar sign as SymbolEOF, which no terminal a frontend can read is allowed
// to carry. That keeps a terminal a grammar author declares from silently taking on the meaning of the error symbol:
// GNU Bison reserves the plain name `error` for this and so cannot be given a token of that name, while a GoLR scanner
// section may well declare one.
var SymbolError = Symbol{
	Name: "$error",
}
