package frontend

import (
	"github.com/backbone81/golr/internal/utils"
)

// FirstSets holds the nullable nonterminals of a grammar and their first sets. A nonterminal is nullable when it can
// derive the empty string, and its first set holds the index of every terminal which can start a string derived from
// it.
type FirstSets struct {
	grammar                  Grammar
	nullableByNonterminalIdx []bool
	firstByNonterminalIdx    []utils.Bitset
}

// NewFirstSets computes the nullable nonterminals and the first sets of the grammar.
func NewFirstSets(grammar Grammar) FirstSets {
	result := FirstSets{
		grammar:                  grammar,
		nullableByNonterminalIdx: make([]bool, len(grammar.Nonterminals)),
		firstByNonterminalIdx:    make([]utils.Bitset, len(grammar.Nonterminals)),
	}
	result.compute()
	return result
}

// compute finds the nullable nonterminals and their first sets in a single fixed-point computation. The right hand side
// of each production is walked until a symbol is reached which cannot vanish. A right hand side which vanishes entirely
// makes its left hand side nullable.
func (f *FirstSets) compute() {
	changed := true
	for changed {
		changed = false
		for _, production := range f.grammar.Productions {
			firstSetChanged, nullable := f.mergeFirstOfSequence(
				production.SymbolRefs,
				&f.firstByNonterminalIdx[production.NonterminalIdx],
			)
			if firstSetChanged {
				changed = true
			}
			if nullable && !f.nullableByNonterminalIdx[production.NonterminalIdx] {
				f.nullableByNonterminalIdx[production.NonterminalIdx] = true
				changed = true
			}
		}
	}
}

// IsNullable reports if the nonterminal can derive the empty string.
func (f *FirstSets) IsNullable(nonterminalIdx int) bool {
	return f.nullableByNonterminalIdx[nonterminalIdx]
}

// IsSequenceNullable reports if every symbol of the sequence can vanish, which holds for the empty sequence.
func (f *FirstSets) IsSequenceNullable(symbolRefs []SymbolRef) bool {
	for _, symbolRef := range symbolRefs {
		if symbolRef.IsTerminal() || !f.nullableByNonterminalIdx[symbolRef.Idx()] {
			return false
		}
	}
	return true
}

// FirstOfSequence merges the terminals which can start the sequence into the set. It reports if the whole sequence can
// vanish, in which case whatever follows the sequence can start it as well.
func (f *FirstSets) FirstOfSequence(symbolRefs []SymbolRef, set *utils.Bitset) bool {
	_, nullable := f.mergeFirstOfSequence(symbolRefs, set)
	return nullable
}

// mergeFirstOfSequence merges the terminals which can start the sequence into the set. It returns two values:
//
//   - The first reports if the set grew, which is how compute notices that the fixed point is not reached yet.
//   - The second reports if the whole sequence can vanish, in which case whatever follows the sequence can start it as
//     well. It holds for the empty sequence.
//
// While compute runs, it works on the intermediate results.
func (f *FirstSets) mergeFirstOfSequence(symbolRefs []SymbolRef, set *utils.Bitset) (bool, bool) {
	changed := false
	for _, symbolRef := range symbolRefs {
		if symbolRef.IsTerminal() {
			changed = set.Add(symbolRef.Idx()) || changed
			return changed, false
		}

		changed = set.Merge(&f.firstByNonterminalIdx[symbolRef.Idx()]) || changed
		if !f.nullableByNonterminalIdx[symbolRef.Idx()] {
			return changed, false
		}
	}
	return changed, true
}
