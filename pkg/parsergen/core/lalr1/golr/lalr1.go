package golr

import (
	intconflict "github.com/backbone81/golr/internal/parsergen/conflict"
	intlalr1golr "github.com/backbone81/golr/internal/parsergen/core/lalr1/golr"
	"github.com/backbone81/golr/pkg/parsergen/backend"
	"github.com/backbone81/golr/pkg/parsergen/conflict"
	"github.com/backbone81/golr/pkg/parsergen/frontend"
)

// GrammarToParser calculates a parser from the context free grammar.
//
// Conflicts are resolved the way GNU Bison and Yacc do: precedence and associativity decide first, a shift beats a
// reduction when precedence has nothing to say, and the production which was declared first wins a conflict between two
// reductions. Which rules apply is composable inside the library, but that is not part of the public API yet, which is
// why this forwards with the default policy instead of being an alias of the internal function.
func GrammarToParser(grammar frontend.Grammar) (backend.Parser, []conflict.Conflict, error) {
	return intlalr1golr.GrammarToParser(grammar, intconflict.DefaultPolicy)
}
