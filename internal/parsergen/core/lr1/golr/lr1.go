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
// The grammar is taken as a frontend produces it. What comes back are canonical LR(1) parser tables a backend can
// serialize, with every conflict of the grammar decided, so that no state is left with more than one action for a
// terminal. Canonical LR(1) is the largest of the cores by a wide margin, which is what makes it a dependable oracle
// and a poor default.
//
// Every conflict which was found is returned, whether it was decided or not, because a parser generator reports the
// conflicts of a grammar to the user even when it decided them on its own. The error reports the conflicts which were
// left undecided, one conflict.UnresolvedConflictError each; no parser can be generated from such a grammar, so the
// parser tables come back empty then and the conflicts are all there is left to report. The construction also gives up
// with ErrStateLimitExceeded on a grammar which needs more states than a parser table can address.
func GrammarToParser(grammar frontend.Grammar) (backend.Parser, []conflict.Conflict, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Cores: LR1: GrammarToParser").End()

	// The builder works on the augmented grammar, so the caller hands us the grammar as the frontend produced it and
	// we augment it here, the same way the LALR(1) and IELR(1) cores do.
	augmentedGrammar := frontend.AugmentGrammar(grammar)

	conflictPolicy := conflict.NewDefaultPolicy(augmentedGrammar)

	// The builder returns the raw table with its conflicts intact, which is what the oracle and differential testing
	// work needs, so callers after those tables use NewLR1Builder directly instead of this function.
	builder := NewLR1Builder(augmentedGrammar)
	if err := builder.Build(); err != nil {
		return backend.Parser{}, nil, err
	}
	parser := builder.Parser()

	// Phase 5 of IELR(1) (section 3.7 of the paper), which the LALR(1) and IELR(1) cores run in the same place. A
	// grammar which is not LR(1) has conflicts here too - canonical LR(1) removes the mysterious LALR(1) conflicts of
	// section 2.5, not the genuine ones - so this is what keeps a conflict-laden table from reaching a backend.
	conflicts, err := conflict.Resolve(&parser, conflictPolicy)
	if err != nil {
		return backend.Parser{}, conflicts, err
	}

	// Resolving a conflict can delete the only shift into a state, which strands that state and everything behind it.
	// This is the unreachable state removal of section 3.8.2, the optional phase 6 of the paper.
	parser, conflicts = conflict.RemoveUnreachableStates(parser, conflicts)
	return parser, conflicts, nil
}
