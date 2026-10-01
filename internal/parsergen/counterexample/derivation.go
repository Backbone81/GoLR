package counterexample

import (
	"strings"

	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// Derivation is a node of a derivation tree: a symbol which is not expanded, a nonterminal expanded by a production,
// or the dot which marks the position of the conflict.
type Derivation struct {
	// Symbol is the symbol of the node. It is not used for the dot.
	Symbol frontend.SymbolRef

	// ProductionIdx is the production which expands the nonterminal, or noProductionIdx when the node is not expanded.
	ProductionIdx int

	// Children holds one node per symbol of the production, and the dot when the conflict lies within the production.
	Children []Derivation

	// Dot reports if the node is the dot.
	Dot bool
}

// NewLeafDerivation returns a symbol which is not expanded.
func NewLeafDerivation(symbol frontend.SymbolRef) Derivation {
	return Derivation{Symbol: symbol, ProductionIdx: noProductionIdx}
}

// NewDotDerivation returns the dot.
func NewDotDerivation() Derivation {
	return Derivation{ProductionIdx: noProductionIdx, Dot: true}
}

// NewExpandedDerivation returns the nonterminal of the production, expanded by the production into the children.
func NewExpandedDerivation(grammar frontend.Grammar, productionIdx int, children []Derivation) Derivation {
	return Derivation{
		Symbol:        frontend.NewNonterminalRef(grammar.Productions[productionIdx].NonterminalIdx),
		ProductionIdx: productionIdx,
		Children:      children,
	}
}

// IsExpanded reports if the node is a nonterminal expanded by a production.
func (d Derivation) IsExpanded() bool {
	return d.ProductionIdx != noProductionIdx
}

// endsWithDot reports if the dot of the derivation is behind the last symbol of its production, which makes the item of
// the dot a reduce item. It reports false when the derivation has no dot.
func (d Derivation) endsWithDot() bool {
	for i, child := range d.Children {
		if child.Dot {
			return i == len(d.Children)-1
		}
		if child.IsExpanded() && child.hasDot() {
			return child.endsWithDot()
		}
	}
	return false
}

// hasDot reports if the dot is part of the derivation.
func (d Derivation) hasDot() bool {
	if d.Dot {
		return true
	}
	for _, child := range d.Children {
		if child.hasDot() {
			return true
		}
	}
	return false
}

// appendLeaves appends the names of the leaves of the derivation, with the dot as •.
func (d Derivation) appendLeaves(grammar frontend.Grammar, leaves []string) []string {
	if !d.IsExpanded() {
		return append(leaves, d.leafName(grammar))
	}
	for _, child := range d.Children {
		leaves = child.appendLeaves(grammar, leaves)
	}
	return leaves
}

// appendLines appends one line per expanded node of the derivation, with the production of the node, and the nodes it
// expands in brackets. The lines of the expanded children follow in order, indented by two more columns.
func (d Derivation) appendLines(grammar frontend.Grammar, indent string, lines []string) []string {
	if !d.IsExpanded() {
		return lines
	}

	var builder strings.Builder
	builder.WriteString(indent)
	builder.WriteString(symbolName(grammar, d.Symbol))
	builder.WriteString(" ->")
	if len(grammar.Productions[d.ProductionIdx].SymbolRefs) == 0 {
		builder.WriteString(" (empty)")
	}
	for _, child := range d.Children {
		builder.WriteString(" ")
		if child.IsExpanded() {
			builder.WriteString("[" + symbolName(grammar, child.Symbol) + "]")
		} else {
			builder.WriteString(child.leafName(grammar))
		}
	}
	lines = append(lines, builder.String())

	for _, child := range d.Children {
		lines = child.appendLines(grammar, indent+"  ", lines)
	}
	return lines
}

// leafName returns the name of the symbol of the node, or • for the dot.
func (d Derivation) leafName(grammar frontend.Grammar) string {
	if d.Dot {
		return "•"
	}
	return symbolName(grammar, d.Symbol)
}

// symbolName returns the name of the symbol as the grammar writes it.
func symbolName(grammar frontend.Grammar, symbol frontend.SymbolRef) string {
	if symbol.IsTerminal() {
		return grammar.Terminals[symbol.Idx()].String()
	}
	return grammar.Nonterminals[symbol.Idx()].String()
}
