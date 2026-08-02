package table_test

import (
	"fmt"
	"unicode"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/scannergen/backend"
	"github.com/backbone81/golr/internal/scannergen/backend/table"
	"github.com/backbone81/golr/internal/scannergen/frontend/dsl"
)

// expectCompressedDecodeEquivalence asserts that for every state and every possible input byte the compressed lookup
// returns exactly what reading the DFA directly returns, and that the accept table reports the same rules the DFA
// states do. This covers both compression steps at once, because the lookup goes through the byte classes and the row
// displacement together.
func expectCompressedDecodeEquivalence(dfa backend.DFA) table.CompressedDFA {
	compressed := table.NewCompressedDFA(dfa)

	Expect(compressed.StateCount()).To(Equal(len(dfa.States)))
	for stateIdx := range dfa.States {
		expectedRuleIdx := table.NoRule
		if dfa.States[stateIdx].Accept {
			expectedRuleIdx = dfa.States[stateIdx].RuleIdx
		}
		Expect(compressed.AcceptRuleIdxByStateIdx[stateIdx]).To(
			Equal(expectedRuleIdx),
			fmt.Sprintf("state %d accepting", stateIdx),
		)

		for byteValue := range table.ByteValueCount {
			Expect(compressed.Transition(stateIdx, byte(byteValue))).To(
				Equal(transitionTarget(dfa.States[stateIdx], byteValue)),
				fmt.Sprintf("state %d on byte 0x%02X", stateIdx, byteValue),
			)
		}
	}
	return compressed
}

var _ = Describe("CompressedDFA", func() {
	It("decodes like the DFA for a scanner over plain ASCII", func() {
		expectCompressedDecodeEquivalence(rulesToDFA(
			dsl.Rule("identifier", dsl.OneOrMore(dsl.CharClass(dsl.CharRange('a', 'z'), dsl.CharRange('A', 'Z')))),
			dsl.Rule("number", dsl.OneOrMore(dsl.CharClass(dsl.CharRange('0', '9')))),
			dsl.Rule("plus", dsl.Literal("+")),
			dsl.Rule("arrow", dsl.Literal("->")),
			dsl.SkipRule("whitespace", dsl.OneOrMore(dsl.CharClass(dsl.CharRange(' ', ' '), dsl.CharRange('\t', '\n')))),
		))
	})

	It("decodes like the DFA for a scanner over the full Unicode range", func() {
		expectCompressedDecodeEquivalence(rulesToDFA(
			dsl.Rule("string", dsl.Concat(
				dsl.Literal(`"`),
				dsl.ZeroOrMore(dsl.NegCharClass(dsl.CharRange('"', '"'), dsl.CharRange('\\', '\\'))),
				dsl.Literal(`"`),
			)),
			dsl.Rule("letter", dsl.OneOrMore(dsl.CharClass(dsl.UnicodeCategory(unicode.L)...))),
			dsl.Rule("any", dsl.Any()),
		))
	})

	It("decodes like the DFA for the GoLR specification", func() {
		expectCompressedDecodeEquivalence(golrSpecDFA())
	})

	It("needs far fewer cells than the uncompressed transition table", func() {
		dfa := golrSpecDFA()
		compressed := expectCompressedDecodeEquivalence(dfa)

		denseCellCount := len(dfa.States) * table.ByteValueCount
		AddReportEntry("cells", fmt.Sprintf(
			"%d states, %d byte classes, %d cells packed, %d cells for a table indexed by byte",
			len(dfa.States), compressed.ByteClasses.Count(), len(compressed.Transitions.Next), denseCellCount,
		))
		Expect(len(compressed.Transitions.Next)).To(BeNumerically("<", denseCellCount/10))
	})

	It("supports a DFA without any state", func() {
		compressed := table.NewCompressedDFA(backend.DFA{})
		Expect(compressed.StateCount()).To(Equal(0))
		Expect(compressed.Transitions.Base).To(BeEmpty())
	})
})
