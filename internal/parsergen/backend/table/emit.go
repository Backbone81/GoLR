package table

import (
	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/utils"
)

// This file holds the last step between a CompressedParser and a backend which writes the tables out as source code.
// Every table driven backend needs exactly these transformations, and they are here rather than in one of the backends
// because the alternative is a copy of them per language.
//
// What they have in common is that they take a value a Go slice can hold but a generated table should not - an absent
// entry reported as a negative number - and replace it with one the narrowest unsigned type of the target language can
// hold. A generated table which stays free of negative entries is what lets every backend pick an unsigned type for it,
// which is what keeps the tables small and keeps a language with an unsigned byte from having to widen them.

// FillHoles replaces the absent entries of a table with zeros, so that its entry type stays free of a negative value
// and can be an unsigned one.
//
// Nothing reads what this writes. A hole of a packed table is a cell the check table keeps the driver from using, and
// an absent default goto belongs to a nonterminal which no state has a goto on, so no reduction ever asks for it.
func FillHoles(values []int) []int {
	result := make([]int, len(values))
	for cellIdx, value := range values {
		result[cellIdx] = max(value, 0)
	}
	return result
}

// FillChecks replaces the holes of a check table with the given value, which is one past the highest column in use and
// therefore matches no lookup.
func FillChecks(values []int, noColumn int) []int {
	result := make([]int, len(values))
	for cellIdx, value := range values {
		result[cellIdx] = value
		if value == utils.NoColumn {
			result[cellIdx] = noColumn
		}
	}
	return result
}

// DefaultActions returns the action every state takes for a terminal its row has no entry for.
//
// A state without a default action carries the error action instead of the absent entry the table package reports. Both
// make such a terminal a syntax error, but the error action is a value like any other, which keeps the emitted table
// free of a negative entry and therefore lets it use an unsigned type.
func DefaultActions(compressed CompressedParser) []int {
	result := make([]int, len(compressed.DefaultActionByStateIdx))
	for stateIdx, action := range compressed.DefaultActionByStateIdx {
		if action == NoAction {
			action = NewErrorAction()
		}
		result[stateIdx] = int(action)
	}
	return result
}

// PopCounts returns, for every production, the number of symbols a reduction by it takes off the stacks. This is the
// length of the right hand side of the production and needs no compression, so it is read off the grammar rather than
// built from the compressed tables.
func PopCounts(parser backend.Parser) []int {
	result := make([]int, len(parser.Grammar.Productions))
	for productionIdx, production := range parser.Grammar.Productions {
		result[productionIdx] = len(production.SymbolRefs)
	}
	return result
}

// Nonterminals returns, for every production, the nonterminal on its left hand side. A reduction looks up its goto with
// it and labels the node it pushes with it.
func Nonterminals(parser backend.Parser) []int {
	result := make([]int, len(parser.Grammar.Productions))
	for productionIdx, production := range parser.Grammar.Productions {
		result[productionIdx] = production.NonterminalIdx
	}
	return result
}
