package frontend_test

import (
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

var _ = Describe("FirstSets", func() {
	// The grammar is
	//
	//	S -> A B "c"
	//	A -> "a" | ε
	//	B -> A A
	//
	// so A and B are nullable, while S is not, and FIRST(S) reaches through both nullable nonterminals to "c".
	const (
		terminalA = iota
		terminalC
	)
	const (
		nonterminalS = iota
		nonterminalA
		nonterminalB
	)
	grammar := frontend.Grammar{
		Terminals:    []frontend.Symbol{{Name: "a"}, {Name: "c"}},
		Nonterminals: []frontend.Symbol{{Name: "S"}, {Name: "A"}, {Name: "B"}},
		Productions: []frontend.Production{
			{NonterminalIdx: nonterminalS, SymbolRefs: []frontend.SymbolRef{
				frontend.NewNonterminalRef(nonterminalA),
				frontend.NewNonterminalRef(nonterminalB),
				frontend.NewTerminalRef(terminalC),
			}},
			{NonterminalIdx: nonterminalA, SymbolRefs: []frontend.SymbolRef{frontend.NewTerminalRef(terminalA)}},
			{NonterminalIdx: nonterminalA},
			{NonterminalIdx: nonterminalB, SymbolRefs: []frontend.SymbolRef{
				frontend.NewNonterminalRef(nonterminalA),
				frontend.NewNonterminalRef(nonterminalA),
			}},
		},
		StartNonterminalIdx: nonterminalS,
	}

	It("should find the nullable nonterminals", func() {
		firstSets := frontend.NewFirstSets(grammar)

		Expect(firstSets.IsNullable(nonterminalS)).To(BeFalse())
		Expect(firstSets.IsNullable(nonterminalA)).To(BeTrue())
		Expect(firstSets.IsNullable(nonterminalB)).To(BeTrue(), "B is nullable only through A")
	})

	It("should tell if a sequence is nullable", func() {
		firstSets := frontend.NewFirstSets(grammar)

		Expect(firstSets.IsSequenceNullable(nil)).To(BeTrue())
		Expect(firstSets.IsSequenceNullable(grammar.Productions[3].SymbolRefs)).To(BeTrue())
		Expect(firstSets.IsSequenceNullable(grammar.Productions[0].SymbolRefs)).To(BeFalse())
	})

	It("should reach through nullable nonterminals for the first set of a sequence", func() {
		firstSets := frontend.NewFirstSets(grammar)

		var set utils.Bitset
		Expect(firstSets.FirstOfSequence(grammar.Productions[0].SymbolRefs, &set)).To(BeFalse())
		Expect(slices.Collect(set.All())).To(Equal([]int{terminalA, terminalC}))
	})

	It("should report a nullable sequence and keep what the set held", func() {
		firstSets := frontend.NewFirstSets(grammar)

		var set utils.Bitset
		set.Add(terminalC)
		Expect(firstSets.FirstOfSequence(grammar.Productions[3].SymbolRefs, &set)).To(BeTrue())
		Expect(slices.Collect(set.All())).To(Equal([]int{terminalA, terminalC}))
	})
})
