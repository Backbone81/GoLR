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
//
// The grammar is taken as a frontend produces it. What comes back are LALR(1) parser tables a backend can serialize,
// with every conflict of the grammar decided, so that no state is left with more than one action for a terminal. LALR(1)
// gives the smallest tables of all the cores, at the price of the mysterious conflicts of section 2.5 of IELR(1):
// conflicts which merging states created and which the grammar does not really have.
//
// Every conflict which was found is returned, whether it was decided or not, because a parser generator reports the
// conflicts of a grammar to the user even when it decided them on its own. The error reports the conflicts which were
// left undecided, one conflict.UnresolvedConflictError each; no parser can be generated from such a grammar, so the
// parser tables come back empty then and the conflicts are all there is left to report.
func GrammarToParser(grammar frontend.Grammar) (backend.Parser, []conflict.Conflict, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Cores: LALR1: GrammarToParser").End()

	// The builder works on the augmented grammar, so the caller hands us the grammar as the frontend produced it and
	// we augment it here.
	augmentedGrammar := frontend.AugmentGrammar(grammar)

	conflictPolicy := conflict.NewDefaultPolicy(augmentedGrammar)

	// Phase 0 of IELR(1) is LALR(1) in full, so this core is that phase on its own, without the phases which split the
	// states it merged too eagerly.
	builder := ielr1golrcore.NewLALR1Builder(augmentedGrammar)
	builder.Build()
	parser := builder.Parser()

	// Phase 5 of IELR(1) (section 3.7 of the paper).
	conflicts, err := conflict.Resolve(&parser, conflictPolicy)
	if err != nil {
		return backend.Parser{}, conflicts, err
	}

	// Resolving a conflict can delete the only shift into a state, which strands that state and everything behind it.
	// This is the unreachable state removal of section 3.8.2, the optional phase 6 of the paper.
	parser, conflicts = conflict.RemoveUnreachableStates(parser, conflicts)
	return parser, conflicts, nil
}
