package frontend_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/frontend"
)

var _ = Describe("AugmentGrammar", func() {
	// Augmenting a grammar inserts the end of input marker in front of every terminal the grammar declared, which moves
	// every terminal index of the grammar back by one. The terminal a production takes its precedence from explicitly
	// is such a terminal index, so it has to move along with the terminals it refers to. A precedence which is not
	// moved silently points at the terminal in front of the one which was meant, which is the end of input marker for
	// the terminal the grammar declared first, and that terminal has no precedence at all.
	It("should move the explicit precedence terminal of a production along with the terminals", func() {
		precedenceTerminalIdx := 0
		grammar := frontend.Grammar{
			Terminals: []frontend.Symbol{
				{Name: "+", Precedence: 1, Associativity: frontend.AssociativityLeft},
				{Name: "a"},
			},
			Nonterminals: []frontend.Symbol{
				{Name: "E"},
			},
			Productions: []frontend.Production{
				{
					NonterminalIdx: 0,
					SymbolRefs: []frontend.SymbolRef{
						frontend.NewTerminalRef(1),
					},
					// The production takes the precedence of "+", which no symbol of its right hand side would give it.
					PrecedenceTerminalIdx: &precedenceTerminalIdx,
				},
			},
			StartNonterminalIdx: 0,
		}

		augmentedGrammar := frontend.AugmentGrammar(grammar)

		// The production of the old grammar is the second production of the augmented grammar, because augmenting puts
		// the production of the new start symbol in front of it.
		augmentedProduction := augmentedGrammar.Productions[1]
		Expect(augmentedProduction.PrecedenceTerminalIdx).ToNot(BeNil())
		Expect(augmentedGrammar.Terminals[*augmentedProduction.PrecedenceTerminalIdx].Name).To(
			Equal("+"),
			"the production is expected to still take its precedence from the terminal it declared",
		)

		// The grammar the caller handed in must be left alone, so the productions of the augmented grammar cannot share
		// the value with it.
		Expect(*grammar.Productions[0].PrecedenceTerminalIdx).To(
			Equal(0),
			"augmenting a grammar is expected to leave the grammar it was given untouched",
		)
	})

	// A production without an explicit precedence has nothing to move, and it inherits its precedence from the
	// rightmost terminal of its right hand side instead, which is moved as part of the right hand side.
	It("should leave a production without an explicit precedence terminal without one", func() {
		grammar := frontend.Grammar{
			Terminals: []frontend.Symbol{
				{Name: "a"},
			},
			Nonterminals: []frontend.Symbol{
				{Name: "E"},
			},
			Productions: []frontend.Production{
				{
					NonterminalIdx: 0,
					SymbolRefs: []frontend.SymbolRef{
						frontend.NewTerminalRef(0),
					},
				},
			},
			StartNonterminalIdx: 0,
		}

		augmentedGrammar := frontend.AugmentGrammar(grammar)

		Expect(augmentedGrammar.Productions[1].PrecedenceTerminalIdx).To(BeNil())
	})

	// An empty production has no right hand side to move, but it can still declare the terminal it takes its precedence
	// from, so it must not be skipped along with the symbols it does not have.
	It("should move the explicit precedence terminal of an empty production", func() {
		precedenceTerminalIdx := 0
		grammar := frontend.Grammar{
			Terminals: []frontend.Symbol{
				{Name: "+", Precedence: 1, Associativity: frontend.AssociativityLeft},
			},
			Nonterminals: []frontend.Symbol{
				{Name: "E"},
			},
			Productions: []frontend.Production{
				{
					NonterminalIdx:        0,
					PrecedenceTerminalIdx: &precedenceTerminalIdx,
				},
			},
			StartNonterminalIdx: 0,
		}

		augmentedGrammar := frontend.AugmentGrammar(grammar)

		augmentedProduction := augmentedGrammar.Productions[1]
		Expect(augmentedProduction.PrecedenceTerminalIdx).ToNot(BeNil())
		Expect(augmentedGrammar.Terminals[*augmentedProduction.PrecedenceTerminalIdx].Name).To(Equal("+"))
	})
})
