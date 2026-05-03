package nfa_test

import (
	"unicode"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	thompsonsnfa "golr/internal/scannergen/core/subset/nfa"
	"golr/internal/scannergen/frontend"
	"golr/internal/scannergen/frontend/dsl"
)

var _ = Describe("CharClass", func() {
	It("should create the correct NFA with a single character range", func() {
		expression := dsl.CharClass(
			dsl.CharRange('a', 'f'),
		)
		gotNfa := thompsonsnfa.FromRegex(expression, 0)

		wantNfa := []thompsonsnfa.State{
			{ // state 0
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  'a',
							High: 'f',
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

	It("should create the correct NFA with two character ranges", func() {
		expression := dsl.CharClass(
			dsl.CharRange('a', 'f'),
			dsl.CharRange('x', 'z'),
		)
		gotNfa := thompsonsnfa.FromRegex(expression, 0)

		wantNfa := []thompsonsnfa.State{
			{ // state 0
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  'a',
							High: 'f',
						},
						NextStateIdx: 1,
					},
					{
						CharRange: frontend.CharRange{
							Low:  'x',
							High: 'z',
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

	It("should create the correct NFA with a single character range negated", func() {
		expression := dsl.NegCharClass(
			dsl.CharRange('u', 'w'),
		)
		gotNfa := thompsonsnfa.FromRegex(expression, 0)

		wantNfa := []thompsonsnfa.State{
			{ // state 0
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  0,
							High: 't',
						},
						NextStateIdx: 1,
					},
					{
						CharRange: frontend.CharRange{
							Low:  'x',
							High: unicode.MaxRune,
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

	It("should create the correct NFA with two character ranges negated", func() {
		expression := dsl.NegCharClass(
			dsl.CharRange('b', 'f'),
			dsl.CharRange('x', 'y'),
		)
		gotNfa := thompsonsnfa.FromRegex(expression, 0)

		wantNfa := []thompsonsnfa.State{
			{ // state 0
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  0,
							High: 'a',
						},
						NextStateIdx: 1,
					},
					{
						CharRange: frontend.CharRange{
							Low:  'g',
							High: 'w',
						},
						NextStateIdx: 1,
					},
					{
						CharRange: frontend.CharRange{
							Low:  'z',
							High: unicode.MaxRune,
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
})
