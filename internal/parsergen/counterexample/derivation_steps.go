package counterexample

import (
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// derivationStep is the first production of a shortest derivation, see EmptyDerivation.
type derivationStep struct {
	// productionIdx is the production of the step, or noProductionIdx when there is no such derivation.
	productionIdx int

	// cost is the number of productions of the whole derivation.
	cost int
}

// noProductionIdx stands for a missing production, like the step of a derivation which does not exist.
const noProductionIdx = -1

// EmptyDerivation returns the production of a shortest derivation of the empty string from the nonterminal. Every
// nonterminal on the right hand side of the production derives the empty string in turn. It reports false when the
// nonterminal is not nullable.
func (t *LookupTables) EmptyDerivation(nonterminalIdx int) (int, bool) {
	productionIdx := t.emptyDerivationStepByNonterminalIdx[nonterminalIdx].productionIdx
	return productionIdx, productionIdx != noProductionIdx
}

// initEmptyDerivationSteps computes the shortest derivation of the empty string from every nullable nonterminal, in a
// fixed-point computation which repeats until no derivation gets shorter.
func (t *LookupTables) initEmptyDerivationSteps() {
	t.emptyDerivationStepByNonterminalIdx = newDerivationSteps(len(t.grammar.Nonterminals))
	changed := true
	for changed {
		changed = false
		for productionIdx, production := range t.grammar.Productions {
			cost, found := t.emptyDerivationCost(production)
			if found && updateDerivationStep(
				&t.emptyDerivationStepByNonterminalIdx[production.NonterminalIdx],
				derivationStep{productionIdx: productionIdx, cost: cost},
			) {
				changed = true
			}
		}
	}
}

// emptyDerivationCost returns the number of productions of the shortest derivation of the empty string which starts
// with the production, as far as it is known yet. It reports false when a symbol of the production cannot vanish yet.
func (t *LookupTables) emptyDerivationCost(production frontend.Production) (int, bool) {
	cost := 1
	for _, symbolRef := range production.SymbolRefs {
		if symbolRef.IsTerminal() {
			return 0, false
		}
		step := t.emptyDerivationStepByNonterminalIdx[symbolRef.Idx()]
		if step.productionIdx == noProductionIdx {
			return 0, false
		}
		cost += step.cost
	}
	return cost, true
}

// newDerivationSteps returns a step for every nonterminal, none of which has a derivation yet.
func newDerivationSteps(count int) []derivationStep {
	result := make([]derivationStep, count)
	for i := range result {
		result[i].productionIdx = noProductionIdx
	}
	return result
}

// updateDerivationStep replaces the step with the candidate when the candidate is shorter, and reports if it did. A
// missing step is replaced by any candidate.
func updateDerivationStep(step *derivationStep, candidate derivationStep) bool {
	if step.productionIdx != noProductionIdx && step.cost <= candidate.cost {
		return false
	}
	*step = candidate
	return true
}
