package golr_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/parsergen/frontend/golr"
)

var _ = Describe("GoLR Grammar Files", func() {
	It("should correctly parse the most minimal grammar", func() {
		source := `
			@scanner {
			}
			@parser {
				file: @empty;
			}
		`
		grammar, err := golr.GrammarFromString(source)
		Expect(err).ToNot(HaveOccurred())
		Expect(grammar).To(Equal(frontend.Grammar{
			Terminals: nil,
			Nonterminals: []frontend.Symbol{
				{
					Name: "file",
				},
			},
			Productions: []frontend.Production{
				{
					NonterminalIdx: 0,
					SymbolRefs:     nil,
				},
			},
			StartNonterminalIdx: 0,
		}))
	})
})
