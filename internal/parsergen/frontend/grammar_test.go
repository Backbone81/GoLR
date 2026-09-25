package frontend_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/frontend"
)

var _ = Describe("Grammar", func() {
	Context("AugmentGrammar", func() {
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

		It("should carry the explicit name of a production across augmentation", func() {
			name := "the_answer"
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
						Name: &name,
					},
				},
				StartNonterminalIdx: 0,
			}

			augmentedGrammar := frontend.AugmentGrammar(grammar)

			Expect(augmentedGrammar.Productions[1].Name).ToNot(BeNil())
			Expect(*augmentedGrammar.Productions[1].Name).To(Equal("the_answer"))

			// The production of the new start symbol never carries a name.
			Expect(augmentedGrammar.Productions[0].Name).To(BeNil())
		})
	})

	Context("Validate", func() {
		It("should reject two productions with the same explicit name", func() {
			first := "dup"
			second := "dup"
			grammar := frontend.Grammar{
				Terminals: []frontend.Symbol{
					{Name: "a"},
					{Name: "b"},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "E"},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs:     []frontend.SymbolRef{frontend.NewTerminalRef(0)},
						Name:           &first,
					},
					{
						NonterminalIdx: 0,
						SymbolRefs:     []frontend.SymbolRef{frontend.NewTerminalRef(1)},
						Name:           &second,
					},
				},
				StartNonterminalIdx: 0,
			}

			warnings, err := grammar.Validate()

			Expect(warnings).To(BeEmpty())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`"dup"`))
			// The message spells both colliding productions out so the author can find them.
			Expect(err.Error()).To(ContainSubstring("E -> a"))
			Expect(err.Error()).To(ContainSubstring("E -> b"))
		})

		It("should accept distinct explicit names and unnamed productions", func() {
			named := "only_one"
			grammar := frontend.Grammar{
				Terminals: []frontend.Symbol{
					{Name: "a"},
					{Name: "b"},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "E"},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs:     []frontend.SymbolRef{frontend.NewTerminalRef(0)},
						Name:           &named,
					},
					{
						NonterminalIdx: 0,
						SymbolRefs:     []frontend.SymbolRef{frontend.NewTerminalRef(1)},
					},
				},
				StartNonterminalIdx: 0,
			}

			Expect(grammar.Validate()).Error().To(Succeed())
		})
	})

	Context("Validate usefulness", func() {
		// s: x | s A x; x: B;
		It("should accept a grammar where every nonterminal is productive and reachable", func() {
			grammar := frontend.Grammar{
				Terminals: []frontend.Symbol{
					{Name: "A"},
					{Name: "B"},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
					{Name: "x"},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs:     []frontend.SymbolRef{frontend.NewNonterminalRef(1)},
					},
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewNonterminalRef(0),
							frontend.NewTerminalRef(0),
							frontend.NewNonterminalRef(1),
						},
					},
					{
						NonterminalIdx: 1,
						SymbolRefs:     []frontend.SymbolRef{frontend.NewTerminalRef(1)},
					},
				},
				StartNonterminalIdx: 0,
			}

			warnings, err := grammar.Validate()

			Expect(err).ToNot(HaveOccurred())
			Expect(warnings).To(BeEmpty())
		})

		DescribeTable("should reject unproductive nonterminals",
			func(grammar frontend.Grammar, expectedMessage string) {
				warnings, err := grammar.Validate()

				Expect(err).To(MatchError(expectedMessage))
				Expect(warnings).To(BeEmpty())
			},
			// s: A | x; x: x B;
			Entry("left recursion without a base case",
				frontend.Grammar{
					Terminals: []frontend.Symbol{
						{Name: "A"},
						{Name: "B"},
					},
					Nonterminals: []frontend.Symbol{
						{Name: "s"},
						{Name: "x"},
					},
					Productions: []frontend.Production{
						{
							NonterminalIdx: 0,
							SymbolRefs:     []frontend.SymbolRef{frontend.NewTerminalRef(0)},
						},
						{
							NonterminalIdx: 0,
							SymbolRefs:     []frontend.SymbolRef{frontend.NewNonterminalRef(1)},
						},
						{
							NonterminalIdx: 1,
							SymbolRefs: []frontend.SymbolRef{
								frontend.NewNonterminalRef(1),
								frontend.NewTerminalRef(1),
							},
						},
					},
					StartNonterminalIdx: 0,
				},
				`nonterminal "x" does not derive any finite string`,
			),
			// s: A | x; x: B x;
			Entry("right recursion without a base case",
				frontend.Grammar{
					Terminals: []frontend.Symbol{
						{Name: "A"},
						{Name: "B"},
					},
					Nonterminals: []frontend.Symbol{
						{Name: "s"},
						{Name: "x"},
					},
					Productions: []frontend.Production{
						{
							NonterminalIdx: 0,
							SymbolRefs:     []frontend.SymbolRef{frontend.NewTerminalRef(0)},
						},
						{
							NonterminalIdx: 0,
							SymbolRefs:     []frontend.SymbolRef{frontend.NewNonterminalRef(1)},
						},
						{
							NonterminalIdx: 1,
							SymbolRefs: []frontend.SymbolRef{
								frontend.NewTerminalRef(1),
								frontend.NewNonterminalRef(1),
							},
						},
					},
					StartNonterminalIdx: 0,
				},
				`nonterminal "x" does not derive any finite string`,
			),
			// s: A | x; x: A y; y: x B;
			Entry("mutual recursion without a base case",
				frontend.Grammar{
					Terminals: []frontend.Symbol{
						{Name: "A"},
						{Name: "B"},
					},
					Nonterminals: []frontend.Symbol{
						{Name: "s"},
						{Name: "x"},
						{Name: "y"},
					},
					Productions: []frontend.Production{
						{
							NonterminalIdx: 0,
							SymbolRefs:     []frontend.SymbolRef{frontend.NewTerminalRef(0)},
						},
						{
							NonterminalIdx: 0,
							SymbolRefs:     []frontend.SymbolRef{frontend.NewNonterminalRef(1)},
						},
						{
							NonterminalIdx: 1,
							SymbolRefs: []frontend.SymbolRef{
								frontend.NewTerminalRef(0),
								frontend.NewNonterminalRef(2),
							},
						},
						{
							NonterminalIdx: 2,
							SymbolRefs: []frontend.SymbolRef{
								frontend.NewNonterminalRef(1),
								frontend.NewTerminalRef(1),
							},
						},
					},
					StartNonterminalIdx: 0,
				},
				"nonterminal \"x\" does not derive any finite string\n"+
					`nonterminal "y" does not derive any finite string`,
			),
			// s: A | undef;
			Entry("a nonterminal without productions",
				frontend.Grammar{
					Terminals: []frontend.Symbol{
						{Name: "A"},
					},
					Nonterminals: []frontend.Symbol{
						{Name: "s"},
						{Name: "undef"},
					},
					Productions: []frontend.Production{
						{
							NonterminalIdx: 0,
							SymbolRefs:     []frontend.SymbolRef{frontend.NewTerminalRef(0)},
						},
						{
							NonterminalIdx: 0,
							SymbolRefs:     []frontend.SymbolRef{frontend.NewNonterminalRef(1)},
						},
					},
					StartNonterminalIdx: 0,
				},
				`nonterminal "undef" has no productions`,
			),
			// s: s A;
			Entry("an unproductive start nonterminal",
				frontend.Grammar{
					Terminals: []frontend.Symbol{
						{Name: "A"},
					},
					Nonterminals: []frontend.Symbol{
						{Name: "s"},
					},
					Productions: []frontend.Production{
						{
							NonterminalIdx: 0,
							SymbolRefs: []frontend.SymbolRef{
								frontend.NewNonterminalRef(0),
								frontend.NewTerminalRef(0),
							},
						},
					},
					StartNonterminalIdx: 0,
				},
				`start nonterminal "s" does not derive any finite string`,
			),
		)

		// s: x A; x: <empty> | x A;
		It("should accept a nonterminal which is productive through an empty right hand side", func() {
			grammar := frontend.Grammar{
				Terminals: []frontend.Symbol{
					{Name: "A"},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
					{Name: "x"},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewNonterminalRef(1),
							frontend.NewTerminalRef(0),
						},
					},
					{
						NonterminalIdx: 1,
					},
					{
						NonterminalIdx: 1,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewNonterminalRef(1),
							frontend.NewTerminalRef(0),
						},
					},
				},
				StartNonterminalIdx: 0,
			}

			warnings, err := grammar.Validate()

			Expect(err).ToNot(HaveOccurred())
			Expect(warnings).To(BeEmpty())
		})

		// s: A | x; x: $error;
		It("should accept a nonterminal which is productive through the error symbol", func() {
			grammar := frontend.Grammar{
				Terminals: []frontend.Symbol{
					{Name: "A"},
					frontend.SymbolError,
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
					{Name: "x"},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs:     []frontend.SymbolRef{frontend.NewTerminalRef(0)},
					},
					{
						NonterminalIdx: 0,
						SymbolRefs:     []frontend.SymbolRef{frontend.NewNonterminalRef(1)},
					},
					{
						NonterminalIdx: 1,
						SymbolRefs:     []frontend.SymbolRef{frontend.NewTerminalRef(1)},
					},
				},
				StartNonterminalIdx: 0,
			}

			warnings, err := grammar.Validate()

			Expect(err).ToNot(HaveOccurred())
			Expect(warnings).To(BeEmpty())
		})

		// s: A; orphan: B;
		It("should warn about an unreachable nonterminal", func() {
			grammar := frontend.Grammar{
				Terminals: []frontend.Symbol{
					{Name: "A"},
					{Name: "B"},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
					{Name: "orphan"},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs:     []frontend.SymbolRef{frontend.NewTerminalRef(0)},
					},
					{
						NonterminalIdx: 1,
						SymbolRefs:     []frontend.SymbolRef{frontend.NewTerminalRef(1)},
					},
				},
				StartNonterminalIdx: 0,
			}

			warnings, err := grammar.Validate()

			Expect(err).ToNot(HaveOccurred())
			Expect(warnings).To(HaveExactElements(
				MatchError(`nonterminal "orphan" is unreachable from the start nonterminal "s"`),
			))
		})

		// s: A; x: y B | A; y: x A | B;
		It("should warn about every nonterminal of an unreachable cluster", func() {
			grammar := frontend.Grammar{
				Terminals: []frontend.Symbol{
					{Name: "A"},
					{Name: "B"},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
					{Name: "x"},
					{Name: "y"},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs:     []frontend.SymbolRef{frontend.NewTerminalRef(0)},
					},
					{
						NonterminalIdx: 1,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewNonterminalRef(2),
							frontend.NewTerminalRef(1),
						},
					},
					{
						NonterminalIdx: 1,
						SymbolRefs:     []frontend.SymbolRef{frontend.NewTerminalRef(0)},
					},
					{
						NonterminalIdx: 2,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewNonterminalRef(1),
							frontend.NewTerminalRef(0),
						},
					},
					{
						NonterminalIdx: 2,
						SymbolRefs:     []frontend.SymbolRef{frontend.NewTerminalRef(1)},
					},
				},
				StartNonterminalIdx: 0,
			}

			warnings, err := grammar.Validate()

			Expect(err).ToNot(HaveOccurred())
			Expect(warnings).To(HaveExactElements(
				MatchError(`nonterminal "x" is unreachable from the start nonterminal "s"`),
				MatchError(`nonterminal "y" is unreachable from the start nonterminal "s"`),
			))
		})

		// s: A | loop; loop: loop B; orphan: B;
		It("should return the warnings together with the error", func() {
			grammar := frontend.Grammar{
				Terminals: []frontend.Symbol{
					{Name: "A"},
					{Name: "B"},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
					{Name: "loop"},
					{Name: "orphan"},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs:     []frontend.SymbolRef{frontend.NewTerminalRef(0)},
					},
					{
						NonterminalIdx: 0,
						SymbolRefs:     []frontend.SymbolRef{frontend.NewNonterminalRef(1)},
					},
					{
						NonterminalIdx: 1,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewNonterminalRef(1),
							frontend.NewTerminalRef(1),
						},
					},
					{
						NonterminalIdx: 2,
						SymbolRefs:     []frontend.SymbolRef{frontend.NewTerminalRef(1)},
					},
				},
				StartNonterminalIdx: 0,
			}

			warnings, err := grammar.Validate()

			Expect(err).To(MatchError(`nonterminal "loop" does not derive any finite string`))
			Expect(warnings).To(HaveExactElements(
				MatchError(`nonterminal "orphan" is unreachable from the start nonterminal "s"`),
			))
		})
	})
})
