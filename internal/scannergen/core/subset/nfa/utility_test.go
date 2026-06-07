package nfa_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/scannergen/backend"
	thompsonsnfa "github.com/backbone81/golr/internal/scannergen/core/subset/nfa"
)

var _ = Describe("Utility", func() {
	Context("Merge", func() {
		It("should panic on empty parameters", func() {
			Expect(func() { thompsonsnfa.Merge() }).To(Panic())
		})

		It("should correctly merge NFAs", func() {
			nfa1 := []thompsonsnfa.State{
				{ // state 0
					Transitions: []thompsonsnfa.Transition{
						{
							ByteRange: backend.ByteRange{
								Low:  's',
								High: 's',
							},
							NextStateIdx: 1,
						},
					},
				},
				{ // state 1
					Accept: true,
				},
			}
			nfa2 := []thompsonsnfa.State{
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

			gotNfa := thompsonsnfa.Merge(nfa1, nfa2)

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
							ByteRange: backend.ByteRange{
								Low:  's',
								High: 's',
							},
							NextStateIdx: 2,
						},
					},
				},
				{ // state 2
					Accept: true,
				},
				{ // state 3
					Transitions: []thompsonsnfa.Transition{
						{
							ByteRange: backend.ByteRange{
								Low:  'a',
								High: 'a',
							},
							NextStateIdx: 4,
						},
					},
				},
				{ // state 4
					Accept: true,
				},
			}
			Expect(gotNfa).To(Equal(wantNfa))
		})
	})
})
