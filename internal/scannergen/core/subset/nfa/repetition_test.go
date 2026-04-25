package nfa_test

import (
	thompsonsnfa "golr/internal/scannergen/core/subset/nfa"
	"golr/internal/scannergen/frontend"
	"golr/internal/scannergen/frontend/dsl"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Repetition", func() {
	It("should create the correct NFA with one repetitions", func() {
		expression := dsl.Repetition(dsl.Literal("a"), 1, 1)
		gotNfa := thompsonsnfa.FromRegex(expression, 0)

		wantNfa := []thompsonsnfa.State{
			{ // state 0
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 1,
					},
				},
			},
			{ // state 1
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  'a',
							High: 'a',
						},
						NextStateIdx: 2,
					},
				},
			},
			{ // state 2
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 3,
					},
				},
			},
			{ // state 3
				Accept: true,
			},
		}
		Expect(gotNfa).To(Equal(wantNfa))
	})

	It("should create the correct NFA with two repetitions", func() {
		expression := dsl.Repetition(dsl.Literal("a"), 2, 2)
		gotNfa := thompsonsnfa.FromRegex(expression, 0)

		wantNfa := []thompsonsnfa.State{
			{ // state 0
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 1,
					},
				},
			},
			{ // state 1
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  'a',
							High: 'a',
						},
						NextStateIdx: 2,
					},
				},
			},
			{ // state 2
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 3,
					},
				},
			},
			{ // state 3
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  'a',
							High: 'a',
						},
						NextStateIdx: 4,
					},
				},
			},
			{ // state 4
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 5,
					},
				},
			},
			{ // state 5
				Accept: true,
			},
		}
		Expect(gotNfa).To(Equal(wantNfa))
	})

	It("should create the correct NFA with one to two repetitions", func() {
		expression := dsl.Repetition(dsl.Literal("a"), 1, 2)
		gotNfa := thompsonsnfa.FromRegex(expression, 0)

		wantNfa := []thompsonsnfa.State{
			{ // state 0
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 1,
					},
				},
			},
			{ // state 1
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  'a',
							High: 'a',
						},
						NextStateIdx: 2,
					},
				},
			},
			{ // state 2
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 3,
					},
					{
						Empty:        true,
						NextStateIdx: 5,
					},
				},
			},
			{ // state 3
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  'a',
							High: 'a',
						},
						NextStateIdx: 4,
					},
				},
			},
			{ // state 4
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 5,
					},
				},
			},
			{ // state 5
				Accept: true,
			},
		}

		Expect(gotNfa).To(Equal(wantNfa))
	})
})
