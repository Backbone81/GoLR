package dfa_test

import (
	"golr/internal/scannergen/backend"
	"golr/internal/scannergen/core/subset/dfa"
	"golr/internal/scannergen/core/subset/nfa"
	"golr/internal/scannergen/frontend"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DFA", func() {
	Context("FromNFA", func() {
		It("should produce the correct DFA for the NFA corresponding to 'a(b|c)*'", func() {
			var n0, n1, n2, n3, n4, n5, n6, n7, n8, n9 nfa.State
			n0.Transitions = []nfa.Transition{
				{
					CharRange: frontend.CharRange{
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
					CharRange: frontend.CharRange{
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
					CharRange: frontend.CharRange{
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

			got := dfa.FromNFA([]nfa.State{n0, n1, n2, n3, n4, n5, n6, n7, n8, n9})

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
})

//nolint:funlen
func BenchmarkFromNFA(b *testing.B) {
	b.Run("[a-zA-Z_][a-zA-Z0-9_]*", func(b *testing.B) {
		expression := frontend.NewNodeConcat(
			frontend.NewNodeCharClass(frontend.CharClass{
				Ranges: []frontend.CharRange{
					{
						Low:  'a',
						High: 'z',
					},
					{
						Low:  'A',
						High: 'Z',
					},
					{
						Low:  '_',
						High: '_',
					},
				},
			}),
			frontend.NewNodeZeroOrMore(
				frontend.NewNodeCharClass(frontend.CharClass{
					Ranges: []frontend.CharRange{
						{
							Low:  'a',
							High: 'z',
						},
						{
							Low:  'A',
							High: 'Z',
						},
						{
							Low:  '0',
							High: '9',
						},
						{
							Low:  '_',
							High: '_',
						},
					},
				}),
			),
		)
		expressionNfa := nfa.FromRegex(expression, 0)
		b.ResetTimer()
		for range b.N {
			_ = dfa.FromNFA(expressionNfa)
		}
	})

	b.Run("[-+]?[0-9]+", func(b *testing.B) {
		expression := frontend.NewNodeConcat(
			frontend.NewNodeOptional(
				frontend.NewNodeCharClass(frontend.CharClass{
					Ranges: []frontend.CharRange{
						{
							Low:  '-',
							High: '-',
						},
						{
							Low:  '+',
							High: '+',
						},
					},
				}),
			),
			frontend.NewNodeOneOrMore(
				frontend.NewNodeCharClass(frontend.CharClass{
					Ranges: []frontend.CharRange{
						{
							Low:  '0',
							High: '9',
						},
					},
				}),
			),
		)
		expressionNfa := nfa.FromRegex(expression, 0)
		b.ResetTimer()
		for range b.N {
			_ = dfa.FromNFA(expressionNfa)
		}
	})

	b.Run("a", func(b *testing.B) {
		expression := frontend.NewNodeLiteral("a")
		expressionNfa := nfa.FromRegex(expression, 0)
		b.ResetTimer()
		for range b.N {
			_ = dfa.FromNFA(expressionNfa)
		}
	})

	b.Run("ab", func(b *testing.B) {
		expression := frontend.NewNodeConcat(
			frontend.NewNodeLiteral("a"),
			frontend.NewNodeLiteral("b"),
		)
		expressionNfa := nfa.FromRegex(expression, 0)
		b.ResetTimer()
		for range b.N {
			_ = dfa.FromNFA(expressionNfa)
		}
	})

	b.Run("a|b", func(b *testing.B) {
		expression := frontend.NewNodeOr(
			frontend.NewNodeLiteral("a"),
			frontend.NewNodeLiteral("b"),
		)
		expressionNfa := nfa.FromRegex(expression, 0)
		b.ResetTimer()
		for range b.N {
			_ = dfa.FromNFA(expressionNfa)
		}
	})

	b.Run("a*", func(b *testing.B) {
		expression := frontend.NewNodeZeroOrMore(
			frontend.NewNodeLiteral("a"),
		)
		expressionNfa := nfa.FromRegex(expression, 0)
		b.ResetTimer()
		for range b.N {
			_ = dfa.FromNFA(expressionNfa)
		}
	})

	b.Run("(b|c)*", func(b *testing.B) {
		expression := frontend.NewNodeZeroOrMore(
			frontend.NewNodeOr(
				frontend.NewNodeLiteral("b"),
				frontend.NewNodeLiteral("c"),
			),
		)
		expressionNfa := nfa.FromRegex(expression, 0)
		b.ResetTimer()
		for range b.N {
			_ = dfa.FromNFA(expressionNfa)
		}
	})

	b.Run("a(b|c)*", func(b *testing.B) {
		expression := frontend.NewNodeConcat(
			frontend.NewNodeLiteral("a"),
			frontend.NewNodeZeroOrMore(
				frontend.NewNodeOr(
					frontend.NewNodeLiteral("b"),
					frontend.NewNodeLiteral("c"),
				),
			),
		)
		expressionNfa := nfa.FromRegex(expression, 0)
		b.ResetTimer()
		for range b.N {
			_ = dfa.FromNFA(expressionNfa)
		}
	})
}
