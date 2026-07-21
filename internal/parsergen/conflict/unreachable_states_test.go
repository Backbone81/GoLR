package conflict_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

var _ = Describe("RemoveUnreachableStates", func() {
	// A parser table which nothing was removed from has no state to strand, which is the case every grammar without a
	// conflict ends up in. The state indexes of such a parser table must survive untouched, because renumbering states
	// for no reason would churn the output of every backend.
	It("should keep every state of a parser table which has nothing unreachable", func() {
		parser := backend.Parser{
			States: []backend.State{
				stateWithTransitions(0, backend.NewTransitionAction(frontend.NewTerminalRef(0), 1)),
				stateWithTransitions(1, backend.NewTransitionAction(frontend.NewNonterminalRef(0), 2)),
				stateWithTransitions(2),
			},
		}
		conflicts := []conflict.Conflict{
			{StateIdx: 0, TerminalIdx: 7},
			{StateIdx: 2, TerminalIdx: 8},
		}

		parser, conflicts = conflict.RemoveUnreachableStates(parser, conflicts)

		Expect(parser.States).To(Equal([]backend.State{
			stateWithTransitions(0, backend.NewTransitionAction(frontend.NewTerminalRef(0), 1)),
			stateWithTransitions(1, backend.NewTransitionAction(frontend.NewNonterminalRef(0), 2)),
			stateWithTransitions(2),
		}))
		Expect(conflicts).To(
			Equal([]conflict.Conflict{
				{StateIdx: 0, TerminalIdx: 7},
				{StateIdx: 2, TerminalIdx: 8},
			}),
			"no state moved, so every conflict is expected to keep the state index it had",
		)
	})

	// This is what section 3.8.2 of IELR(1) is about. Resolving a conflict which a shift loses removes the transition
	// on the conflicted terminal, and the state that transition led into is stranded when it was the only way in. The
	// states which were only reachable through the stranded state go with it, and the pass has to settle that cascade
	// on its own.
	It("should remove a stranded state together with the states behind it", func() {
		// State 0 reaches state 1 and state 1 reaches state 4. States 2 and 3 are what is left over after the
		// transition which led into state 2 was removed: they are only reachable through each other, so nothing can
		// reach them anymore. State 3 still transitions into state 1, which is reachable, but a transition out of a
		// stranded state does not make the stranded state reachable.
		parser := backend.Parser{
			States: []backend.State{
				stateWithTransitions(0, backend.NewTransitionAction(frontend.NewTerminalRef(0), 1)),
				stateWithTransitions(1, backend.NewTransitionAction(frontend.NewTerminalRef(1), 4)),
				stateWithTransitions(2, backend.NewTransitionAction(frontend.NewTerminalRef(0), 3)),
				stateWithTransitions(
					3,
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 2),
					backend.NewTransitionAction(frontend.NewTerminalRef(2), 1),
				),
				stateWithTransitions(4),
			},
		}

		parser, _ = conflict.RemoveUnreachableStates(parser, nil)

		// The states which are left keep the order they were in, and their transitions point at the states they
		// pointed at before, under the new state indexes.
		Expect(parser.States).To(Equal([]backend.State{
			stateWithTransitions(0, backend.NewTransitionAction(frontend.NewTerminalRef(0), 1)),
			stateWithTransitions(1, backend.NewTransitionAction(frontend.NewTerminalRef(1), 2)),
			stateWithTransitions(4),
		}))
	})

	// A conflict names the state it occurred in, so removing states has to bring those state indexes up to date.
	It("should move the conflicts of the states which were kept to their new state index", func() {
		// State 1 is unreachable, so state 2 moves up into its place and the conflict of state 2 has to move with it.
		parser := backend.Parser{
			States: []backend.State{
				stateWithTransitions(0, backend.NewTransitionAction(frontend.NewTerminalRef(0), 2)),
				stateWithTransitions(1),
				stateWithTransitions(2),
			},
		}
		conflicts := []conflict.Conflict{
			{StateIdx: 0, TerminalIdx: 7},
			{StateIdx: 2, TerminalIdx: 8},
		}

		_, conflicts = conflict.RemoveUnreachableStates(parser, conflicts)

		Expect(conflicts).To(Equal([]conflict.Conflict{
			{StateIdx: 0, TerminalIdx: 7},
			{StateIdx: 1, TerminalIdx: 8},
		}))
	})

	// A conflict of a state which is gone has no state index left to be reported under, and the state it happened in is
	// one no input can reach anymore, so there is nothing left to report.
	It("should drop the conflicts of the states which were removed", func() {
		parser := backend.Parser{
			States: []backend.State{
				stateWithTransitions(0, backend.NewTransitionAction(frontend.NewTerminalRef(0), 1)),
				stateWithTransitions(1),
				stateWithTransitions(2),
			},
		}
		conflicts := []conflict.Conflict{
			{StateIdx: 2, TerminalIdx: 7},
		}

		_, conflicts = conflict.RemoveUnreachableStates(parser, conflicts)

		Expect(conflicts).To(BeEmpty())
	})

	// Reachability is about the symbols the parser can move over, which are terminals and nonterminals alike, so a
	// state which is only reachable through a goto is reachable all the same.
	It("should keep a state which is only reachable through a nonterminal transition", func() {
		parser := backend.Parser{
			States: []backend.State{
				stateWithTransitions(0, backend.NewTransitionAction(frontend.NewNonterminalRef(0), 1)),
				stateWithTransitions(1),
			},
		}

		parser, _ = conflict.RemoveUnreachableStates(parser, nil)

		Expect(parser.States).To(Equal([]backend.State{
			stateWithTransitions(0, backend.NewTransitionAction(frontend.NewNonterminalRef(0), 1)),
			stateWithTransitions(1),
		}))
	})

	// A parser table without any state has no start state to reach anything from, so there is nothing to remove and
	// nothing to index into.
	It("should do nothing to an empty parser table", func() {
		parser := backend.Parser{}

		parser, conflicts := conflict.RemoveUnreachableStates(parser, nil)

		Expect(parser.States).To(BeEmpty())
		Expect(conflicts).To(BeEmpty())
	})

	// The end to end case: real parser tables of an ambiguous grammar, with the conflicts resolved the way a parser
	// generator resolves them. Whatever resolution stranded has to be gone afterwards, and every state which is left
	// has to be reachable and to transition into states which exist.
	It("should leave every state of a resolved parser table reachable", func() {
		parser := buildLR1Parser(conflict.PrecedenceTestGrammar)
		conflicts, err := conflict.Resolve(&parser, conflict.NewDefaultPolicy(conflict.PrecedenceTestGrammar))
		Expect(err).ToNot(HaveOccurred())
		stateCountBefore := len(parser.States)

		parser, conflicts = conflict.RemoveUnreachableStates(parser, conflicts)

		Expect(len(parser.States)).To(BeNumerically("<=", stateCountBefore))
		for stateIdx := range parser.States {
			for _, transitionAction := range parser.States[stateIdx].TransitionActions.All() {
				Expect(transitionAction.StateIdx()).To(
					BeNumerically("<", len(parser.States)),
					"a transition of state %d is expected to point at a state which is still there",
					stateIdx,
				)
			}
		}
		Expect(reachableStateCount(parser)).To(
			Equal(len(parser.States)),
			"every state which is left is expected to be reachable from the start state",
		)
		for _, c := range conflicts {
			Expect(c.StateIdx).To(
				BeNumerically("<", len(parser.States)),
				"a conflict is expected to name a state which is still there",
			)
		}
	})
})

// stateWithTransitions returns a state with the transition actions and a kernel item which tells it apart from the
// other states of a test, so that a state which was moved to a different index can be recognized.
func stateWithTransitions(id int, transitionActions ...backend.TransitionAction) backend.State {
	return backend.State{
		KernelItems:       backend.NewCoreSet(backend.NewCore(id, 0)),
		TransitionActions: backend.NewTransitionActionSet(transitionActions...),
	}
}

// reachableStateCount returns the number of states the parser can reach from the start state.
func reachableStateCount(parser backend.Parser) int {
	reachable := make(map[int]struct{})
	pending := []int{0}
	reachable[0] = struct{}{}
	for len(pending) > 0 {
		stateIdx := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		for _, transitionAction := range parser.States[stateIdx].TransitionActions.All() {
			if _, ok := reachable[transitionAction.StateIdx()]; ok {
				continue
			}
			reachable[transitionAction.StateIdx()] = struct{}{}
			pending = append(pending, transitionAction.StateIdx())
		}
	}
	return len(reachable)
}
