package nfa_test

import (
	thompsonsnfa "golr/internal/scannergen/core/subset/nfa"
	"golr/internal/scannergen/frontend"
	"golr/internal/scannergen/frontend/dsl"
	"unicode"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Any", func() {
	It("should create the correct NFA", func() {
		expression := dsl.Any()
		gotNfa := thompsonsnfa.FromRegex(expression, 0)

		wantNfa := []thompsonsnfa.State{
			{ // state 0
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  0,
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
