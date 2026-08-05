package table

import (
	"slices"

	"github.com/backbone81/golr/internal/parsergen/backend"
)

// NewDefaultGotos returns, for every nonterminal, the state which a goto on it leads to in most of the states which
// have such a goto, or NoGoto for a nonterminal which no state has a goto on.
//
// This is the goto table's counterpart of the default action of a state, see NewDefaultActions, and it is what makes
// the goto rows sparse. A goto row holds a handful of entries spread over all nonterminals, so the goto table of a
// large grammar is the biggest of the tables by some margin while packing far less densely than the action table.
// Taking the most frequent target of a nonterminal out of every row which agrees with it, see ApplyDefaultGotos,
// removes those entries from the table entirely.
//
// The default belongs to the nonterminal and not to the state, which is the other way around than for the actions. A
// state is asked for its action once per lookahead, so what its row has in common is a reduction it performs for many
// terminals, while a goto answers where one nonterminal takes the parser, and the states such a nonterminal leads to
// are few compared to the states which have a goto on it.
//
// Ties go to the lowest state index, so that a grammar always yields the same tables.
func NewDefaultGotos(parser backend.Parser) []int {
	targetsByNonterminalIdx := make([][]int, len(parser.Grammar.Nonterminals))
	for stateIdx := range parser.States {
		for _, transitionAction := range parser.States[stateIdx].TransitionActions.All() {
			if transitionAction.SymbolRef().IsTerminal() {
				// A transition on a terminal is a shift, which lives in a table of its own, see NewActionTable.
				continue
			}
			nonterminalIdx := transitionAction.SymbolRef().Idx()
			targetsByNonterminalIdx[nonterminalIdx] = append(
				targetsByNonterminalIdx[nonterminalIdx],
				transitionAction.StateIdx(),
			)
		}
	}

	result := make([]int, len(targetsByNonterminalIdx))
	for nonterminalIdx, targetStateIdxs := range targetsByNonterminalIdx {
		result[nonterminalIdx] = mostFrequentTarget(targetStateIdxs)
	}
	return result
}

// mostFrequentTarget returns the state which appears most often among the targets of the gotos on a single nonterminal,
// or NoGoto when there is no such goto at all.
//
// Sorting the targets puts equal ones next to each other, so the longest run of equal values names the state to default
// to. A run has to be strictly longer than the best one so far to win, which makes the lowest state index win a tie no
// matter in which order the gotos were collected.
func mostFrequentTarget(targetStateIdxs []int) int {
	slices.Sort(targetStateIdxs)

	result := NoGoto
	var resultCount int
	for runStart := 0; runStart < len(targetStateIdxs); {
		runEnd := runStart
		for runEnd < len(targetStateIdxs) && targetStateIdxs[runEnd] == targetStateIdxs[runStart] {
			runEnd++
		}
		if runEnd-runStart > resultCount {
			result, resultCount = targetStateIdxs[runStart], runEnd-runStart
		}
		runStart = runEnd
	}
	return result
}

// ApplyDefaultGotos removes from the given rows of the goto table every goto which the default goto of its nonterminal
// already covers. The rows are modified in place.
//
// This is what makes the default gotos a compression instead of just another table. What is left in a row are only the
// gotos which deviate from the default of their nonterminal, so the row displacement has that many fewer entries to
// place, and a state which deviates in none of its gotos keeps an empty row, which costs no cell at all.
//
// A lookup can no longer tell a state which has no goto on a nonterminal from a state whose goto the default covers,
// because both are an absent entry. Nothing reads such a lookup: a reduction only ever asks for the nonterminal it has
// just built, in a state whose goto the LR construction created together with the production, see NewGotoTable.
func ApplyDefaultGotos(rows [][]int, defaultGotoByNonterminalIdx []int) {
	for _, row := range rows {
		for nonterminalIdx, targetStateIdx := range row {
			if targetStateIdx == defaultGotoByNonterminalIdx[nonterminalIdx] {
				row[nonterminalIdx] = NoGoto
			}
		}
	}
}
