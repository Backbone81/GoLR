package golr

import (
	intconflict "github.com/backbone81/golr/internal/parsergen/conflict"
	intlr1golr "github.com/backbone81/golr/internal/parsergen/core/lr1/golr"
	"github.com/backbone81/golr/pkg/parsergen/backend"
	"github.com/backbone81/golr/pkg/parsergen/conflict"
	"github.com/backbone81/golr/pkg/parsergen/frontend"
)

// GrammarToParser calculates a parser from the context free grammar.
func GrammarToParser(grammar frontend.Grammar) (backend.Parser, []conflict.Conflict, error) {
	return intlr1golr.GrammarToParser(grammar, intconflict.DefaultPolicy)
}
