package oracle_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/core/ielr1/golr/oracle"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

var _ = Describe("LALR(1) Parser Differ", func() {
	// The grammar names two terminals so that lookahead sets render with readable names in the reported differences.
	grammar := frontend.Grammar{
		Terminals: []frontend.Symbol{{Name: "a"}, {Name: "b"}},
	}

	// startState is state 0 in both parser tables. Its kernel items differ from the state which carries the difference,
	// so a report which names the wrong state (the start state) instead of the right one is easy to catch.
	startState := backend.State{
		KernelItems: backend.NewCoreSet(backend.NewCore(0, 0)),
	}

	// diffStateKernelItems belong to the state which carries the difference in every case below. It is not the start
	// state, so the reported difference must name these kernel items and not the start state's.
	diffStateKernelItems := backend.NewCoreSet(backend.NewCore(1, 2))

	It("names the state which holds more than one reduce action for a production", func() {
		// The want state holds two reduce actions for the same production, which a real LALR(1) merge would have folded
		// into a single reduce action with the union of both lookahead sets. This is the case which used to be reported
		// against the start state instead of the state which actually holds the duplicate.
		want := backend.Parser{
			Grammar: grammar,
			States: []backend.State{
				startState,
				{
					KernelItems: diffStateKernelItems,
					ReduceActions: backend.NewReduceActionSet(
						backend.NewReduceAction(backend.NewLookaheadSet(0), 1),
						backend.NewReduceAction(backend.NewLookaheadSet(1), 1),
					),
				},
			},
		}
		got := backend.Parser{
			Grammar: grammar,
			States: []backend.State{
				startState,
				{
					KernelItems: diffStateKernelItems,
					ReduceActions: backend.NewReduceActionSet(
						backend.NewReduceAction(backend.NewLookaheadSet(0), 1),
					),
				},
			},
		}

		Expect(oracle.DiffLALR1ParserStates(want, got)).To(ConsistOf(
			"want: state {(production 1, position 2)}: production 1 has more than one reduce action",
		))
	})

	It("names the state whose reduce action lookahead set differs", func() {
		want := backend.Parser{
			Grammar: grammar,
			States: []backend.State{
				startState,
				{
					KernelItems: diffStateKernelItems,
					ReduceActions: backend.NewReduceActionSet(
						backend.NewReduceAction(backend.NewLookaheadSet(0), 1),
					),
				},
			},
		}
		got := backend.Parser{
			Grammar: grammar,
			States: []backend.State{
				startState,
				{
					KernelItems: diffStateKernelItems,
					ReduceActions: backend.NewReduceActionSet(
						backend.NewReduceAction(backend.NewLookaheadSet(0, 1), 1),
					),
				},
			},
		}

		Expect(oracle.DiffLALR1ParserStates(want, got)).To(ConsistOf(
			"state {(production 1, position 2)}: reduce action for production 1 on {a, b}, want {a}",
		))
	})
})
