package golr

import (
	intconflict "github.com/backbone81/golr/internal/parsergen/conflict"
	intcore "github.com/backbone81/golr/internal/parsergen/core"
	intlalr1golr "github.com/backbone81/golr/internal/parsergen/core/lalr1/golr"
	"github.com/backbone81/golr/pkg/parsergen/backend"
	"github.com/backbone81/golr/pkg/parsergen/conflict"
	"github.com/backbone81/golr/pkg/parsergen/core"
	"github.com/backbone81/golr/pkg/parsergen/frontend"
	"github.com/backbone81/golr/pkg/utils"
)

// GrammarToParser calculates a parser from the context free grammar.
//
// Conflicts are resolved the way GNU Bison and Yacc do: precedence and associativity decide first, a shift beats a
// reduction when precedence has nothing to say, and the production which was declared first wins a conflict between two
// reductions.
//
// A conflict left unresolved, as with core.FailOnConflicts, makes it fail with one conflict.UnresolvedConflictError per
// conflict. The parser tables still come back then, holding the conflicting actions, so the conflicts can be reported
// with report.WriteUnresolvedConflictReport, but they must not be handed to a backend. A failure on warnings, as with
// core.FailOnWarnings, returns the finished tables as well.
func GrammarToParser(grammar frontend.Grammar, options ...core.Option) (
	backend.Parser,
	[]conflict.Conflict,
	[]utils.Warning,
	error,
) {
	config := intcore.ConfigFromOptions(options...)
	policyFactory := intconflict.SelectPolicy(config.FailOnShiftReduceConflicts, config.FailOnReduceReduceConflicts)
	return intlalr1golr.GrammarToParser(grammar, policyFactory, options...)
}
