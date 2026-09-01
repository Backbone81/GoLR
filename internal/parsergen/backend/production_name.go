package backend

import (
	"strconv"

	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// ProductionNames returns a name for every production of the augmented grammar, index aligned with grammar.Productions.
// Index 0 is the $accept production and stays empty: it is never reduced and never becomes a parse node.
//
// A production with an explicit Name keeps it. Every other production gets the automatic name <lhs>_<n>, where n is its
// 1-based position among all productions sharing its left hand side, explicitly named ones counted. An automatic name
// which clashes with an explicit name gets a _1, _2, ... suffix until it is free.
func ProductionNames(grammar frontend.Grammar) []string {
	taken := make(map[string]struct{})
	for _, production := range grammar.Productions {
		if production.Name != nil {
			taken[*production.Name] = struct{}{}
		}
	}

	names := make([]string, len(grammar.Productions))
	countByNonterminalIdx := make(map[int]int)
	for i, production := range grammar.Productions {
		if i == 0 {
			continue
		}
		countByNonterminalIdx[production.NonterminalIdx]++
		if production.Name != nil {
			names[i] = *production.Name
			continue
		}

		base := grammar.Nonterminals[production.NonterminalIdx].Name + "_" +
			strconv.Itoa(countByNonterminalIdx[production.NonterminalIdx])
		name := base
		for suffix := 1; ; suffix++ {
			if _, clash := taken[name]; !clash {
				break
			}
			name = base + "_" + strconv.Itoa(suffix)
		}
		taken[name] = struct{}{}
		names[i] = name
	}
	return names
}
