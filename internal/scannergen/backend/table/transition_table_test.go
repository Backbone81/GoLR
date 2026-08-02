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

// transitionTarget looks the transition of a state up, by finding the byte range which contains the byte. It is the
// reference the compressed lookup is compared against, and is deliberately written independently of the code under
// test.
func transitionTarget(state backend.State, byteValue int) int {
	for _, transition := range state.Transitions {
		if int(transition.ByteRange.Low) <= byteValue && byteValue <= int(transition.ByteRange.High) {
			return transition.StateIdx
		}
	}
	return table.NoTransition
}

// expectDecodeEquivalence asserts that for every state and every possible input byte the compressed lookup returns
// exactly what reading the DFA directly returns. This is the property the whole compression rests on, and it holds
// independently of how many classes the partition ended up with.
func expectDecodeEquivalence(dfa backend.DFA) {
	transitionTable := table.NewTransitionTable(dfa)

	Expect(transitionTable.Rows).To(HaveLen(len(dfa.States)))
	for stateIdx := range dfa.States {
		row := transitionTable.Rows[stateIdx]
		Expect(row).To(HaveLen(transitionTable.ByteClasses.Count()))

		for byteValue := range table.ByteValueCount {
			classIdx := transitionTable.ByteClasses.ClassByByte[byteValue]
			Expect(row[classIdx]).To(
				Equal(transitionTarget(dfa.States[stateIdx], byteValue)),
				fmt.Sprintf("state %d on byte 0x%02X in class %d", stateIdx, byteValue, classIdx),
			)
		}
	}
}

var _ = Describe("TransitionTable", func() {
	It("decodes like the DFA for a scanner over plain ASCII", func() {
		expectDecodeEquivalence(rulesToDFA(
			dsl.Rule("identifier", dsl.OneOrMore(dsl.CharClass(dsl.CharRange('a', 'z'), dsl.CharRange('A', 'Z')))),
			dsl.Rule("number", dsl.OneOrMore(dsl.CharClass(dsl.CharRange('0', '9')))),
			dsl.Rule("plus", dsl.Literal("+")),
			dsl.Rule("arrow", dsl.Literal("->")),
			dsl.SkipRule("whitespace", dsl.OneOrMore(dsl.CharClass(dsl.CharRange(' ', ' '), dsl.CharRange('\t', '\n')))),
		))
	})

	It("decodes like the DFA for a scanner over the full Unicode range", func() {
		// The multi byte UTF-8 states are what the byte classes are supposed to collapse, so the interesting case
		// is a scanner which accepts arbitrary runes.
		expectDecodeEquivalence(rulesToDFA(
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
		expectDecodeEquivalence(golrSpecDFA())
	})
})
