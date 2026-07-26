package frontend_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/frontend"
)

var _ = Describe("ErrorTerminalRef", func() {
	It("returns the reference for the error symbol at whatever index it sits", func() {
		grammar := frontend.Grammar{
			Terminals: []frontend.Symbol{
				frontend.SymbolEOF,
				{Name: "SEMICOLON"},
				frontend.SymbolError,
			},
		}

		errorRef, ok := frontend.ErrorTerminalRef(grammar)

		Expect(ok).To(BeTrue())
		Expect(errorRef).To(Equal(frontend.NewTerminalRef(2)))
	})

	It("reports that a grammar which does not reference the error symbol does not carry it", func() {
		grammar := frontend.Grammar{
			Terminals: []frontend.Symbol{
				frontend.SymbolEOF,
				{Name: "SEMICOLON"},
			},
		}

		_, ok := frontend.ErrorTerminalRef(grammar)

		Expect(ok).To(BeFalse())
	})

	It("does not mistake a nonterminal of the same name for the error symbol", func() {
		grammar := frontend.Grammar{
			Terminals: []frontend.Symbol{
				frontend.SymbolEOF,
			},
			Nonterminals: []frontend.Symbol{
				frontend.SymbolError,
			},
		}

		_, ok := frontend.ErrorTerminalRef(grammar)

		Expect(ok).To(BeFalse())
	})
})
