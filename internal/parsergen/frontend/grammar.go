package frontend

import (
	"errors"
	"fmt"
	"strings"
)

// Grammar is a context free grammar.
type Grammar struct {
	// Terminals is the list of terminals which occur in the grammar.
	Terminals []Symbol `json:"terminals" yaml:"terminals"`

	// Nonterminals is the list of nonterminals which occur in the grammar.
	Nonterminals []Symbol `json:"nonterminals" yaml:"nonterminals"`

	// Productions is the list of productions describing all rules of the grammar. The productions reference terminals
	// and nonterminals defined in Terminals and Nonterminals.
	Productions []Production `json:"productions" yaml:"productions"`

	// StartNonterminalIdx is the nonterminal index which marks the start of the grammar.
	StartNonterminalIdx int `json:"startNonterminalIdx" yaml:"startNonterminalIdx"`
}

// Validate checks if the grammar is correct. If the validation fails, an error with the details about the failure is
// returned.
func (g Grammar) Validate() error {
	// we expect the start nonterminal to reference an existing nonterminal
	if g.StartNonterminalIdx < 0 || len(g.Nonterminals) <= g.StartNonterminalIdx {
		return errors.New("start nonterminal index out of bounds")
	}

	// we expect to have at least one production
	if len(g.Productions) < 1 {
		return errors.New("at least one production is required for a valid grammar")
	}
	for i, production := range g.Productions {
		// we expect the left hand side to reference an existing nonterminal
		if production.NonterminalIdx < 0 || len(g.Nonterminals) <= production.NonterminalIdx {
			return fmt.Errorf("nonterminal index out of bounds on the left hand side on production #%d", i)
		}
		for j, symbolRef := range production.SymbolRefs {
			if symbolRef.IsTerminal() {
				// we expect the right hand side to reference existing terminals
				if len(g.Terminals) <= symbolRef.Idx() {
					return fmt.Errorf("terminal index out of bounds on the right hand side for symbol #%d on production #%d", j, i)
				}
			} else {
				// we expect the right hand side to reference existing nonterminals
				if len(g.Nonterminals) <= symbolRef.Idx() {
					return fmt.Errorf("nonterminal index out of bounds on the right hand side for symbol #%d on production #%d", j, i)
				}
			}
		}
	}
	return nil
}

// RenumberNonterminalsInDeclarationOrder rewrites all nonterminal indices of the grammar so they follow declaration
// order, the order in which nonterminals first appear on a production left hand side. Frontends intern nonterminals in
// order of first appearance, which also counts right hand side references and therefore can assign a lower index to a
// nonterminal that is used before it is declared. Renumbering into declaration order matches the numbering the
// Bison-backed core assigns, which keeps generated parsers diff-friendly when switching cores.
//
// Any nonterminal that never appears on a left hand side (referenced but never defined) keeps a stable position after
// the declared ones, in its original order, so this function is safe to call on grammars which have not been validated.
func RenumberNonterminalsInDeclarationOrder(grammar *Grammar) {
	// Build the permutation from old index to new index by walking the productions in source order and assigning the
	// next new index the first time a left hand side nonterminal is seen.
	newIdxByOldIdx := make([]int, len(grammar.Nonterminals))
	for i := range newIdxByOldIdx {
		newIdxByOldIdx[i] = -1
	}
	nextNewIdx := 0
	for _, production := range grammar.Productions {
		if newIdxByOldIdx[production.NonterminalIdx] == -1 {
			newIdxByOldIdx[production.NonterminalIdx] = nextNewIdx
			nextNewIdx++
		}
	}
	for oldIdx := range grammar.Nonterminals {
		if newIdxByOldIdx[oldIdx] == -1 {
			newIdxByOldIdx[oldIdx] = nextNewIdx
			nextNewIdx++
		}
	}

	// Reorder the nonterminal symbols into their new positions.
	reorderedNonterminals := make([]Symbol, len(grammar.Nonterminals))
	for oldIdx, newIdx := range newIdxByOldIdx {
		reorderedNonterminals[newIdx] = grammar.Nonterminals[oldIdx]
	}
	grammar.Nonterminals = reorderedNonterminals

	// Remap every reference to a nonterminal: the left hand side of each production, the nonterminal symbols on the
	// right hand sides and the start nonterminal.
	for i := range grammar.Productions {
		production := &grammar.Productions[i]
		production.NonterminalIdx = newIdxByOldIdx[production.NonterminalIdx]
		for j, symbolRef := range production.SymbolRefs {
			if symbolRef.IsNonterminal() {
				production.SymbolRefs[j] = NewNonterminalRef(newIdxByOldIdx[symbolRef.Idx()])
			}
		}
	}
	grammar.StartNonterminalIdx = newIdxByOldIdx[grammar.StartNonterminalIdx]
}

// Grammar implements fmt.Stringer.
var _ fmt.Stringer = (*Grammar)(nil)

// String returns a string representation of the grammar.
func (g Grammar) String() string {
	var builder strings.Builder

	for i := range g.Terminals {
		fmt.Fprintf(&builder, "terminal %d: %s\n", i, g.Terminals[i])
	}
	builder.WriteString("\n")

	for i := range g.Nonterminals {
		fmt.Fprintf(&builder, "nonterminal %d: %s\n", i, g.Nonterminals[i])
	}
	builder.WriteString("\n")

	for i := range g.Productions {
		fmt.Fprintf(&builder, "production %d: %s\n", i, g.Productions[i])
	}
	builder.WriteString("\n")

	fmt.Fprintf(&builder, "start nonterminal: %d\n", g.StartNonterminalIdx)
	builder.WriteString("\n")
	return builder.String()
}

// AugmentGrammar returns an augmented grammar where a new start symbol was introduced with a production where the right
// hand side is the old start symbol.
func AugmentGrammar(grammar Grammar) Grammar {
	augmentedGrammar := Grammar{
		Terminals:           make([]Symbol, 0, len(grammar.Terminals)+1),
		Nonterminals:        make([]Symbol, 0, len(grammar.Nonterminals)+1),
		Productions:         make([]Production, 0, len(grammar.Productions)+1),
		StartNonterminalIdx: 0,
	}

	augmentedGrammar.Terminals = append(augmentedGrammar.Terminals, SymbolEOF)
	augmentedGrammar.Terminals = append(augmentedGrammar.Terminals, grammar.Terminals...)
	terminalOffset := len(augmentedGrammar.Terminals) - len(grammar.Terminals)

	augmentedGrammar.Nonterminals = append(augmentedGrammar.Nonterminals, Symbol{Name: "$accept"})
	augmentedGrammar.Nonterminals = append(augmentedGrammar.Nonterminals, grammar.Nonterminals...)
	nonterminalOffset := len(augmentedGrammar.Nonterminals) - len(grammar.Nonterminals)

	augmentedGrammar.Productions = append(augmentedGrammar.Productions, Production{
		NonterminalIdx: 0, // the new start nonterminal
		SymbolRefs: []SymbolRef{
			// the old start symbol index was moved back by one because of the new start symbol inserted at the start
			// of the nonterminal list
			NewNonterminalRef(grammar.StartNonterminalIdx + nonterminalOffset),

			// the new EOF symbol which was inserted at the start of the terminal list
			NewTerminalRef(0),
		},
	})
	augmentedGrammar.Productions = append(augmentedGrammar.Productions, grammar.Productions...)

	// Now we need to adjust the symbol indexes for all productions we copied over from the old grammar to accommodate
	// for the additional terminal and nonterminal we inserted.
	for i := 1; i < len(augmentedGrammar.Productions); i++ {
		production := &augmentedGrammar.Productions[i]
		production.NonterminalIdx += nonterminalOffset

		if production.PrecedenceTerminalIdx != nil {
			// The terminal a production takes its precedence from explicitly is a terminal index like any other, so it
			// moves along with the terminals. NOTE: We need to point at a new value instead of adding to the existing
			// one, because the productions were copied over from the old grammar and still share the value with it.
			precedenceTerminalIdx := *production.PrecedenceTerminalIdx + terminalOffset
			production.PrecedenceTerminalIdx = &precedenceTerminalIdx
		}

		if len(production.SymbolRefs) == 0 {
			// We only want to augment symbol indexes when there are symbols at all.
			continue
		}

		// NOTE: We need to make a copy if the symbol indexes slice, otherwise we modify the existing one
		augmentedSymbolRefs := make([]SymbolRef, len(production.SymbolRefs))
		for j := range production.SymbolRefs {
			symbolRef := production.SymbolRefs[j]
			if symbolRef.IsNonterminal() {
				augmentedSymbolRefs[j] = NewNonterminalRef(symbolRef.Idx() + nonterminalOffset)
			} else {
				augmentedSymbolRefs[j] = NewTerminalRef(symbolRef.Idx() + terminalOffset)
			}
		}
		production.SymbolRefs = augmentedSymbolRefs
	}
	return augmentedGrammar
}
