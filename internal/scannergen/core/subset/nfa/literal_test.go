package nfa_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	thompsonsnfa "golr/internal/scannergen/core/subset/nfa"
	"golr/internal/scannergen/frontend"
	"golr/internal/scannergen/frontend/dsl"
)

var _ = Describe("Literal", func() {
	It("should create the correct NFA with a single character string", func() {
		expression := dsl.Literal("a")
		gotNfa := thompsonsnfa.FromRegex(expression, 0)

		wantNfa := []thompsonsnfa.State{
			{ // state 0
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  'a',
							High: 'a',
						},
						NextStateIdx: 1,
					},
				},
			},
			{ // state 1
				Accept: true,
			},
		}
		Expect(gotNfa).To(Equal(wantNfa))
	})

	It("should create the correct NFA with a multi character string", func() {
		expression := dsl.Literal("bar")
		gotNfa := thompsonsnfa.FromRegex(expression, 0)

		wantNfa := []thompsonsnfa.State{
			{ // state 0
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  'b',
							High: 'b',
						},
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
						CharRange: frontend.CharRange{
							Low:  'r',
							High: 'r',
						},
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
})
