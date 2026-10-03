package counterexample

import (
	"strings"

	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// Counterexample shows where a conflict between two items comes from, with one derivation per item.
type Counterexample struct {
	// Unifying reports if both derivations have the same leaves, which proves the grammar ambiguous.
	Unifying bool

	// NonterminalIdx is the nonterminal both derivations derive: the unifying nonterminal of a unifying counterexample,
	// the innermost nonterminal or the start nonterminal of a nonunifying one.
	NonterminalIdx int

	// Derivations holds the derivations D1 and D2 of the two parsers. The first uses the reduction of the conflict, the
	// second the shift or the other reduction.
	Derivations [2]Derivation
}

// Header renders the line which introduces the counterexample and names its nonterminal. Only a unifying
// counterexample shows an ambiguity, a nonunifying one leaves open whether the grammar is ambiguous.
func (c Counterexample) Header(grammar frontend.Grammar) string {
	name := grammar.Nonterminals[c.NonterminalIdx].String()
	if c.Unifying {
		return "ambiguous for " + name + ":"
	}
	return "conflict within " + name + ":"
}

// Lines renders the counterexample, one line per element: the leaves of the derivations as an example, followed by the
// two derivations with one production per line. A unifying counterexample has a single example, a nonunifying one an
// example in front of each derivation.
func (c Counterexample) Lines(grammar frontend.Grammar) []string {
	labels := [2]string{"using the reduction:", "using the shift:"}
	if c.Derivations[1].endsWithDot() {
		labels = [2]string{"using the first reduction:", "using the second reduction:"}
	}

	var lines []string
	for i, derivation := range c.Derivations {
		if i == 0 || !c.Unifying {
			lines = append(lines, "example: "+strings.Join(derivation.appendLeaves(grammar, nil), " "))
		}
		lines = append(lines, labels[i])
		lines = derivation.appendLines(grammar, "  ", lines)
	}
	return lines
}
