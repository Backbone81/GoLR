package nfa_test

import (
	thompsonsnfa "golr/internal/scannergen/core/subset/nfa"
	"golr/internal/scannergen/frontend"
	"unicode"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CharClass", func() {
	It("should create the correct NFA with a single character range", func() {
		expression := frontend.NewNodeCharClass(frontend.CharClass{
			Ranges: []frontend.CharRange{
				{
					Low:  'a',
					High: 'f',
				},
			},
		})
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
		expression := frontend.NewNodeCharClass(frontend.CharClass{
			Ranges: []frontend.CharRange{
				{
					Low:  'a',
					High: 'f',
				},
				{
					Low:  'x',
					High: 'z',
				},
			},
		})
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
		expression := frontend.NewNodeCharClass(frontend.CharClass{
			Negate: true,
			Ranges: []frontend.CharRange{
				{
					Low:  'u',
					High: 'w',
				},
			},
		})
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
		expression := frontend.NewNodeCharClass(frontend.CharClass{
			Negate: true,
			Ranges: []frontend.CharRange{
				{
					Low:  'b',
					High: 'f',
				},
				{
					Low:  'x',
					High: 'y',
				},
			},
		})
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
