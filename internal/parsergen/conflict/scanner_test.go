package conflict_test

import (
	"maps"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

// The scanner works on the actions of a state alone, so the states of these tests are built by hand instead of being
// generated from a grammar. That keeps what the scanner is expected to report next to the actions it reports it from,
// and it reaches the states a grammar of a test would never produce.
var _ = Describe("ConflictScanner", func() {
	DescribeTable("should report the conflicted terminals of a state",
		func(state backend.State, want []conflict.TerminalContributions) {
			var scanner conflict.Scanner
			Expect(scanner.Conflicts(&state)).To(Equal(want))
		},
		Entry("a state without any action has no conflict",
			backend.State{},
			nil,
		),
		Entry("a shift alone is no conflict",
			newTestState(
				[]int{1},
				nil,
			),
			nil,
		),
		Entry("a reduction alone is no conflict",
			newTestState(
				nil,
				map[int][]int{7: {1, 2, 3}},
			),
			nil,
		),
		Entry("two reductions on terminals of their own are no conflict",
			newTestState(
				nil,
				map[int][]int{7: {1, 2}, 8: {3, 4}},
			),
			nil,
		),
		Entry("a shift and a reduction on the same terminal is a shift/reduce conflict",
			newTestState(
				[]int{2},
				map[int][]int{7: {1, 2, 3}},
			),
			[]conflict.TerminalContributions{
				{
					TerminalIdx: 2,
					Contributions: conflict.NewContributionSet(
						conflict.NewShiftContribution(),
						conflict.NewReduceContribution(7),
					),
				},
			},
		),
		Entry("two reductions on the same terminal are a reduce/reduce conflict",
			newTestState(
				nil,
				map[int][]int{7: {1, 2}, 8: {2, 3}},
			),
			[]conflict.TerminalContributions{
				{
					TerminalIdx: 2,
					Contributions: conflict.NewContributionSet(
						conflict.NewReduceContribution(7),
						conflict.NewReduceContribution(8),
					),
				},
			},
		),
		// A terminal is recorded as conflicted once per action beyond the first, so a terminal with three actions is
		// the case where recording it more than once must not report it more than once.
		Entry("a shift and two reductions on the same terminal are a single conflict with three contributions",
			newTestState(
				[]int{2},
				map[int][]int{7: {2}, 8: {2}},
			),
			[]conflict.TerminalContributions{
				{
					TerminalIdx: 2,
					Contributions: conflict.NewContributionSet(
						conflict.NewShiftContribution(),
						conflict.NewReduceContribution(7),
						conflict.NewReduceContribution(8),
					),
				},
			},
		),
		Entry("a goto never takes part in a conflict",
			backend.State{
				// The nonterminal index collides with the terminal index of the reduction on purpose: a goto and a
				// reduction which carry the same index are not two actions on the same terminal.
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewNonterminalRef(2), 1),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(utils.NewBitset(2), 7),
				),
			},
			nil,
		),
		Entry("conflicts are reported in ascending terminal order",
			newTestState(
				[]int{9, 1, 5},
				map[int][]int{7: {1, 5, 9}},
			),
			[]conflict.TerminalContributions{
				{
					TerminalIdx:   1,
					Contributions: conflict.NewContributionSet(conflict.NewShiftContribution(), conflict.NewReduceContribution(7)),
				},
				{
					TerminalIdx:   5,
					Contributions: conflict.NewContributionSet(conflict.NewShiftContribution(), conflict.NewReduceContribution(7)),
				},
				{
					TerminalIdx:   9,
					Contributions: conflict.NewContributionSet(conflict.NewShiftContribution(), conflict.NewReduceContribution(7)),
				},
			},
		),
		// The scanner keeps its terminals in bitsets, so a terminal beyond the first chunk is the case where those
		// bitsets have to grow to reach it.
		Entry("a conflict on a terminal beyond the first bitset chunk is reported",
			newTestState(
				[]int{200},
				map[int][]int{7: {200}},
			),
			[]conflict.TerminalContributions{
				{
					TerminalIdx:   200,
					Contributions: conflict.NewContributionSet(conflict.NewShiftContribution(), conflict.NewReduceContribution(7)),
				},
			},
		),
		Entry("a conflict in a later chunk does not hide one in an earlier chunk",
			newTestState(
				[]int{3, 200},
				map[int][]int{7: {3, 200}},
			),
			[]conflict.TerminalContributions{
				{
					TerminalIdx:   3,
					Contributions: conflict.NewContributionSet(conflict.NewShiftContribution(), conflict.NewReduceContribution(7)),
				},
				{
					TerminalIdx:   200,
					Contributions: conflict.NewContributionSet(conflict.NewShiftContribution(), conflict.NewReduceContribution(7)),
				},
			},
		),
	)

	// The scanner keeps its bitsets across states so that they stop growing, which is only correct as long as nothing
	// of the state it saw before is left in them.
	It("should not carry the conflicts of one state over to the next", func() {
		conflicted := newTestState([]int{2, 200}, map[int][]int{7: {2, 200}})
		clean := newTestState([]int{3}, map[int][]int{7: {4}})

		var reused conflict.Scanner
		Expect(reused.Conflicts(&conflicted)).ToNot(BeEmpty())
		Expect(reused.Conflicts(&clean)).To(BeEmpty())

		// Whatever the scanner saw before, it has to agree with a scanner which saw nothing at all.
		var fresh conflict.Scanner
		Expect(reused.Conflicts(&conflicted)).To(Equal(fresh.Conflicts(&conflicted)))
	})
})

// newTestState builds a state which shifts each of the terminals and reduces each of the productions on the terminals
// of its lookahead set.
func newTestState(shiftTerminalIdxs []int, lookaheadTerminalIdxsByProductionIdx map[int][]int) backend.State {
	var transitionActions backend.TransitionActionSet
	for i, terminalIdx := range shiftTerminalIdxs {
		// The target state is irrelevant for the scanner, but it has to differ per transition so that the transitions
		// do not collapse into one another in the set.
		transitionActions.Add(backend.NewTransitionAction(frontend.NewTerminalRef(terminalIdx), i+1))
	}

	var reduceActions backend.ReduceActionSet
	for _, productionIdx := range slices.Sorted(maps.Keys(lookaheadTerminalIdxsByProductionIdx)) {
		reduceActions.Add(backend.NewReduceAction(
			utils.NewBitset(lookaheadTerminalIdxsByProductionIdx[productionIdx]...),
			productionIdx,
		))
	}

	return backend.State{
		TransitionActions: transitionActions,
		ReduceActions:     reduceActions,
	}
}
