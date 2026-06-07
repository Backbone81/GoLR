package nfa

import (
	"github.com/backbone81/golr/internal/scannergen/frontend"
)

func (b *ThompsonsConstruction) fromCharClass(regexNode *frontend.CharClass, ruleIdx int, states []State) []State {
	states = append(states,
		// start state
		State{
			RuleIdx: ruleIdx,
		},
	)
	startStateIdx := len(states) - 1

	// invert the character ranges if needed
	characterRanges := regexNode.Ranges
	if regexNode.Negate {
		negatedRanges := frontend.NegateCharRanges(characterRanges)
		characterRanges = negatedRanges
	}

	// Make sure that all character ranges are within the same length of bytes when UTF-8 encoded.
	characterRanges = frontend.SplitCharRanges(characterRanges, MaxUTF8Rune1Byte+1)
	characterRanges = frontend.SplitCharRanges(characterRanges, MaxUTF8Rune2Bytes+1)
	characterRanges = frontend.SplitCharRanges(characterRanges, MaxUTF8Rune3Bytes+1)

	// add transitions from the start state to the accepting state
	for _, characterRange := range characterRanges {
		states = BuildUTF8Encoding(characterRange, ruleIdx, states, startStateIdx)
	}

	// Add accepting state and fix the transitions to accepting state
	states = append(states,
		State{RuleIdx: ruleIdx, Accept: true},
	)
	acceptStateIdx := len(states) - 1

	for i := range states {
		for j := range states[i].Transitions {
			if states[i].Transitions[j].NextStateIdx == -1 {
				states[i].Transitions[j].NextStateIdx = acceptStateIdx
			}
		}
	}
	return states
}
