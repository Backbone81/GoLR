package dfa_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"golr/internal/scannergen/backend"
	"golr/internal/scannergen/core/subset/dfa"
	"golr/internal/scannergen/core/subset/nfa"
	"golr/internal/scannergen/frontend"
	"golr/internal/scannergen/frontend/dsl"
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

func BenchmarkFromNFA(b *testing.B) {
	b.Run("[a-zA-Z_][a-zA-Z0-9_]*", func(b *testing.B) {
		expression := dsl.Concat(
			dsl.CharClass(
				dsl.CharRange('a', 'z'),
				dsl.CharRange('A', 'Z'),
				dsl.CharRange('_', '_'),
			),
			dsl.ZeroOrMore(
				dsl.CharClass(
					dsl.CharRange('a', 'z'),
					dsl.CharRange('A', 'Z'),
					dsl.CharRange('0', '9'),
					dsl.CharRange('_', '_'),
				),
			),
		)
		expressionNfa := nfa.FromRegex(expression, 0)
		b.ResetTimer()
		for range b.N {
			_ = dfa.FromNFA(expressionNfa)
		}
	})

	b.Run("[-+]?[0-9]+", func(b *testing.B) {
		expression := dsl.Concat(
			dsl.Optional(
				dsl.CharClass(
					dsl.CharRange('-', '-'),
					dsl.CharRange('+', '+'),
				),
			),
			dsl.OneOrMore(
				dsl.CharClass(
					dsl.CharRange('0', '9'),
				),
			),
		)
		expressionNfa := nfa.FromRegex(expression, 0)
		b.ResetTimer()
		for range b.N {
			_ = dfa.FromNFA(expressionNfa)
		}
	})

	b.Run("a", func(b *testing.B) {
		expression := dsl.Literal("a")
		expressionNfa := nfa.FromRegex(expression, 0)
		b.ResetTimer()
		for range b.N {
			_ = dfa.FromNFA(expressionNfa)
		}
	})

	b.Run("ab", func(b *testing.B) {
		expression := dsl.Concat(
			dsl.Literal("a"),
			dsl.Literal("b"),
		)
		expressionNfa := nfa.FromRegex(expression, 0)
		b.ResetTimer()
		for range b.N {
			_ = dfa.FromNFA(expressionNfa)
		}
	})

	b.Run("a|b", func(b *testing.B) {
		expression := dsl.Or(
			dsl.Literal("a"),
			dsl.Literal("b"),
		)
		expressionNfa := nfa.FromRegex(expression, 0)
		b.ResetTimer()
		for range b.N {
			_ = dfa.FromNFA(expressionNfa)
		}
	})

	b.Run("a*", func(b *testing.B) {
		expression := dsl.ZeroOrMore(
			dsl.Literal("a"),
		)
		expressionNfa := nfa.FromRegex(expression, 0)
		b.ResetTimer()
		for range b.N {
			_ = dfa.FromNFA(expressionNfa)
		}
	})

	b.Run("(b|c)*", func(b *testing.B) {
		expression := dsl.ZeroOrMore(
			dsl.Or(
				dsl.Literal("b"),
				dsl.Literal("c"),
			),
		)
		expressionNfa := nfa.FromRegex(expression, 0)
		b.ResetTimer()
		for range b.N {
			_ = dfa.FromNFA(expressionNfa)
		}
	})

	b.Run("a(b|c)*", func(b *testing.B) {
		expression := dsl.Concat(
			dsl.Literal("a"),
			dsl.ZeroOrMore(
				dsl.Or(
					dsl.Literal("b"),
					dsl.Literal("c"),
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
