package dsl

import "golr/internal/parsergen/frontend"

// Grammar describes the context free grammar.
type Grammar struct {
	idxForTerminal    map[string]int
	idxForNonterminal map[string]int

	result frontend.Grammar
}

// NewGrammar creates a new grammar to add terminals, nonterminals and productions to.
func NewGrammar() *Grammar {
	return &Grammar{
		idxForTerminal:    make(map[string]int),
		idxForNonterminal: make(map[string]int),
	}
}

// Build returns the grammar described so far.
func (g *Grammar) Build() frontend.Grammar {
	return g.result
}

// Terminal adds the given terminal to the grammar. Returns a symbol reference to it. If the terminal already exists
// in the grammar, the existing terminal is returned.
func (g *Grammar) Terminal(name string) frontend.SymbolRef {
	idx, ok := g.idxForTerminal[name]
	if !ok {
		g.result.Terminals = append(g.result.Terminals, frontend.Symbol{
			Name: name,
		})
		idx = len(g.result.Terminals) - 1
		g.idxForTerminal[name] = idx
	}
	return frontend.NewTerminalRef(idx)
}

// Nonterminal adds the given nonterminal to the grammar. Returns a symbol reference to it. If the nonterminal already
// exists in the grammar, the existing nonterminal is returned.
func (g *Grammar) Nonterminal(name string) frontend.SymbolRef {
	idx, ok := g.idxForNonterminal[name]
	if !ok {
		g.result.Nonterminals = append(g.result.Nonterminals, frontend.Symbol{
			Name: name,
		})
		idx = len(g.result.Nonterminals) - 1
		g.idxForNonterminal[name] = idx
	}
	return frontend.NewNonterminalRef(idx)
}

// Production adds a production with the nonterminal on the left hand side and the symbols on the right hand side.
func (g *Grammar) Production(nonterminal frontend.SymbolRef, symbols ...frontend.SymbolRef) {
	if !nonterminal.IsNonterminal() {
		panic("nonterminal expected on left hand side of the production")
	}
	if len(g.result.Productions) == 0 {
		g.result.StartNonterminalIdx = nonterminal.Idx()
	}
	g.result.Productions = append(g.result.Productions, frontend.Production{
		NonterminalIdx: nonterminal.Idx(),
		SymbolRefs:     symbols,
	})
}
