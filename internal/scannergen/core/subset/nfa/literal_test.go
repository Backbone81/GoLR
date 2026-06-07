package nfa_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/scannergen/backend"
	thompsonsnfa "github.com/backbone81/golr/internal/scannergen/core/subset/nfa"
	"github.com/backbone81/golr/internal/scannergen/frontend/dsl"
)

var _ = Describe("Literal", func() {
	It("should create the correct NFA with a single character string", func() {
		expression := dsl.Literal("a")
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
						ByteRange: backend.ByteRange{
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
						ByteRange: backend.ByteRange{
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
						ByteRange: backend.ByteRange{
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
