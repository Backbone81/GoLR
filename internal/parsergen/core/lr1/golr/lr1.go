package golr

import (
	"context"
	"runtime/trace"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// GrammarToParser calculates a parser from the context free grammar.
func GrammarToParser(augmentedGrammar frontend.Grammar) (backend.Parser, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Cores: LR1: GrammarToParser").End()

	builder := NewLR1Builder(augmentedGrammar)
	if err := builder.Build(); err != nil {
		return backend.Parser{}, err
	}
	return builder.Parser(), nil
}
