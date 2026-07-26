package backend_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

var _ = Describe("ErrorShiftStateIdx", func() {
	// The error symbol sits at terminal index 1 here, the position the GoLR frontend gives it in an augmented grammar:
	// behind the end of input marker, which AugmentGrammar puts first.
	errorRef := frontend.NewTerminalRef(1)
	otherRef := frontend.NewTerminalRef(2)

	It("returns the state the error transition leads to", func() {
		state := backend.State{
			TransitionActions: backend.NewTransitionActionSet(
				backend.NewTransitionAction(errorRef, 42),
			),
		}

		stateIdx, ok := backend.ErrorShiftStateIdx(&state, errorRef)

		Expect(ok).To(BeTrue())
		Expect(stateIdx).To(Equal(42))
	})

	It("finds the error transition among the transitions of other symbols", func() {
		state := backend.State{
			TransitionActions: backend.NewTransitionActionSet(
				backend.NewTransitionAction(frontend.NewTerminalRef(0), 7),
				backend.NewTransitionAction(errorRef, 42),
				backend.NewTransitionAction(otherRef, 8),
				backend.NewTransitionAction(frontend.NewNonterminalRef(0), 9),
			),
		}

		stateIdx, ok := backend.ErrorShiftStateIdx(&state, errorRef)

		Expect(ok).To(BeTrue())
		Expect(stateIdx).To(Equal(42))
	})

	It("reports no transition for a state which shifts other symbols only", func() {
		state := backend.State{
			TransitionActions: backend.NewTransitionActionSet(
				backend.NewTransitionAction(frontend.NewTerminalRef(0), 7),
				backend.NewTransitionAction(otherRef, 8),
			),
		}

		_, ok := backend.ErrorShiftStateIdx(&state, errorRef)

		Expect(ok).To(BeFalse())
	})

	It("reports no transition for a state which shifts nothing beyond the error symbol index", func() {
		// The transition set is ordered by symbol, so a state whose transitions all sit below the error symbol makes the
		// search run off the end of the set.
		state := backend.State{
			TransitionActions: backend.NewTransitionActionSet(
				backend.NewTransitionAction(frontend.NewTerminalRef(0), 7),
			),
		}

		_, ok := backend.ErrorShiftStateIdx(&state, errorRef)

		Expect(ok).To(BeFalse())
	})

	It("reports no transition for a state without any transition", func() {
		state := backend.State{}

		_, ok := backend.ErrorShiftStateIdx(&state, errorRef)

		Expect(ok).To(BeFalse())
	})
})
