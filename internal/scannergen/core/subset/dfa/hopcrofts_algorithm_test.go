package dfa_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"golr/internal/scannergen/backend"
	"golr/internal/scannergen/core/subset/dfa"
	"golr/internal/scannergen/frontend"
)

var _ = Describe("HopcroftsAlgorithm", func() {
	It("should produce the correct minimal DFA for the DFA corresponding to 'fee|fie'", func() {
		var s0, s1, s2, s3, s4, s5 backend.State
		s0.Transitions = []backend.Transition{
			{
				CharRange: frontend.CharRange{
					Low:  'f',
					High: 'f',
				},
				StateIdx: 1,
			},
		}
		s1.Transitions = []backend.Transition{
			{
				CharRange: frontend.CharRange{
					Low:  'e',
					High: 'e',
				},
				StateIdx: 2,
			},
			{
				CharRange: frontend.CharRange{
					Low:  'i',
					High: 'i',
				},
				StateIdx: 4,
			},
		}
		s2.Transitions = []backend.Transition{
			{
				CharRange: frontend.CharRange{
					Low:  'e',
					High: 'e',
				},
				StateIdx: 3,
			},
		}
		s3.Accept = true
		s4.Transitions = []backend.Transition{
			{
				CharRange: frontend.CharRange{
					Low:  'e',
					High: 'e',
				},
				StateIdx: 5,
			},
		}
		s5.Accept = true
		inputDFA := []backend.State{s0, s1, s2, s3, s4, s5}

		got := dfa.NewHopcroftsAlgorithm().Build(inputDFA)

		var m0, m1, m2, m3 backend.State
		m0.Transitions = []backend.Transition{
			{
				CharRange: frontend.CharRange{
					Low:  'f',
					High: 'f',
				},
				StateIdx: 1,
			},
		}
		m1.Transitions = []backend.Transition{
			{
				CharRange: frontend.CharRange{
					Low:  'e',
					High: 'e',
				},
				StateIdx: 2,
			},
			{
				CharRange: frontend.CharRange{
					Low:  'i',
					High: 'i',
				},
				StateIdx: 2,
			},
		}
		m2.Transitions = []backend.Transition{
			{
				CharRange: frontend.CharRange{
					Low:  'e',
					High: 'e',
				},
				StateIdx: 3,
			},
		}
		m3.Accept = true
		want := []backend.State{m0, m1, m2, m3}
		Expect(got).To(Equal(want))
	})

	It("should produce the correct minimal DFA for the DFA corresponding to 'a(b|c)*'", func() {
		var d0, d1, d2, d3 backend.State
		d0.Transitions = []backend.Transition{
			{
				CharRange: frontend.CharRange{
					Low:  'a',
					High: 'a',
				},
				StateIdx: 1,
			},
		}
		d1.Transitions = []backend.Transition{
			{
				CharRange: frontend.CharRange{
					Low:  'b',
					High: 'b',
				},
				StateIdx: 2,
			},
			{
				CharRange: frontend.CharRange{
					Low:  'c',
					High: 'c',
				},
				StateIdx: 3,
			},
		}
		d1.Accept = true
		d2.Transitions = []backend.Transition{
			{
				CharRange: frontend.CharRange{
					Low:  'b',
					High: 'b',
				},
				StateIdx: 2,
			},
			{
				CharRange: frontend.CharRange{
					Low:  'c',
					High: 'c',
				},
				StateIdx: 3,
			},
		}
		d2.Accept = true
		d3.Transitions = []backend.Transition{
			{
				CharRange: frontend.CharRange{
					Low:  'b',
					High: 'b',
				},
				StateIdx: 2,
			},
			{
				CharRange: frontend.CharRange{
					Low:  'c',
					High: 'c',
				},
				StateIdx: 3,
			},
		}
		d3.Accept = true
		inputDFA := []backend.State{d0, d1, d2, d3}

		got := dfa.NewHopcroftsAlgorithm().Build(inputDFA)

		var m0, m1 backend.State
		m0.Transitions = []backend.Transition{
			{
				CharRange: frontend.CharRange{
					Low:  'a',
					High: 'a',
				},
				StateIdx: 1,
			},
		}
		m1.Transitions = []backend.Transition{
			{
				CharRange: frontend.CharRange{
					Low:  'b',
					High: 'b',
				},
				StateIdx: 1,
			},
			{
				CharRange: frontend.CharRange{
					Low:  'c',
					High: 'c',
				},
				StateIdx: 1,
			},
		}
		m1.Accept = true
		want := []backend.State{m0, m1}
		Expect(got).To(Equal(want))
	})
})
