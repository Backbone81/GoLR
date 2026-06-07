package nfa_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	thompsonsnfa "github.com/backbone81/golr/internal/scannergen/core/subset/nfa"
	"github.com/backbone81/golr/internal/scannergen/frontend"
	"github.com/backbone81/golr/internal/scannergen/frontend/dsl"
)

var _ = Describe("Any", func() {
	It("should create the correct NFA", func() {
		expression := dsl.Any()
		gotNfa := thompsonsnfa.FromRegex(expression, 0)

		wantNfa := []thompsonsnfa.State{
			{ // state 0: start
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  0x00,
							High: '\n' - 1,
						},
						NextStateIdx: 4,
					},
					{
						CharRange: frontend.CharRange{
							Low:  '\n' + 1,
							High: 0x7F,
						},
						NextStateIdx: 4,
					},
					{
						CharRange: frontend.CharRange{
							Low:  0xC2,
							High: 0xDF,
						},
						NextStateIdx: 1,
					},
					{
						CharRange: frontend.CharRange{
							Low:  0xE0,
							High: 0xEF,
						},
						NextStateIdx: 2,
					},
					{
						CharRange: frontend.CharRange{
							Low:  0xF0,
							High: 0xF4,
						},
						NextStateIdx: 3,
					},
				},
			},
			{ // state 1:  expecting 1 more continuation byte, then accept
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  0x80,
							High: 0xBF,
						},
						NextStateIdx: 4,
					},
				},
			},
			{ // state 2: expecting 2 more continuation bytes
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  0x80,
							High: 0xBF,
						},
						NextStateIdx: 1,
					},
				},
			},
			{ // state 3: expecting 3 more continuation bytes
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  0x80,
							High: 0xBF,
						},
						NextStateIdx: 2,
					},
				},
			},
			{ // state 4: accept
				Accept: true,
			},
		}
		Expect(gotNfa).To(Equal(wantNfa))
	})
})
