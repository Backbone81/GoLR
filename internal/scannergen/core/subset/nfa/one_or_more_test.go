package nfa_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/scannergen/backend"
	thompsonsnfa "github.com/backbone81/golr/internal/scannergen/core/subset/nfa"
	"github.com/backbone81/golr/internal/scannergen/frontend/dsl"
)

var _ = Describe("OneOrMore", func() {
	It("should create the correct NFA with a single character literal", func() {
		expression := dsl.OneOrMore(dsl.Literal("a"))
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
