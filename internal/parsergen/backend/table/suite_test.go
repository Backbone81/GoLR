package table_test

import (
	"bytes"
	"slices"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/backend/table"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	"github.com/backbone81/golr/internal/parsergen/core"
	ielr1golrcore "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	bisonfrontend "github.com/backbone81/golr/internal/parsergen/frontend/bison"
	"github.com/backbone81/golr/testdata"
)

func TestTable(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Table Suite")
}

// wellKnownParser builds the parser tables of one of the well-known grammars the same way the parser generator does.
// Those grammars are the real-world tables the compression is checked against, next to the small ones built by hand.
func wellKnownParser(wellKnownGrammar testdata.WellKnownGrammar, options ...core.Option) backend.Parser {
	grammar, err := bisonfrontend.ToGrammar(bytes.NewBuffer(wellKnownGrammar.Content), wellKnownGrammar.FileName)
	Expect(err).ToNot(HaveOccurred())

	parser, _, err := ielr1golrcore.GrammarToParser(grammar, conflict.DefaultPolicy, options...)
	Expect(err).ToNot(HaveOccurred())
	Expect(parser.States).ToNot(BeEmpty())

	return parser
}

// wellKnownGrammarByTitle returns the well-known grammar with the given title, for the specs which need one particular
// grammar instead of all of them.
func wellKnownGrammarByTitle(title string) testdata.WellKnownGrammar {
	idx := slices.IndexFunc(testdata.WellKnownGrammars, func(wellKnownGrammar testdata.WellKnownGrammar) bool {
		return wellKnownGrammar.Title == title
	})
	Expect(idx).ToNot(Equal(-1), "no well-known grammar titled %q", title)

	return testdata.WellKnownGrammars[idx]
}

// referenceAction returns what the parser does in the given state when the scanner delivers the given terminal, by
// reading the state directly. It is the reference the compressed lookup is compared against, and is deliberately
// written independently of the code under test.
//
// The actions the state has of its own come first, whether they shift, reduce or reject the terminal, and only a
// terminal without one of those reaches the default action of the state.
func referenceAction(state backend.State, terminalIdx int, errorTerminalIdx int) table.Action {
	terminalRef := frontend.NewTerminalRef(terminalIdx)
	for _, transitionAction := range state.TransitionActions.All() {
		if transitionAction.SymbolRef() == terminalRef {
			return table.NewShiftAction(transitionAction.StateIdx())
		}
	}

	for _, reduceAction := range state.ReduceActions.All() {
		if reduceAction.LookaheadSet.Contains(terminalIdx) {
			return table.NewReduceAction(reduceAction.ProductionIdx)
		}
	}

	// The error symbol is the one terminal a rejection cannot apply to, because no scanner delivers it, see
	// fillActionRow.
	if terminalIdx != errorTerminalIdx && state.RejectedTerminals.Contains(terminalIdx) {
		return table.NewErrorAction()
	}

	return referenceDefaultAction(state)
}

// referenceDefaultAction returns the action the given state takes for a terminal it has no action of its own for, by
// reading the state directly.
func referenceDefaultAction(state backend.State) table.Action {
	if state.DefaultReduceProductionIdx != nil {
		return table.NewReduceAction(*state.DefaultReduceProductionIdx)
	}

	// The native GoLR cores keep the accept a reduce of the accept production with an empty reduction lookahead set,
	// which makes it the unconditional action of its state.
	for _, reduceAction := range state.ReduceActions.All() {
		if reduceAction.ProductionIdx == 0 {
			return table.NewAcceptAction()
		}
	}

	return table.NoAction
}

// referenceGoto returns the state the parser continues in when it reduced to the given nonterminal and uncovered the
// given state, by reading the state directly.
func referenceGoto(state backend.State, nonterminalIdx int) int {
	nonterminalRef := frontend.NewNonterminalRef(nonterminalIdx)
	for _, transitionAction := range state.TransitionActions.All() {
		if transitionAction.SymbolRef() == nonterminalRef {
			return transitionAction.StateIdx()
		}
	}
	return table.NoGoto
}
