package backend

import (
	"context"
	"runtime/trace"

	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

// NeverReducedWarnings returns a warning for every production which no state of the finished parser tables reduces,
// in production index order. A state reduces a production when it has a reduce action for it with a non-empty
// lookahead set or has it as its default reduction.
//
// It skips the accept production and the productions of nonterminals which are unreachable from the start nonterminal,
// because those are already reported by frontend.Grammar.Validate. A production which is only unreducible because
// another production is never reduced is not reported either, the warning for the cause is enough to act on.
//
// Run it after conflict.RemoveUnreachableStates, so that a reduction left only in a stranded state does not count.
func NeverReducedWarnings(parser Parser) []utils.Warning {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Backend: NeverReducedWarnings").End()

	reduced := make([]bool, len(parser.Grammar.Productions))
	for _, state := range parser.States {
		for _, reduceAction := range state.ReduceActions.All() {
			if !reduceAction.LookaheadSet.IsEmpty() {
				reduced[reduceAction.ProductionIdx] = true
			}
		}
		if state.DefaultReduceProductionIdx != nil {
			reduced[*state.DefaultReduceProductionIdx] = true
		}
	}

	reachable := frontend.ReachableNonterminals(parser.Grammar)
	var warnings []utils.Warning
	for productionIdx, production := range parser.Grammar.Productions {
		if productionIdx == 0 || reduced[productionIdx] || !reachable[production.NonterminalIdx] {
			continue
		}
		warnings = append(warnings, utils.Warningf(
			"production %s is never reduced after conflict resolution",
			frontend.FormatProduction(parser.Grammar, productionIdx),
		))
	}
	return warnings
}
