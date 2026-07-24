package golr

import (
	"context"
	"runtime/trace"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	"github.com/backbone81/golr/internal/parsergen/core"
	ielr1golrcore "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// GrammarToParser calculates a parser from the context free grammar.
//
// The grammar is taken as a frontend produces it. What comes back are LALR(1) parser tables a backend can serialize,
// with every conflict of the grammar decided, so that no state is left with more than one action for a terminal.
// LALR(1) gives the smallest tables of all the cores, at the price of the mysterious conflicts of section 2.5 of
// IELR(1): conflicts which merging states created and which the grammar does not really have.
//
// The policy factory decides the conflicts, see conflict.PolicyFactory. Pass conflict.DefaultPolicy to decide them
// the way GNU Bison and Yacc do.
//
// Every conflict which was found is returned, whether it was decided or not, because a parser generator reports the
// conflicts of a grammar to the user even when it decided them on its own. The error reports the conflicts which were
// left undecided, one conflict.UnresolvedConflictError each; no parser can be generated from such a grammar, so the
// parser tables come back empty then and the conflicts are all there is left to report. The construction also gives up
// with backend.ErrStateLimitExceeded on a grammar which needs more states than a parser table can address.
func GrammarToParser(
	grammar frontend.Grammar,
	policyFactory conflict.PolicyFactory,
	options ...core.Option,
) (backend.Parser, []conflict.Conflict, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Core: LALR1: GoLR: GrammarToParser").End()

	config := core.ConfigFromOptions(options...)

	parser, err := GrammarToUnresolvedParser(grammar, policyFactory)
	if err != nil {
		return backend.Parser{}, nil, err
	}

	// Phase 5 of IELR(1) (section 3.7 of the paper).
	conflicts, err := conflict.Resolve(&parser, policyFactory(parser.Grammar))
	if err != nil {
		return backend.Parser{}, conflicts, err
	}

	if config.DefaultReductions {
		backend.ApplyDefaultReductions(&parser)
	}

	// Resolving a conflict can delete the only shift into a state, which strands that state and everything behind it.
	// This is the unreachable state removal of section 3.8.2, the optional phase 6 of the paper.
	parser, conflicts = conflict.RemoveUnreachableStates(parser, conflicts)
	return parser, conflicts, nil
}

// GrammarToUnresolvedParser calculates the LALR(1) parser tables of the grammar and stops there, so the tables come
// back with their conflicts intact and their unreachable states in place, before phase 5 decides anything.
//
// This is what the oracle and differential testing work is after, because the mysterious conflicts of section 2.5 of
// IELR(1) are the whole point of comparing LALR(1) against canonical LR(1), and a table whose conflicts are already
// resolved has none left to compare. It saves those callers from augmenting the grammar and driving the builder
// themselves, and it gives the three GoLR cores one shape to be called with. Reach for GrammarToParser whenever the
// tables are meant for a backend, because a table with conflicts in it is not a parser.
//
// LALR(1) needs no policy to be constructed: it merges every state with the same core, no matter how the conflicts that
// causes are decided afterwards. The factory is taken all the same, so that the three GoLR cores agree on their
// signature and a caller can switch between them, and because IELR(1) does need the policy while it builds.
//
// It gives up with backend.ErrStateLimitExceeded on a grammar which needs more states than a parser table can address.
func GrammarToUnresolvedParser(
	grammar frontend.Grammar,
	policyFactory conflict.PolicyFactory,
) (backend.Parser, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Core: LALR1: GoLR: GrammarToUnresolvedParser").End()

	// The builder works on the augmented grammar, so the caller hands us the grammar as the frontend produced it and
	// we augment it here.
	augmentedGrammar := frontend.AugmentGrammar(grammar)

	// Phase 0 of IELR(1) is LALR(1) in full, so this core is that phase on its own, without the phases which split the
	// states it merged too eagerly.
	builder := ielr1golrcore.NewLALR1Builder(augmentedGrammar)
	if err := builder.Build(); err != nil {
		return backend.Parser{}, err
	}
	return builder.Parser(), nil
}
