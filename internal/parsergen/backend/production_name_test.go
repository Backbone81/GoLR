package backend_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// productionNamePtr returns a pointer to s, for the Production.Name field.
func productionNamePtr(s string) *string {
	return &s
}

// productionNameGrammar builds an augmented grammar whose production 0 is $accept and whose remaining productions have
// the given left hand side nonterminal index and optional explicit name. Right hand sides do not matter for naming.
func productionNameGrammar(productions ...frontend.Production) frontend.Grammar {
	return frontend.Grammar{
		Nonterminals: []frontend.Symbol{
			{Name: "$accept"},
			{Name: "S"},
			{Name: "A"},
		},
		Productions: append([]frontend.Production{{NonterminalIdx: 0}}, productions...),
	}
}

var _ = Describe("ProductionNames", func() {
	It("leaves the $accept slot empty", func() {
		names := backend.ProductionNames(productionNameGrammar(
			frontend.Production{NonterminalIdx: 1},
		))

		Expect(names[0]).To(Equal(""))
	})

	It("numbers the alternatives of a left hand side 1-based, single-alternative rules included", func() {
		names := backend.ProductionNames(productionNameGrammar(
			frontend.Production{NonterminalIdx: 1},
			frontend.Production{NonterminalIdx: 1},
			frontend.Production{NonterminalIdx: 2},
		))

		Expect(names).To(Equal([]string{"", "S_1", "S_2", "A_1"}))
	})

	It("keeps an explicit name but still counts its position", func() {
		names := backend.ProductionNames(productionNameGrammar(
			frontend.Production{NonterminalIdx: 1},
			frontend.Production{NonterminalIdx: 1, Name: productionNamePtr("special")},
			frontend.Production{NonterminalIdx: 1},
		))

		Expect(names).To(Equal([]string{"", "S_1", "special", "S_3"}))
	})

	It("suffixes an automatic name which clashes with an explicit one until it is free", func() {
		names := backend.ProductionNames(productionNameGrammar(
			frontend.Production{NonterminalIdx: 1},
			frontend.Production{NonterminalIdx: 1, Name: productionNamePtr("S_1")},
		))

		Expect(names).To(Equal([]string{"", "S_1_1", "S_1"}))
	})

	It("walks the suffix past several explicit clashes", func() {
		names := backend.ProductionNames(productionNameGrammar(
			frontend.Production{NonterminalIdx: 1},
			frontend.Production{NonterminalIdx: 1, Name: productionNamePtr("S_1")},
			frontend.Production{NonterminalIdx: 1, Name: productionNamePtr("S_1_1")},
		))

		Expect(names).To(Equal([]string{"", "S_1_2", "S_1", "S_1_1"}))
	})

	It("is deterministic across calls", func() {
		grammar := productionNameGrammar(
			frontend.Production{NonterminalIdx: 1},
			frontend.Production{NonterminalIdx: 1, Name: productionNamePtr("S_1")},
			frontend.Production{NonterminalIdx: 2},
		)

		Expect(backend.ProductionNames(grammar)).To(Equal(backend.ProductionNames(grammar)))
	})
})
