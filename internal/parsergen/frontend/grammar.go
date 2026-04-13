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
	StartNonterminalIdx int `json:"startNonterminalIdx" yaml:"start_nonterminal_idx"`
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

// Grammar implements fmt.Stringer
var _ fmt.Stringer = (*Grammar)(nil)

// String returns a string representation of the grammar.
func (g Grammar) String() string {
	var builder strings.Builder

	for i := range g.Terminals {
		builder.WriteString(fmt.Sprintf("terminal %d: %s\n", i, g.Terminals[i]))
	}
	builder.WriteString("\n")

	for i := range g.Nonterminals {
		builder.WriteString(fmt.Sprintf("nonterminal %d: %s\n", i, g.Nonterminals[i]))
	}
	builder.WriteString("\n")

	for i := range g.Productions {
		builder.WriteString(fmt.Sprintf("production %d: %s\n", i, g.Productions[i]))
	}
	builder.WriteString("\n")

	builder.WriteString(fmt.Sprintf("start nonterminal: %d\n", g.StartNonterminalIdx))
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
