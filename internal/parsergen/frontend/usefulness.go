package frontend

import (
	"errors"
	"fmt"

	"github.com/backbone81/golr/internal/utils"
)

// checkUsefulness returns an error for every unproductive nonterminal and a warning for every nonterminal which is
// unreachable from the start nonterminal, both in nonterminal index order. The grammar must have valid indices.
func checkUsefulness(g Grammar) ([]utils.Warning, error) {
	productive := productiveNonterminals(g)
	reachable := ReachableNonterminals(g)

	hasProductions := make([]bool, len(g.Nonterminals))
	for _, production := range g.Productions {
		hasProductions[production.NonterminalIdx] = true
	}

	var errs []error
	var warnings []utils.Warning
	for nonterminalIdx, nonterminal := range g.Nonterminals {
		if !productive[nonterminalIdx] {
			kind := "nonterminal"
			if nonterminalIdx == g.StartNonterminalIdx {
				kind = "start nonterminal"
			}
			if hasProductions[nonterminalIdx] {
				errs = append(errs, fmt.Errorf("%s %q does not derive any finite string", kind, nonterminal.Name))
			} else {
				errs = append(errs, fmt.Errorf("%s %q has no productions", kind, nonterminal.Name))
			}
		}
		if !reachable[nonterminalIdx] {
			warnings = append(warnings, utils.Warningf(
				"nonterminal %q is unreachable from the start nonterminal %q",
				nonterminal.Name,
				g.Nonterminals[g.StartNonterminalIdx].Name,
			))
		}
	}
	return warnings, errors.Join(errs...)
}

// productiveNonterminals reports for every nonterminal if it derives at least one finite string of terminals.
func productiveNonterminals(g Grammar) []bool {
	analysis := newProductivityAnalysis(g)
	analysis.run()
	return analysis.productive
}

// productivityAnalysis finds the productive nonterminals. A production without nonterminals on its right hand side
// makes its left hand side productive right away. Every other production counts the nonterminals on its right hand side
// which are not confirmed productive yet. Confirming a nonterminal decrements the count of every production using it,
// and a production reaching zero confirms its left hand side in turn. Nonterminals never confirmed are unproductive.
// Every right hand side symbol is visited once, so the cost is linear in the grammar size.
type productivityAnalysis struct {
	grammar Grammar

	// pendingByProductionIdx counts the nonterminals on the right hand side not confirmed productive yet.
	pendingByProductionIdx []int

	// productionIdxsByNonterminalIdx lists the productions using a nonterminal on the right hand side, once per
	// occurrence.
	productionIdxsByNonterminalIdx [][]int

	// productive holds the confirmed nonterminals.
	productive []bool

	// worklist holds the confirmed nonterminals whose productions have not been decremented yet.
	worklist []int
}

// newProductivityAnalysis counts the pending nonterminal occurrences of every production and indexes the productions
// by the nonterminals on their right hand sides.
func newProductivityAnalysis(g Grammar) *productivityAnalysis {
	analysis := &productivityAnalysis{
		grammar:                        g,
		pendingByProductionIdx:         make([]int, len(g.Productions)),
		productionIdxsByNonterminalIdx: make([][]int, len(g.Nonterminals)),
		productive:                     make([]bool, len(g.Nonterminals)),
	}
	for productionIdx, production := range g.Productions {
		for _, symbolRef := range production.SymbolRefs {
			if symbolRef.IsNonterminal() {
				analysis.pendingByProductionIdx[productionIdx]++
				analysis.productionIdxsByNonterminalIdx[symbolRef.Idx()] = append(
					analysis.productionIdxsByNonterminalIdx[symbolRef.Idx()],
					productionIdx,
				)
			}
		}
	}
	return analysis
}

// run marks the left hand sides of the productions without nonterminals as productive and propagates from there.
func (a *productivityAnalysis) run() {
	for productionIdx, pending := range a.pendingByProductionIdx {
		if pending == 0 {
			a.markProductive(a.grammar.Productions[productionIdx].NonterminalIdx)
		}
	}
	for len(a.worklist) > 0 {
		nonterminalIdx := a.worklist[len(a.worklist)-1]
		a.worklist = a.worklist[:len(a.worklist)-1]
		for _, productionIdx := range a.productionIdxsByNonterminalIdx[nonterminalIdx] {
			a.pendingByProductionIdx[productionIdx]--
			if a.pendingByProductionIdx[productionIdx] == 0 {
				a.markProductive(a.grammar.Productions[productionIdx].NonterminalIdx)
			}
		}
	}
}

// markProductive marks the nonterminal as productive and queues it, unless it is productive already.
func (a *productivityAnalysis) markProductive(nonterminalIdx int) {
	if a.productive[nonterminalIdx] {
		return
	}
	a.productive[nonterminalIdx] = true
	a.worklist = append(a.worklist, nonterminalIdx)
}

// ReachableNonterminals reports for every nonterminal if the start nonterminal derives a sentential form containing it.
// It is a breadth-first search starting at the start nonterminal: every nonterminal on the right hand side of a
// production of a reached nonterminal is reached as well. Every production is visited at most once, so the cost is
// linear in the grammar size.
func ReachableNonterminals(g Grammar) []bool {
	productionIdxsByNonterminalIdx := make([][]int, len(g.Nonterminals))
	for productionIdx, production := range g.Productions {
		productionIdxsByNonterminalIdx[production.NonterminalIdx] = append(
			productionIdxsByNonterminalIdx[production.NonterminalIdx],
			productionIdx,
		)
	}

	reachable := make([]bool, len(g.Nonterminals))
	reachable[g.StartNonterminalIdx] = true
	queue := []int{g.StartNonterminalIdx}
	for len(queue) > 0 {
		nonterminalIdx := queue[0]
		queue = queue[1:]
		for _, productionIdx := range productionIdxsByNonterminalIdx[nonterminalIdx] {
			for _, symbolRef := range g.Productions[productionIdx].SymbolRefs {
				if symbolRef.IsNonterminal() && !reachable[symbolRef.Idx()] {
					reachable[symbolRef.Idx()] = true
					queue = append(queue, symbolRef.Idx())
				}
			}
		}
	}
	return reachable
}
