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
// The policy factory decides the conflicts, see conflict.PolicyFactory. Pass conflict.DefaultPolicy to decide them
// the way GNU Bison and Yacc do.
//
// Every conflict which was found is returned, whether it was decided or not, because a parser generator reports the
// conflicts of a grammar to the user even when it decided them on its own. The error reports the conflicts which were
// left undecided, one conflict.UnresolvedConflictError each; no parser can be generated from such a grammar, so the
// parser tables come back empty then and the conflicts are all there is left to report. The construction also gives up
// with ErrStateLimitExceeded on a grammar which needs more states than a parser table can address.
func GrammarToParser(
	grammar frontend.Grammar,
	policyFactory conflict.PolicyFactory,
) (backend.Parser, []conflict.Conflict, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Cores: LR1: GrammarToParser").End()

	parser, err := GrammarToUnresolvedParser(grammar, policyFactory)
	if err != nil {
		return backend.Parser{}, nil, err
	}

	// The parser carries the augmented grammar the builder worked on, which is the grammar the policy has to be made
	// from, see conflict.PolicyFactory.
	conflictPolicy := policyFactory(parser.Grammar)

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

// GrammarToUnresolvedParser calculates the canonical LR(1) parser tables of the grammar and stops there, so the tables
// come back with their conflicts intact and their unreachable states in place, as definition 2.34 of IELR(1) describes
// them and before phase 5 decides anything.
//
// This is what the oracle and differential testing work is after, because a table whose conflicts are already resolved
// cannot be compared on where it has conflicts. It saves those callers from augmenting the grammar and driving the
// builder themselves, and it gives the three GoLR cores one shape to be called with. Reach for GrammarToParser whenever
// the tables are meant for a backend, because a table with conflicts in it is not a parser.
//
// Canonical LR(1) needs no policy to be constructed: it splits every state its items ask for, no matter how the
// conflicts are decided afterwards. The factory is taken all the same, so that the three GoLR cores agree on their
// signature and a caller can switch between them, and because IELR(1) does need the policy while it builds.
func GrammarToUnresolvedParser(
	grammar frontend.Grammar,
	policyFactory conflict.PolicyFactory,
) (backend.Parser, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Cores: LR1: GrammarToUnresolvedParser").End()

	// The builder works on the augmented grammar, so the caller hands us the grammar as the frontend produced it and
	// we augment it here, the same way the LALR(1) and IELR(1) cores do.
	augmentedGrammar := frontend.AugmentGrammar(grammar)

	builder := NewLR1Builder(augmentedGrammar)
	if err := builder.Build(); err != nil {
		return backend.Parser{}, err
	}
	return builder.Parser(), nil
}
