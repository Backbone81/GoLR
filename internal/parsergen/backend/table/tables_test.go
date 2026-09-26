package table_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/backend/table"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

// symbolName names the token constant of a terminal after the terminal itself.
func symbolName(symbol frontend.Symbol) string {
	return symbol.Name
}

var _ = Describe("Tables", func() {
	options := table.TablesOptions{
		NewIntArray:  utils.NewIntArray,
		TerminalName: symbolName,
	}

	It("lists the token of every terminal by its column", func() {
		parser := backend.Parser{
			Grammar: handBuiltGrammar(3, 1),
			States:  []backend.State{{}},
		}

		tables := table.NewTables(parser, options)

		Expect(tables.TokenByTerminalColumn).To(HaveLen(4))
		Expect(tables.TokenByTerminalColumn[table.NoTerminalColumn]).To(BeEmpty())
		for terminalIdx, terminal := range parser.Grammar.Terminals {
			Expect(tables.TokenByTerminalColumn[table.TerminalColumn(terminalIdx)]).To(Equal(terminal.Name))
		}
		Expect(tables.TokenByTerminalColumn[tables.ErrorTerminalColumn]).To(Equal(frontend.SymbolError.Name))
	})

	It("marks the consistent states with one byte each", func() {
		parser := backend.Parser{
			Grammar: handBuiltGrammar(3, 1),
			States: []backend.State{
				{DefaultReduceProductionIdx: ptr(2)},
				{
					TransitionActions: backend.NewTransitionActionSet(
						backend.NewTransitionAction(frontend.NewTerminalRef(0), 0),
					),
					DefaultReduceProductionIdx: ptr(2),
				},
			},
		}

		tables := table.NewTables(parser, options)

		Expect(tables.ConsistentByState.Values).To(Equal([]int{1, 0}))
		Expect(tables.ConsistentByState.Type).To(Equal("uint8"))
	})
})
