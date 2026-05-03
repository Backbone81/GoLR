package nfa_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	thompsonsnfa "golr/internal/scannergen/core/subset/nfa"
	"golr/internal/scannergen/frontend"
	"golr/internal/scannergen/frontend/dsl"
)

var _ = Describe("Or", func() {
	It("should create the correct NFA with two alternatives", func() {
		expression := dsl.Or(
			dsl.Literal("a"),
			dsl.Literal("b"),
		)
		gotNfa := thompsonsnfa.FromRegex(expression, 0)

		wantNfa := []thompsonsnfa.State{
			{ // state 0
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 1,
					},
					{
						Empty:        true,
						NextStateIdx: 3,
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
						NextStateIdx: 5,
					},
				},
			},
			{ // state 3
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  'b',
							High: 'b',
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
