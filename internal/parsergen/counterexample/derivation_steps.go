package counterexample

import (
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// derivationStep is the first production of a shortest derivation, see DerivationStartingWith and EmptyDerivation.
type derivationStep struct {
	// productionIdx is the production of the step, or noProductionIdx when there is no such derivation.
	productionIdx int

	// position is the position of the symbol which begins with the terminal. It is not used for an empty derivation.
	position int

	// cost is the number of productions of the whole derivation.
	cost int
}

// noProductionIdx stands for a missing production, like the step of a derivation which does not exist.
const noProductionIdx = -1

// DerivationStartingWith returns the first step of a shortest derivation of the nonterminal whose leaves begin with the
// terminal. It completes a counterexample where a nonterminal follows the dot, but the conflict terminal has to follow
// it (section 4, the example of `num • <digit> ? stmt stmt`).
//
// The step is a production of the nonterminal, and the position of the symbol in it which begins with the terminal. The
// symbols in front of that position derive the empty string, see EmptyDerivation. The symbol at the position is the
// terminal, or a nonterminal whose derivation continues with DerivationStartingWith. The symbols behind it are not
// expanded, because nonterminals stay nonterminals where terminals are not germane (section 3.2). It reports false when
// no derivation of the nonterminal begins with the terminal.
//
// A derivation is shorter than another when it uses fewer productions. The steps are computed on first use of the
// terminal.
func (t *LookupTables) DerivationStartingWith(nonterminalIdx int, terminalIdx int) (int, int, bool) {
	steps, found := t.derivationStepsByTerminalIdx[terminalIdx]
	if !found {
		steps = t.computeDerivationStepsStartingWith(terminalIdx)
		t.derivationStepsByTerminalIdx[terminalIdx] = steps
	}
	step := steps[nonterminalIdx]
	return step.productionIdx, step.position, step.productionIdx != noProductionIdx
}

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

// computeDerivationStepsStartingWith computes the first step of the shortest derivation beginning with the terminal for
// every nonterminal, in a fixed-point computation which repeats until no derivation gets shorter.
func (t *LookupTables) computeDerivationStepsStartingWith(terminalIdx int) []derivationStep {
	steps := newDerivationSteps(len(t.grammar.Nonterminals))
	changed := true
	for changed {
		changed = false
		for productionIdx, production := range t.grammar.Productions {
			step := t.shortestDerivationStepStartingWith(productionIdx, terminalIdx, steps)
			if step.productionIdx != noProductionIdx && updateDerivationStep(&steps[production.NonterminalIdx], step) {
				changed = true
			}
		}
	}
	return steps
}

// shortestDerivationStepStartingWith returns the shortest derivation which begins with the terminal and starts with
// the production, as far as the steps know yet. Every symbol of the production can begin with the terminal, as long as
// the symbols in front of it vanish.
func (t *LookupTables) shortestDerivationStepStartingWith(
	productionIdx int,
	terminalIdx int,
	steps []derivationStep,
) derivationStep {
	result := derivationStep{productionIdx: noProductionIdx}
	// The production itself, plus the derivations of the empty string of the symbols in front of the position.
	prefixCost := 1
	for position, symbolRef := range t.grammar.Productions[productionIdx].SymbolRefs {
		if symbolRef.IsTerminal() {
			if symbolRef.Idx() == terminalIdx {
				updateDerivationStep(
					&result,
					derivationStep{productionIdx: productionIdx, position: position, cost: prefixCost},
				)
			}
			return result
		}

		if step := steps[symbolRef.Idx()]; step.productionIdx != noProductionIdx {
			updateDerivationStep(
				&result,
				derivationStep{productionIdx: productionIdx, position: position, cost: prefixCost + step.cost},
			)
		}
		emptyStep := t.emptyDerivationStepByNonterminalIdx[symbolRef.Idx()]
		if emptyStep.productionIdx == noProductionIdx {
			return result
		}
		prefixCost += emptyStep.cost
	}
	return result
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
