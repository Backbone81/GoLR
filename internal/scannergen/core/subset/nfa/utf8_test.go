package nfa_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/scannergen/core/subset/nfa"
	"github.com/backbone81/golr/internal/scannergen/frontend"
)

func buildNFAForRange(charRange frontend.CharRange) []nfa.State {
	states := []nfa.State{
		{
			RuleIdx: 0,
		},
	}
	return nfa.BuildUTF8Encoding(charRange, 0, states, 0)
}

var _ = Describe("BuildUTF8Encoding", func() {
	Context("1-byte ASCII", func() {
		It("produces a direct transition for a single ASCII character", func() {
			Expect(buildNFAForRange(frontend.CharRange{Low: 'A', High: 'A'})).To(Equal([]nfa.State{
				{ // state 0: start
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 'A', High: 'A'},
							NextStateIdx: -1,
						},
					},
				},
			}))
		})

		It("produces a direct transition for an ASCII range", func() {
			Expect(buildNFAForRange(frontend.CharRange{Low: 'a', High: 'z'})).To(Equal([]nfa.State{
				{ // state 0: start
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 'a', High: 'z'},
							NextStateIdx: -1,
						},
					},
				},
			}))
		})
	})

	Context("2-byte UTF-8", func() {
		It("produces a chain of two fixed transitions for a single 2-byte character", func() {
			// U+00E9 'é' encodes as 0xC3 0xA9
			Expect(buildNFAForRange(frontend.CharRange{Low: 0xE9, High: 0xE9})).To(Equal([]nfa.State{
				{ // state 0: start
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0xC3, High: 0xC3},
							NextStateIdx: 1,
						},
					},
				},
				{ // state 1: after leading byte 0xC3
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0xA9, High: 0xA9},
							NextStateIdx: -1,
						},
					},
				},
			}))
		})

		It("produces one intermediate state when the whole range shares the same leading byte", func() {
			// U+00C0–U+00FF all encode to 0xC3 [0x80–0xBF]
			Expect(buildNFAForRange(frontend.CharRange{Low: 0xC0, High: 0xFF})).To(Equal([]nfa.State{
				{ // state 0: start
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0xC3, High: 0xC3},
							NextStateIdx: 1,
						},
					},
				},
				{ // state 1: after leading byte 0xC3
					RuleIdx: 0,
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0x80, High: 0xBF},
							NextStateIdx: -1,
						},
					},
				},
			}))
		})

		It("produces low and high paths with no middle when leading bytes are adjacent", func() {
			// U+0080–U+00FF: leading byte 0xC2 for U+0080–U+00BF, 0xC3 for U+00C0–U+00FF
			Expect(buildNFAForRange(frontend.CharRange{Low: 0x80, High: 0xFF})).To(Equal([]nfa.State{
				{ // state 0: start
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0xC2, High: 0xC2},
							NextStateIdx: 1,
						},
						{
							CharRange:    frontend.CharRange{Low: 0xC3, High: 0xC3},
							NextStateIdx: 2,
						},
					},
				},
				{ // state 1: after leading byte 0xC2 (low path)
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0x80, High: 0xBF},
							NextStateIdx: -1,
						},
					},
				},
				{ // state 2: after leading byte 0xC3 (high path)
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0x80, High: 0xBF},
							NextStateIdx: -1,
						},
					},
				},
			}))
		})

		It("produces low, middle, and high paths when leading bytes span more than two values", func() {
			// U+00A0–U+014F: 0xC2 (partial low), 0xC3–0xC4 (full middle), 0xC5 (partial high)
			// U+00A0 = 0xC2 0xA0, U+014F = 0xC5 0x8F
			Expect(buildNFAForRange(frontend.CharRange{Low: 0xA0, High: 0x14F})).To(Equal([]nfa.State{
				{ // state 0: start
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0xC2, High: 0xC2},
							NextStateIdx: 1,
						},
						{
							CharRange:    frontend.CharRange{Low: 0xC3, High: 0xC4},
							NextStateIdx: 2,
						},
						{
							CharRange:    frontend.CharRange{Low: 0xC5, High: 0xC5},
							NextStateIdx: 3,
						},
					},
				},
				{ // state 1: after 0xC2 (low path, partial continuation 0xA0–0xBF)
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0xA0, High: 0xBF},
							NextStateIdx: -1,
						},
					},
				},
				{ // state 2: after 0xC3–0xC4 (middle path, full continuation)
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0x80, High: 0xBF},
							NextStateIdx: -1,
						},
					},
				},
				{ // state 3: after 0xC5 (high path, partial continuation 0x80–0x8F)
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0x80, High: 0x8F},
							NextStateIdx: -1,
						},
					},
				},
			}))
		})
	})

	Context("3-byte UTF-8", func() {
		It("produces a chain of three transitions when all bytes are fixed except the last", func() {
			// U+0800–U+083F both encode to 0xE0 0xA0 [0x80–0xBF]
			Expect(buildNFAForRange(frontend.CharRange{Low: 0x800, High: 0x83F})).To(Equal([]nfa.State{
				{ // state 0: start
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0xE0, High: 0xE0},
							NextStateIdx: 1,
						},
					},
				},
				{ // state 1: after leading byte 0xE0
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0xA0, High: 0xA0},
							NextStateIdx: 2,
						},
					},
				},
				{ // state 2: after second byte 0xA0
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0x80, High: 0xBF},
							NextStateIdx: -1,
						},
					},
				},
			}))
		})

		It("produces a split on the second byte when the leading byte is fixed but the second byte varies", func() {
			// U+0800–U+0FFF: leading byte 0xE0 fixed, second byte varies from 0xA0 to 0xBF
			// 0x800 = 0xE0 0xA0 0x80, 0xFFF = 0xE0 0xBF 0xBF
			Expect(buildNFAForRange(frontend.CharRange{Low: 0x800, High: 0xFFF})).To(Equal([]nfa.State{
				{ // state 0: start
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0xE0, High: 0xE0},
							NextStateIdx: 1,
						},
					},
				},
				{ // state 1: after leading byte 0xE0, second byte splits into low/middle/high
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0xA0, High: 0xA0},
							NextStateIdx: 2,
						},
						{
							CharRange:    frontend.CharRange{Low: 0xA1, High: 0xBE},
							NextStateIdx: 3,
						},
						{
							CharRange:    frontend.CharRange{Low: 0xBF, High: 0xBF},
							NextStateIdx: 4,
						},
					},
				},
				{ // state 2: after second byte 0xA0 (low path)
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0x80, High: 0xBF},
							NextStateIdx: -1,
						},
					},
				},
				{ // state 3: after second byte 0xA1–0xBE (middle path)
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0x80, High: 0xBF},
							NextStateIdx: -1,
						},
					},
				},
				{ // state 4: after second byte 0xBF (high path)
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0x80, High: 0xBF},
							NextStateIdx: -1,
						},
					},
				},
			}))
		})
	})

	Context("4-byte UTF-8", func() {
		It("produces a chain of four fixed transitions for a single 4-byte character", func() {
			// U+1F600 😀 encodes as 0xF0 0x9F 0x98 0x80
			Expect(buildNFAForRange(frontend.CharRange{Low: 0x1F600, High: 0x1F600})).To(Equal([]nfa.State{
				{ // state 0: start
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0xF0, High: 0xF0},
							NextStateIdx: 1,
						},
					},
				},
				{ // state 1: after leading byte 0xF0
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0x9F, High: 0x9F},
							NextStateIdx: 2,
						},
					},
				},
				{ // state 2: after second byte 0x9F
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0x98, High: 0x98},
							NextStateIdx: 3,
						},
					},
				},
				{ // state 3: after third byte 0x98
					Transitions: []nfa.Transition{
						{
							CharRange:    frontend.CharRange{Low: 0x80, High: 0x80},
							NextStateIdx: -1,
						},
					},
				},
			}))
		})
	})
})
