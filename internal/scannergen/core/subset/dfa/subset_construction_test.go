package dfa_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/scannergen/backend"
	"github.com/backbone81/golr/internal/scannergen/core/subset/dfa"
	"github.com/backbone81/golr/internal/scannergen/core/subset/nfa"
	"github.com/backbone81/golr/internal/utils"
)

var _ = Describe("SubsetConstruction", func() {
	It("should produce the correct DFA for the NFA corresponding to 'a(b|c)*'", func() {
		var n0, n1, n2, n3, n4, n5, n6, n7, n8, n9 nfa.State
		n0.Transitions = []nfa.Transition{
			{
				ByteRange: backend.ByteRange{
					Low:  'a',
					High: 'a',
				},
				NextStateIdx: 1,
			},
		}
		n1.Transitions = []nfa.Transition{
			{
				Empty:        true,
				NextStateIdx: 2,
			},
		}
		n2.Transitions = []nfa.Transition{
			{
				Empty:        true,
				NextStateIdx: 3,
			},
			{
				Empty:        true,
				NextStateIdx: 9,
			},
		}
		n3.Transitions = []nfa.Transition{
			{
				Empty:        true,
				NextStateIdx: 4,
			},
			{
				Empty:        true,
				NextStateIdx: 6,
			},
		}
		n4.Transitions = []nfa.Transition{
			{
				ByteRange: backend.ByteRange{
					Low:  'b',
					High: 'b',
				},
				NextStateIdx: 5,
			},
		}
		n5.Transitions = []nfa.Transition{
			{
				Empty:        true,
				NextStateIdx: 8,
			},
		}
		n6.Transitions = []nfa.Transition{
			{
				ByteRange: backend.ByteRange{
					Low:  'c',
					High: 'c',
				},
				NextStateIdx: 7,
			},
		}
		n7.Transitions = []nfa.Transition{
			{
				Empty:        true,
				NextStateIdx: 8,
			},
		}
		n8.Transitions = []nfa.Transition{
			{
				Empty:        true,
				NextStateIdx: 3,
			},
			{
				Empty:        true,
				NextStateIdx: 9,
			},
		}
		n9.Accept = true

		got := dfa.NewSubsetConstruction([]nfa.State{n0, n1, n2, n3, n4, n5, n6, n7, n8, n9}).Build()

		var d0, d1, d2, d3 backend.State
		d0.Transitions = []backend.Transition{
			{
				ByteRange: backend.ByteRange{
					Low:  'a',
					High: 'a',
				},
				StateIdx: 1,
			},
		}
		d1.Transitions = []backend.Transition{
			{
				ByteRange: backend.ByteRange{
					Low:  'b',
					High: 'b',
				},
				StateIdx: 2,
			},
			{
				ByteRange: backend.ByteRange{
					Low:  'c',
					High: 'c',
				},
				StateIdx: 3,
			},
		}
		d1.Accept = true
		d2.Transitions = []backend.Transition{
			{
				ByteRange: backend.ByteRange{
					Low:  'b',
					High: 'b',
				},
				StateIdx: 2,
			},
			{
				ByteRange: backend.ByteRange{
					Low:  'c',
					High: 'c',
				},
				StateIdx: 3,
			},
		}
		d2.Accept = true
		d3.Transitions = []backend.Transition{
			{
				ByteRange: backend.ByteRange{
					Low:  'b',
					High: 'b',
				},
				StateIdx: 2,
			},
			{
				ByteRange: backend.ByteRange{
					Low:  'c',
					High: 'c',
				},
				StateIdx: 3,
			},
		}
		d3.Accept = true
		want := []backend.State{
			d0, d1, d2, d3,
		}
		Expect(got).To(Equal(want))
	})

	Context("EmptyClosure", func() {
		It("should return the identity for states without an empty transition", func() {
			var a1, a2, b1, b2 nfa.State
			a1.Transitions = []nfa.Transition{
				{
					ByteRange: backend.ByteRange{
						Low:  'a',
						High: 'a',
					},
					NextStateIdx: 1,
				},
			}
			b1.Transitions = []nfa.Transition{
				{
					ByteRange: backend.ByteRange{
						Low:  'b',
						High: 'b',
					},
					NextStateIdx: 3,
				},
			}
			got := dfa.NewSubsetConstruction([]nfa.State{a1, a2, b1, b2}).EmptyClosure(utils.NewOrderedSet(0, 2))
			want := utils.NewOrderedSet[int](0, 2)
			Expect(got.Equal(&want)).To(BeTrue())
		})

		It("should add a state on an empty transition", func() {
			var a1, a2, e1, e2 nfa.State
			a1.Transitions = []nfa.Transition{
				{
					ByteRange: backend.ByteRange{
						Low:  'a',
						High: 'a',
					},
					NextStateIdx: 1,
				},
			}
			e1.Transitions = []nfa.Transition{
				{
					Empty:        true,
					NextStateIdx: 3,
				},
			}
			got := dfa.NewSubsetConstruction([]nfa.State{a1, a2, e1, e2}).EmptyClosure(utils.NewOrderedSet(0, 2))
			want := utils.NewOrderedSet(0, 2, 3)
			Expect(got.Equal(&want)).To(BeTrue())
		})

		It("should behave correctly with loops through empty transitions", func() {
			var a1, a2, e1, e2 nfa.State
			a1.Transitions = []nfa.Transition{
				{
					ByteRange: backend.ByteRange{
						Low:  'a',
						High: 'a',
					},
					NextStateIdx: 1,
				},
			}
			e1.Transitions = []nfa.Transition{
				{
					Empty:        true,
					NextStateIdx: 3,
				},
			}
			e2.Transitions = []nfa.Transition{
				{
					Empty:        true,
					NextStateIdx: 2,
				},
			}
			got := dfa.NewSubsetConstruction([]nfa.State{a1, a2, e1, e2}).EmptyClosure(utils.NewOrderedSet(0, 2))
			want := utils.NewOrderedSet(0, 2, 3)
			Expect(got.Equal(&want)).To(BeTrue())
		})

		It("should behave correctly with multiple states transitioning to the same state", func() {
			var a1, b1, e1 nfa.State
			a1.Transitions = []nfa.Transition{
				{
					Empty:        true,
					NextStateIdx: 2,
				},
			}
			b1.Transitions = []nfa.Transition{
				{
					Empty:        true,
					NextStateIdx: 2,
				},
			}
			got := dfa.NewSubsetConstruction([]nfa.State{a1, b1, e1}).EmptyClosure(utils.NewOrderedSet(0, 1))
			want := utils.NewOrderedSet(0, 1, 2)
			Expect(got.Equal(&want)).To(BeTrue())
		})
	})

	Context("GetCharRanges", func() {
		It("should return an empty slice when there are no transitions", func() {
			var s1 nfa.State
			characterRanges := dfa.NewSubsetConstruction([]nfa.State{s1}).GetByteRanges(utils.NewOrderedSet(0))
			Expect(characterRanges).To(BeEmpty())
		})

		It("should return an empty slice when there are only empty transitions", func() {
			var s1, s2 nfa.State
			s1.Transitions = []nfa.Transition{
				{
					Empty:        true,
					NextStateIdx: 1,
				},
			}
			characterRanges := dfa.NewSubsetConstruction([]nfa.State{s1, s2}).GetByteRanges(utils.NewOrderedSet(0))
			Expect(characterRanges).To(BeEmpty())
		})

		It("should return the identical character range when a single character range is present", func() {
			var s1, s2 nfa.State
			s1.Transitions = []nfa.Transition{
				{
					ByteRange: backend.ByteRange{
						Low:  'a',
						High: 'z',
					},
					NextStateIdx: 1,
				},
			}
			characterRanges := dfa.NewSubsetConstruction([]nfa.State{s1, s2}).GetByteRanges(utils.NewOrderedSet(0))
			Expect(characterRanges).To(Equal([]backend.ByteRange{
				{
					Low:  'a',
					High: 'z',
				},
			}))
		})

		It("should return the correct character ranges of multiple transitions", func() {
			var a1, b1, s2 nfa.State
			a1.Transitions = []nfa.Transition{
				{
					ByteRange: backend.ByteRange{
						Low:  'a',
						High: 'z',
					},
					NextStateIdx: 2,
				},
			}
			b1.Transitions = []nfa.Transition{
				{
					ByteRange: backend.ByteRange{
						Low:  '0',
						High: '9',
					},
					NextStateIdx: 2,
				},
			}
			characterRanges := dfa.NewSubsetConstruction([]nfa.State{a1, b1, s2}).GetByteRanges(utils.NewOrderedSet(0, 1))
			Expect(characterRanges).To(Equal([]backend.ByteRange{
				{
					Low:  'a',
					High: 'z',
				},
				{
					Low:  '0',
					High: '9',
				},
			}))
		})

		It("should return the correct character ranges on overlap", func() {
			var a1, b1, s2 nfa.State
			a1.Transitions = []nfa.Transition{
				{
					ByteRange: backend.ByteRange{
						Low:  'a',
						High: 'u',
					},
					NextStateIdx: 2,
				},
			}
			b1.Transitions = []nfa.Transition{
				{
					ByteRange: backend.ByteRange{
						Low:  'd',
						High: 'z',
					},
					NextStateIdx: 2,
				},
			}
			characterRanges := dfa.NewSubsetConstruction([]nfa.State{a1, b1, s2}).GetByteRanges(utils.NewOrderedSet(0, 1))
			Expect(characterRanges).To(Equal([]backend.ByteRange{
				{
					Low:  'a',
					High: 'c',
				},
				{
					Low:  'd',
					High: 'u',
				},
				{
					Low:  'v',
					High: 'z',
				},
			}))
		})

		It("should return the correct character ranges on identity", func() {
			var a1, b1, s2 nfa.State
			a1.Transitions = []nfa.Transition{
				{
					ByteRange: backend.ByteRange{
						Low:  'a',
						High: 'z',
					},
					NextStateIdx: 2,
				},
			}
			b1.Transitions = []nfa.Transition{
				{
					ByteRange: backend.ByteRange{
						Low:  'a',
						High: 'z',
					},
					NextStateIdx: 2,
				},
			}
			characterRanges := dfa.NewSubsetConstruction([]nfa.State{a1, b1, s2}).GetByteRanges(utils.NewOrderedSet(0, 1))
			Expect(characterRanges).To(Equal([]backend.ByteRange{
				{
					Low:  'a',
					High: 'z',
				},
			}))
		})
	})
})
