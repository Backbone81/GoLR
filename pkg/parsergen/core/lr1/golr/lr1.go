package golr

import (
	intconflict "github.com/backbone81/golr/internal/parsergen/conflict"
	intcore "github.com/backbone81/golr/internal/parsergen/core"
	intlr1golr "github.com/backbone81/golr/internal/parsergen/core/lr1/golr"
	"github.com/backbone81/golr/pkg/parsergen/backend"
	"github.com/backbone81/golr/pkg/parsergen/conflict"
	"github.com/backbone81/golr/pkg/parsergen/core"
	"github.com/backbone81/golr/pkg/parsergen/frontend"
)

// GrammarToParser calculates a parser from the context free grammar.
func GrammarToParser(grammar frontend.Grammar, options ...core.Option) (backend.Parser, []conflict.Conflict, error) {
	config := intcore.ConfigFromOptions(options...)
	policyFactory := intconflict.SelectPolicy(config.FailOnShiftReduceConflicts, config.FailOnReduceReduceConflicts)
	return intlr1golr.GrammarToParser(grammar, policyFactory, options...)
}
