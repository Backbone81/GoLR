package golr

import (
	"context"
	"runtime/trace"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	ielr1golrcore "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// GrammarToParser calculates a parser from the context free grammar.
func GrammarToParser(grammar frontend.Grammar) (backend.Parser, []conflict.Conflict, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Cores: LALR1: GrammarToParser").End()

	augmentedGrammar := frontend.AugmentGrammar(grammar)

	conflictPolicy := conflict.NewDefaultPolicy(augmentedGrammar)
	builder := ielr1golrcore.NewLALR1Builder(augmentedGrammar)
	builder.Build()
	parser := builder.Parser()

	conflicts, err := conflict.Resolve(&parser, conflictPolicy)
	if err != nil {
		return backend.Parser{}, conflicts, err
	}
	return parser, conflicts, nil
}
