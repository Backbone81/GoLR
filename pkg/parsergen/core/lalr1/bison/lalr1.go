package bison

import (
	intconflict "github.com/backbone81/golr/internal/parsergen/conflict"
	intlalr1bison "github.com/backbone81/golr/internal/parsergen/core/lalr1/bison"
	"github.com/backbone81/golr/pkg/parsergen/backend"
	"github.com/backbone81/golr/pkg/parsergen/conflict"
	"github.com/backbone81/golr/pkg/parsergen/frontend"
)

// GrammarToParser calculates a parser from the context free grammar.
//
// Conflicts are resolved by GNU Bison itself, which this core shells out to, so the parser tables come back with the
// conflicts already decided the way GNU Bison and Yacc decide them.
func GrammarToParser(grammar frontend.Grammar) (backend.Parser, []conflict.Conflict, error) {
	return intlalr1bison.GrammarToParser(grammar, intconflict.DefaultPolicy)
}
