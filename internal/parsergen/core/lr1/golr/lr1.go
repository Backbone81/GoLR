package golr

import (
	"context"
	"runtime/trace"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// GrammarToParser calculates a parser from the context free grammar.
func GrammarToParser(grammar frontend.Grammar) (backend.Parser, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Cores: LR1: GrammarToParser").End()

	// The builder works on the augmented grammar, so the caller hands us the grammar as the frontend produced it and
	// we augment it here, the same way the LALR(1) and IELR(1) cores do.
	builder := NewLR1Builder(frontend.AugmentGrammar(grammar))
	if err := builder.Build(); err != nil {
		return backend.Parser{}, err
	}
	return builder.Parser(), nil
}
