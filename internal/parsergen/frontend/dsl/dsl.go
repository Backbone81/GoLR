package dsl

import "golr/internal/parsergen/frontend"

// Grammar describes the context free grammar.
type Grammar struct {
	idxForTerminal    map[string]int
	idxForNonterminal map[string]int

	result            frontend.Grammar
	currentPrecedence int
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

// TerminalWithAlias adds the given terminal with an alias to the grammar. Returns a symbol reference to it. If the
// terminal already exists in the grammar, the existing terminal is returned.
func (g *Grammar) TerminalWithAlias(name string, alias string) frontend.SymbolRef {
	symbolRef := g.Terminal(name)
	g.result.Terminals[symbolRef.Idx()].Alias = alias
	g.idxForTerminal[alias] = symbolRef.Idx()
	return symbolRef
}

// Left sets the left associativity for the listed symbols.
func (g *Grammar) Left(symbolRefs ...frontend.SymbolRef) {
	g.currentPrecedence++
	for _, symbolRef := range symbolRefs {
		g.setAssociativity(symbolRef, frontend.AssociativityLeft)
	}
}

// Right sets the right associativity for the listed symbols.
func (g *Grammar) Right(symbolRefs ...frontend.SymbolRef) {
	g.currentPrecedence++
	for _, symbolRef := range symbolRefs {
		g.setAssociativity(symbolRef, frontend.AssociativityRight)
	}
}

// Nonassoc sets the listed symbols as not being associative.
func (g *Grammar) Nonassoc(symbolRefs ...frontend.SymbolRef) {
	g.currentPrecedence++
	for _, symbolRef := range symbolRefs {
		g.setAssociativity(symbolRef, frontend.AssociativityNone)
	}
}

// Precedence sets the precedence for the listed symbols.
func (g *Grammar) Precedence(symbolRefs ...frontend.SymbolRef) {
	g.currentPrecedence++
	for _, symbolRef := range symbolRefs {
		g.setAssociativity(symbolRef, frontend.AssociativityUndeclared)
	}
}

func (g *Grammar) setAssociativity(symbolRef frontend.SymbolRef, associativity frontend.Associativity) {
	if !symbolRef.IsTerminal() {
		panic("terminal expected for setting associativity and precedence")
	}
	g.result.Terminals[symbolRef.Idx()].Associativity = associativity
	g.result.Terminals[symbolRef.Idx()].Precedence = g.currentPrecedence
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
func (g *Grammar) Production(nonterminal frontend.SymbolRef) *ProductionBuilder {
	if !nonterminal.IsNonterminal() {
		panic("nonterminal expected on left hand side of the production")
	}

	return &ProductionBuilder{
		grammar: g,
		lhs:     nonterminal,
	}
}

type ProductionBuilder struct {
	grammar       *Grammar
	lhs           frontend.SymbolRef
	productionIdx int
	rhsSeen       bool
}

// Rhs defines all symbols on the right hand side of the production.
func (b *ProductionBuilder) Rhs(symbolRefs ...frontend.SymbolRef) *ProductionBuilder {
	if len(b.grammar.result.Productions) == 0 {
		b.grammar.result.StartNonterminalIdx = b.lhs.Idx()
	}

	b.grammar.result.Productions = append(b.grammar.result.Productions, frontend.Production{
		NonterminalIdx: b.lhs.Idx(),
		SymbolRefs:     symbolRefs,
	})

	// We need to keep the production index around for a later call to Prec. In case some other production was added
	// in between, the stored index makes sure we do not modify the wrong production.
	b.productionIdx = len(b.grammar.result.Productions) - 1
	b.rhsSeen = true
	return b
}

// Prec sets the precedence of the production to the precedence of the given terminal.
func (b *ProductionBuilder) Prec(symbolRef frontend.SymbolRef) {
	if !symbolRef.IsTerminal() {
		panic("terminal expected for prec")
	}
	if !b.rhsSeen {
		panic("the right hand side of the production must be defined before setting the precedence")
	}

	terminalIdx := symbolRef.Idx()
	b.grammar.result.Productions[b.productionIdx].PrecedenceTerminalIdx = &terminalIdx
}
