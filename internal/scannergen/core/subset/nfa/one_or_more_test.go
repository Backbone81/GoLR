package nfa_test

import (
	thompsonsnfa "golr/internal/scannergen/core/subset/nfa"
	"golr/internal/scannergen/frontend"
	"golr/internal/scannergen/frontend/dsl"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("OneOrMore", func() {
	It("should create the correct NFA with a single character literal", func() {
		expression := dsl.OneOrMore(dsl.Literal("a"))
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
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 0,
					},
				},
			},
		}
		Expect(gotNfa).To(Equal(wantNfa))
	})
})
