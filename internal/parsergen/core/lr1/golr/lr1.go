package golr

import (
	"context"
	"runtime/trace"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// GrammarToParser calculates a parser from the context free grammar.
//
// Like the LALR(1) and IELR(1) cores, it resolves the conflicts of the table before returning it: the builder produces
// the raw canonical LR(1) table with its conflicts intact, and conflict.Resolve applies the resolution policy as a
// final pass, removing the losing actions and reporting any conflict the policy left undecided as an error. A grammar
// which is not LR(1) has conflicts here too — canonical LR(1) removes the mysterious LALR(1) conflicts, not the genuine
// ones — so resolving is what keeps a malformed, conflict-laden table from being handed to a backend. Callers which
// need the raw table for oracle work use NewLR1Builder directly instead of this interface.
func GrammarToParser(grammar frontend.Grammar) (backend.Parser, []conflict.Conflict, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Cores: LR1: GrammarToParser").End()

	// The builder works on the augmented grammar, so the caller hands us the grammar as the frontend produced it and
	// we augment it here, the same way the LALR(1) and IELR(1) cores do.
	augmentedGrammar := frontend.AugmentGrammar(grammar)

	conflictPolicy := conflict.NewDefaultPolicy(augmentedGrammar)
	builder := NewLR1Builder(augmentedGrammar)
	if err := builder.Build(); err != nil {
		return backend.Parser{}, nil, err
	}
	parser := builder.Parser()

	conflicts, err := conflict.Resolve(&parser, conflictPolicy)
	if err != nil {
		return backend.Parser{}, conflicts, err
	}
	return parser, conflicts, nil
}
