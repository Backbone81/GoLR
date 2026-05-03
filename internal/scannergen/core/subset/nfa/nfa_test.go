package nfa_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	thompsonsnfa "golr/internal/scannergen/core/subset/nfa"
	"golr/internal/scannergen/frontend"
	"golr/internal/scannergen/frontend/dsl"
)

var _ = Describe("NFA", func() {
	It("should build the correct NFA for regex 'a'", func() {
		expression := dsl.Literal("a")
		expressionNFA := thompsonsnfa.FromRegex(expression, 0)

		wantNfa := []thompsonsnfa.State{
			{ // state 0
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
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

		Expect(expressionNFA).To(Equal(wantNfa))
	})

	It("should build the correct NFA for regex 'b'", func() {
		expression := dsl.Literal("b")
		expressionNFA := thompsonsnfa.FromRegex(expression, 0)

		wantNfa := []thompsonsnfa.State{
			{ // state 0
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  'b',
							High: 'b',
						},
						NextStateIdx: 1,
					},
				},
			},
			{ // state 1
				Accept: true,
			},
		}

		Expect(expressionNFA).To(Equal(wantNfa))
	})

	It("should build the correct NFA for regex 'ab'", func() {
		expression := dsl.Concat(
			dsl.Literal("a"),
			dsl.Literal("b"),
		)
		expressionNFA := thompsonsnfa.FromRegex(expression, 0)

		wantNfa := []thompsonsnfa.State{
			{ // state 0
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  'a',
							High: 'a',
						},
						NextStateIdx: 1,
					},
				},
			},
			{ // state 1
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 2,
					},
				},
			},
			{ // state 2
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  'b',
							High: 'b',
						},
						NextStateIdx: 3,
					},
				},
			},
			{ // state 3
				Accept: true,
			},
		}

		Expect(expressionNFA).To(Equal(wantNfa))
	})

	It("should build the correct NFA for regex 'a|b'", func() {
		expression := dsl.Or(
			dsl.Literal("a"),
			dsl.Literal("b"),
		)
		expressionNFA := thompsonsnfa.FromRegex(expression, 0)

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
						CharRange: frontend.CharRange{
							Low:  'a',
							High: 'a',
						},
						NextStateIdx: 2,
					},
				},
			},
			{ // state 2
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 5,
					},
				},
			},
			{ // state 3
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  'b',
							High: 'b',
						},
						NextStateIdx: 4,
					},
				},
			},
			{ // state 4
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 5,
					},
				},
			},
			{ // state 5
				Accept: true,
			},
		}

		Expect(expressionNFA).To(Equal(wantNfa))
	})

	It("should build the correct NFA for regex 'a*'", func() {
		expression := dsl.ZeroOrMore(
			dsl.Literal("a"),
		)
		expressionNFA := thompsonsnfa.FromRegex(expression, 0)

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
						CharRange: frontend.CharRange{
							Low:  'a',
							High: 'a',
						},
						NextStateIdx: 2,
					},
				},
			},
			{ // state 2
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
			{ // state 3
				Accept: true,
			},
		}

		Expect(expressionNFA).To(Equal(wantNfa))
	})

	It("should build the correct NFA for regex '(b|c)*'", func() {
		expression := dsl.ZeroOrMore(
			dsl.Or(
				dsl.Literal("b"),
				dsl.Literal("c"),
			),
		)
		expressionNFA := thompsonsnfa.FromRegex(expression, 0)

		wantNfa := []thompsonsnfa.State{
			{ // state 0
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 1,
					},
					{
						Empty:        true,
						NextStateIdx: 7,
					},
				},
			},
			{ // state 1
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 2,
					},
					{
						Empty:        true,
						NextStateIdx: 4,
					},
				},
			},
			{ // state 2
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  'b',
							High: 'b',
						},
						NextStateIdx: 3,
					},
				},
			},
			{ // state 3
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 6,
					},
				},
			},
			{ // state 4
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  'c',
							High: 'c',
						},
						NextStateIdx: 5,
					},
				},
			},
			{ // state 5
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 6,
					},
				},
			},
			{ // state 6
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 1,
					},
					{
						Empty:        true,
						NextStateIdx: 7,
					},
				},
			},
			{ // state 7
				Accept: true,
			},
		}

		Expect(expressionNFA).To(Equal(wantNfa))
	})

	It("should build the correct NFA for regex 'a(b|c)*'", func() {
		expression := dsl.Concat(
			dsl.Literal("a"),
			dsl.ZeroOrMore(
				dsl.Or(
					dsl.Literal("b"),
					dsl.Literal("c"),
				),
			),
		)
		expressionNFA := thompsonsnfa.FromRegex(expression, 0)

		wantNfa := []thompsonsnfa.State{
			{ // state 0
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  'a',
							High: 'a',
						},
						NextStateIdx: 1,
					},
				},
			},
			{ // state 1
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 2,
					},
				},
			},
			{ // state 2
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 3,
					},
					{
						Empty:        true,
						NextStateIdx: 9,
					},
				},
			},
			{ // state 3
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 4,
					},
					{
						Empty:        true,
						NextStateIdx: 6,
					},
				},
			},
			{ // state 4
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  'b',
							High: 'b',
						},
						NextStateIdx: 5,
					},
				},
			},
			{ // state 5
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 8,
					},
				},
			},
			{ // state 6
				Transitions: []thompsonsnfa.Transition{
					{
						CharRange: frontend.CharRange{
							Low:  'c',
							High: 'c',
						},
						NextStateIdx: 7,
					},
				},
			},
			{ // state 7
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 8,
					},
				},
			},
			{ // state 8
				Transitions: []thompsonsnfa.Transition{
					{
						Empty:        true,
						NextStateIdx: 3,
					},
					{
						Empty:        true,
						NextStateIdx: 9,
					},
				},
			},
			{ // state 9
				Accept: true,
			},
		}
		Expect(expressionNFA).To(Equal(wantNfa))
	})
})

//nolint:funlen,gocognit,cyclop
func BenchmarkFromRegex(b *testing.B) {
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
		for range b.N {
			_ = thompsonsnfa.FromRegex(expression, 0)
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
		for range b.N {
			_ = thompsonsnfa.FromRegex(expression, 0)
		}
	})

	b.Run("a", func(b *testing.B) {
		expression := dsl.Literal("a")
		for range b.N {
			_ = thompsonsnfa.FromRegex(expression, 0)
		}
	})

	b.Run("ab", func(b *testing.B) {
		expression := dsl.Concat(
			dsl.Literal("a"),
			dsl.Literal("b"),
		)
		for range b.N {
			_ = thompsonsnfa.FromRegex(expression, 0)
		}
	})

	b.Run("a|b", func(b *testing.B) {
		expression := dsl.Or(
			dsl.Literal("a"),
			dsl.Literal("b"),
		)
		for range b.N {
			_ = thompsonsnfa.FromRegex(expression, 0)
		}
	})

	b.Run("a*", func(b *testing.B) {
		expression := dsl.ZeroOrMore(
			dsl.Literal("a"),
		)
		for range b.N {
			_ = thompsonsnfa.FromRegex(expression, 0)
		}
	})

	b.Run("(b|c)*", func(b *testing.B) {
		expression := dsl.ZeroOrMore(
			dsl.Or(
				dsl.Literal("b"),
				dsl.Literal("c"),
			),
		)
		for range b.N {
			_ = thompsonsnfa.FromRegex(expression, 0)
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
		for range b.N {
			_ = thompsonsnfa.FromRegex(expression, 0)
		}
	})
}
