package golr

import (
	intconflict "github.com/backbone81/golr/internal/parsergen/conflict"
	intcore "github.com/backbone81/golr/internal/parsergen/core"
	intlr1golr "github.com/backbone81/golr/internal/parsergen/core/lr1/golr"
	"github.com/backbone81/golr/pkg/parsergen/backend"
	"github.com/backbone81/golr/pkg/parsergen/conflict"
	"github.com/backbone81/golr/pkg/parsergen/core"
	"github.com/backbone81/golr/pkg/parsergen/frontend"
	"github.com/backbone81/golr/pkg/utils"
)

// GrammarToParser calculates a parser from the context free grammar.
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
	return intlr1golr.GrammarToParser(grammar, policyFactory, options...)
}
