package nfa_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/scannergen/backend"
	thompsonsnfa "github.com/backbone81/golr/internal/scannergen/core/subset/nfa"
	"github.com/backbone81/golr/internal/scannergen/frontend/dsl"
)

var _ = Describe("Concat", func() {
	It("should create the correct NFA with two children", func() {
		expression := dsl.Concat(
			dsl.Literal("a"),
			dsl.Literal("b"),
		)
		gotNfa := thompsonsnfa.FromRegex(expression, 0)

		wantNfa := []thompsonsnfa.State{
			{ // state 0
				Transitions: []thompsonsnfa.Transition{
					{
						ByteRange: backend.ByteRange{
							Low:  'a',
							High: 'a',
						},
						NextStateIdx: 1,
					},
				},
			},
			{ // state 1
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 2,
					},
				},
			},
			{ // state 2
				Transitions: []thompsonsnfa.Transition{
					{
						ByteRange: backend.ByteRange{
							Low:  'b',
							High: 'b',
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
