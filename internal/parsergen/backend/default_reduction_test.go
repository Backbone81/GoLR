package backend_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

// ptr returns a pointer to a copy of the value, for building the expected DefaultReduceProductionIdx.
func ptr[T any](value T) *T {
	return &value
}

var _ = Describe("ApplyDefaultReductions", func() {
	It("picks the reduce action with the widest lookahead set as the default", func() {
		state := backend.State{
			ReduceActions: backend.NewReduceActionSet(
				// production 3 covers two terminals, production 5 covers three.
				backend.NewReduceAction(utils.NewBitset(0, 1), 3),
				backend.NewReduceAction(utils.NewBitset(2, 4, 6), 5),
			),
		}
		parser := backend.Parser{States: []backend.State{state}}

		backend.ApplyDefaultReductions(&parser)

		// The wider production 5 becomes the default and is removed from the explicit reduce actions, leaving only the
		// narrower production 3 behind.
		Expect(parser.States[0].DefaultReduceProductionIdx).To(Equal(ptr(5)))
		Expect(parser.States[0].ReduceActions).To(Equal(backend.NewReduceActionSet(
			backend.NewReduceAction(utils.NewBitset(0, 1), 3),
		)))
	})

	It("breaks ties on lookahead width by the lowest production index", func() {
		state := backend.State{
			ReduceActions: backend.NewReduceActionSet(
				backend.NewReduceAction(utils.NewBitset(0, 1), 2),
				backend.NewReduceAction(utils.NewBitset(2, 3), 4),
			),
		}
		parser := backend.Parser{States: []backend.State{state}}

		backend.ApplyDefaultReductions(&parser)

		Expect(parser.States[0].DefaultReduceProductionIdx).To(Equal(ptr(2)))
		Expect(parser.States[0].ReduceActions).To(Equal(backend.NewReduceActionSet(
			backend.NewReduceAction(utils.NewBitset(2, 3), 4),
		)))
	})

	It("promotes the sole reduce action of a state, leaving no explicit reduce action behind", func() {
		state := backend.State{
			ReduceActions: backend.NewReduceActionSet(
				backend.NewReduceAction(utils.NewBitset(0), 7),
			),
		}
		parser := backend.Parser{States: []backend.State{state}}

		backend.ApplyDefaultReductions(&parser)

		Expect(parser.States[0].DefaultReduceProductionIdx).To(Equal(ptr(7)))
		Expect(parser.States[0].ReduceActions.IsEmpty()).To(BeTrue())
	})

	It("compresses a state which also shifts, without touching its shift", func() {
		transitionActions := backend.NewTransitionActionSet()
		state := backend.State{
			TransitionActions: transitionActions,
			ReduceActions: backend.NewReduceActionSet(
				backend.NewReduceAction(utils.NewBitset(1, 2), 3),
			),
		}
		parser := backend.Parser{States: []backend.State{state}}

		backend.ApplyDefaultReductions(&parser)

		Expect(parser.States[0].DefaultReduceProductionIdx).To(Equal(ptr(3)))
		Expect(parser.States[0].ReduceActions.IsEmpty()).To(BeTrue())
		// The shift set is left exactly as it was.
		Expect(parser.States[0].TransitionActions).To(Equal(transitionActions))
	})

	It("leaves the accept action, encoded as a reduce with an empty lookahead set, untouched", func() {
		state := backend.State{
			ReduceActions: backend.NewReduceActionSet(
				backend.NewReduceAction(utils.NewBitset(), 0),
			),
		}
		parser := backend.Parser{States: []backend.State{state}}

		backend.ApplyDefaultReductions(&parser)

		Expect(parser.States[0].DefaultReduceProductionIdx).To(BeNil())
		Expect(parser.States[0].ReduceActions).To(Equal(backend.NewReduceActionSet(
			backend.NewReduceAction(utils.NewBitset(), 0),
		)))
	})

	It("does not overwrite a default reduction the state already carries, staying idempotent", func() {
		state := backend.State{
			DefaultReduceProductionIdx: ptr(9),
			ReduceActions: backend.NewReduceActionSet(
				backend.NewReduceAction(utils.NewBitset(0, 1, 2), 4),
			),
		}
		parser := backend.Parser{States: []backend.State{state}}

		backend.ApplyDefaultReductions(&parser)

		// The pre-existing default is kept and the explicit reduce actions are left alone.
		Expect(parser.States[0].DefaultReduceProductionIdx).To(Equal(ptr(9)))
		Expect(parser.States[0].ReduceActions).To(Equal(backend.NewReduceActionSet(
			backend.NewReduceAction(utils.NewBitset(0, 1, 2), 4),
		)))
	})

	It("leaves a state without any reduce action alone", func() {
		parser := backend.Parser{States: []backend.State{{}}}

		backend.ApplyDefaultReductions(&parser)

		Expect(parser.States[0].DefaultReduceProductionIdx).To(BeNil())
	})

	Describe("states which shift the error symbol", func() {
		// errorGrammar is a grammar carrying the error symbol at terminal index 1, the position the GoLR frontend gives
		// it in an augmented grammar: behind the end of input marker, which AugmentGrammar puts first.
		errorGrammar := frontend.Grammar{
			Terminals: []frontend.Symbol{
				frontend.SymbolEOF,
				frontend.SymbolError,
				{Name: "SEMICOLON"},
			},
		}
		errorRef := frontend.NewTerminalRef(1)
		otherRef := frontend.NewTerminalRef(2)

		It("leaves a state which shifts the error symbol uncompressed", func() {
			state := backend.State{
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(errorRef, 42),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(utils.NewBitset(0, 2), 3),
				),
			}
			parser := backend.Parser{Grammar: errorGrammar, States: []backend.State{state}}

			backend.ApplyDefaultReductions(&parser)

			// The state has to report the error where it happens, so that the recovery handler still finds it on the
			// stack. Its reduce action stays explicit.
			Expect(parser.States[0].DefaultReduceProductionIdx).To(BeNil())
			Expect(parser.States[0].ReduceActions).To(Equal(backend.NewReduceActionSet(
				backend.NewReduceAction(utils.NewBitset(0, 2), 3),
			)))
		})

		It("finds the error transition among the transitions of other symbols", func() {
			state := backend.State{
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(0), 7),
					backend.NewTransitionAction(errorRef, 42),
					backend.NewTransitionAction(otherRef, 8),
					backend.NewTransitionAction(frontend.NewNonterminalRef(0), 9),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(utils.NewBitset(0, 2), 3),
				),
			}
			parser := backend.Parser{Grammar: errorGrammar, States: []backend.State{state}}

			backend.ApplyDefaultReductions(&parser)

			Expect(parser.States[0].DefaultReduceProductionIdx).To(BeNil())
		})

		It("compresses a state of the same grammar which shifts another symbol", func() {
			state := backend.State{
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(otherRef, 42),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(utils.NewBitset(0, 2), 3),
				),
			}
			parser := backend.Parser{Grammar: errorGrammar, States: []backend.State{state}}

			backend.ApplyDefaultReductions(&parser)

			Expect(parser.States[0].DefaultReduceProductionIdx).To(Equal(ptr(3)))
			Expect(parser.States[0].ReduceActions.IsEmpty()).To(BeTrue())
		})

		It("compresses a state whose reduce lookahead set contains the error symbol", func() {
			state := backend.State{
				ReduceActions: backend.NewReduceActionSet(
					// Terminal 1 is the error symbol. It never arrives from the scanner, so a lookahead set which
					// contains it does not make the state a resynchronization point.
					backend.NewReduceAction(utils.NewBitset(0, 1), 3),
				),
			}
			parser := backend.Parser{Grammar: errorGrammar, States: []backend.State{state}}

			backend.ApplyDefaultReductions(&parser)

			Expect(parser.States[0].DefaultReduceProductionIdx).To(Equal(ptr(3)))
			Expect(parser.States[0].ReduceActions.IsEmpty()).To(BeTrue())
		})

		It("compresses a state of a grammar which does not carry the error symbol at all", func() {
			grammar := frontend.Grammar{
				Terminals: []frontend.Symbol{
					frontend.SymbolEOF,
					{Name: "SEMICOLON"},
				},
			}
			state := backend.State{
				TransitionActions: backend.NewTransitionActionSet(
					// The terminal index the error symbol would occupy in the grammar above, held by an ordinary
					// terminal here.
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 42),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(utils.NewBitset(0), 3),
				),
			}
			parser := backend.Parser{Grammar: grammar, States: []backend.State{state}}

			backend.ApplyDefaultReductions(&parser)

			Expect(parser.States[0].DefaultReduceProductionIdx).To(Equal(ptr(3)))
			Expect(parser.States[0].ReduceActions.IsEmpty()).To(BeTrue())
		})
	})
})
