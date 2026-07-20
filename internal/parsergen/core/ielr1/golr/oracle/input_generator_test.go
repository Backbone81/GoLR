package oracle_test

import (
	"math/rand"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ielr1golrcore "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	"github.com/backbone81/golr/internal/parsergen/core/ielr1/golr/oracle"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// listGrammar derives `b a*`: S -> S a | b. It is unambiguous and LR(0), so the resolved parser accepts exactly the
// language and every generated sentence must be accepted.
var listGrammar = frontend.Grammar{
	Terminals:    []frontend.Symbol{{Name: "a"}, {Name: "b"}},
	Nonterminals: []frontend.Symbol{{Name: "S"}},
	Productions: []frontend.Production{
		{NonterminalIdx: 0, SymbolRefs: []frontend.SymbolRef{frontend.NewNonterminalRef(0), frontend.NewTerminalRef(0)}},
		{NonterminalIdx: 0, SymbolRefs: []frontend.SymbolRef{frontend.NewTerminalRef(1)}},
	},
	StartNonterminalIdx: 0,
}

// parensGrammar derives balanced parentheses `(^n )^n`: S -> ( S ) | ε. It is LR(1) and conflict free, its start symbol
// is nullable, and it is recursive, so it exercises both empty derivations and the termination fallback.
var parensGrammar = frontend.Grammar{
	Terminals:    []frontend.Symbol{{Name: "("}, {Name: ")"}},
	Nonterminals: []frontend.Symbol{{Name: "S"}},
	Productions: []frontend.Production{
		{NonterminalIdx: 0, SymbolRefs: []frontend.SymbolRef{frontend.NewTerminalRef(0), frontend.NewNonterminalRef(0), frontend.NewTerminalRef(1)}},
		{NonterminalIdx: 0, SymbolRefs: nil},
	},
	StartNonterminalIdx: 0,
}

// nullablePairGrammar derives the four sentences {ε, a, b, ab}: S -> A B, A -> a | ε, B -> b | ε. It is LALR(1) and
// conflict free and every nonterminal has a nullable alternative, so it exercises nullable expansions.
var nullablePairGrammar = frontend.Grammar{
	Terminals:    []frontend.Symbol{{Name: "a"}, {Name: "b"}},
	Nonterminals: []frontend.Symbol{{Name: "S"}, {Name: "A"}, {Name: "B"}},
	Productions: []frontend.Production{
		{NonterminalIdx: 0, SymbolRefs: []frontend.SymbolRef{frontend.NewNonterminalRef(1), frontend.NewNonterminalRef(2)}},
		{NonterminalIdx: 1, SymbolRefs: []frontend.SymbolRef{frontend.NewTerminalRef(0)}},
		{NonterminalIdx: 1, SymbolRefs: nil},
		{NonterminalIdx: 2, SymbolRefs: []frontend.SymbolRef{frontend.NewTerminalRef(1)}},
		{NonterminalIdx: 2, SymbolRefs: nil},
	},
	StartNonterminalIdx: 0,
}

// acceptsSentence reports whether the resolved parser of the un-augmented grammar accepts the sentence. The sentence is
// already in the augmented alphabet the generator and the parser table share, so it is interpreted as is. This is only a
// valid membership oracle for conflict-free grammars, where the resolved table accepts exactly the language; hence the
// assertion that the grammar resolves without conflicts.
func acceptsSentence(grammar frontend.Grammar, sentence []int) bool {
	parser, conflicts, err := ielr1golrcore.GrammarToParser(grammar)
	Expect(err).NotTo(HaveOccurred())
	Expect(conflicts).To(BeEmpty())

	interpreter := oracle.NewParserInterpreter(parser, sentence)
	for interpreter.Next() {
	}
	return interpreter.Value().Kind == oracle.ParserActionAccept
}

var _ = Describe("Input Generator", func() {
	Describe("reproducibility", func() {
		It("produces the same stream of sentences for the same seed", func() {
			augmented := frontend.AugmentGrammar(parensGrammar)
			first := oracle.NewInputGenerator(augmented, rand.New(rand.NewSource(1)))
			second := oracle.NewInputGenerator(augmented, rand.New(rand.NewSource(1)))
			for range 50 {
				Expect(first.Generate()).To(Equal(second.Generate()))
			}
		})

		It("produces different sentences across a run for a recursive grammar", func() {
			generator := oracle.NewInputGenerator(frontend.AugmentGrammar(parensGrammar), rand.New(rand.NewSource(1)))
			seen := map[string][]int{}
			for range 50 {
				sentence := generator.Generate()
				seen[intsKey(sentence)] = sentence
			}
			// A recursive grammar with a length budget yields a spread of sentences, not one repeated sentence.
			Expect(len(seen)).To(BeNumerically(">", 1))
		})
	})

	Describe("generated sentences are in the language", func() {
		DescribeTable("every generated sentence is accepted by the resolved parser",
			func(grammar frontend.Grammar) {
				generator := oracle.NewInputGenerator(frontend.AugmentGrammar(grammar), rand.New(rand.NewSource(20)))
				for range 200 {
					sentence := generator.Generate()
					// The augmented alphabet reserves index 0 for EOF, which the generator never emits, and places the
					// grammar's own terminals at indexes 1..len(Terminals).
					for _, terminalIdx := range sentence {
						Expect(terminalIdx).To(And(BeNumerically(">=", 1), BeNumerically("<=", len(grammar.Terminals))))
					}
					Expect(acceptsSentence(grammar, sentence)).To(BeTrue(), "sentence %v was rejected", sentence)
				}
			},
			Entry("list grammar b a*", listGrammar),
			Entry("balanced parentheses", parensGrammar),
			Entry("nullable pair", nullablePairGrammar),
		)

		It("derives the empty sentence when Start is nullable", func() {
			generator := oracle.NewInputGenerator(frontend.AugmentGrammar(parensGrammar), rand.New(rand.NewSource(7)))
			sawEmpty := false
			for range 200 {
				if len(generator.Generate()) == 0 {
					sawEmpty = true
					break
				}
			}
			Expect(sawEmpty).To(BeTrue())
		})
	})

	Describe("termination fallback", func() {
		It("uses only shortest productions once the expansion budget is zero", func() {
			// With no expansions to spend, every nonterminal is expanded by a shortest-terminating production. For the
			// list grammar that is S -> b, so the derivation always collapses to the single terminal b, which augmentation
			// (EOF prepended at index 0) places at index 2.
			generator := oracle.NewInputGenerator(frontend.AugmentGrammar(listGrammar), rand.New(rand.NewSource(3)))
			generator.MaxExpansions = 0
			for range 20 {
				Expect(generator.Generate()).To(Equal([]int{2}))
			}
		})

		It("keeps sentences shorter on a smaller expansion budget", func() {
			augmented := frontend.AugmentGrammar(parensGrammar)
			longGenerator := oracle.NewInputGenerator(augmented, rand.New(rand.NewSource(5)))
			longGenerator.MaxExpansions = 40
			shortGenerator := oracle.NewInputGenerator(augmented, rand.New(rand.NewSource(5)))
			shortGenerator.MaxExpansions = 4

			longTotal, shortTotal := 0, 0
			for range 200 {
				longTotal += len(longGenerator.Generate())
				shortTotal += len(shortGenerator.Generate())
			}
			Expect(shortTotal).To(BeNumerically("<", longTotal))
		})
	})

	Describe("robustness on the random grammar corpus", func() {
		It("terminates and stays in range for every generated grammar", func() {
			grammarGenerator := oracle.DefaultGrammarGenerator(rand.New(rand.NewSource(42)))
			for range 100 {
				grammar := grammarGenerator.Generate()
				inputGenerator := oracle.NewInputGenerator(frontend.AugmentGrammar(grammar), rand.New(rand.NewSource(99)))
				for range 20 {
					sentence := inputGenerator.Generate()
					// Never the EOF terminal at index 0, always one of the grammar's own terminals at 1..len(Terminals).
					for _, terminalIdx := range sentence {
						Expect(terminalIdx).To(And(BeNumerically(">=", 1), BeNumerically("<=", len(grammar.Terminals))))
					}
					// A finite budget with a shortest-production fallback keeps sentences bounded; a runaway would blow
					// far past this before the test noticed.
					Expect(len(sentence)).To(BeNumerically("<", 100000))
				}
			}
		})
	})
})

// intsKey turns a sentence into a map key so distinct sentences can be counted.
func intsKey(sentence []int) string {
	key := make([]byte, 0, len(sentence)*2)
	for _, terminalIdx := range sentence {
		key = append(key, byte(terminalIdx), ',')
	}
	return string(key)
}
