package nfa

import "golr/internal/scannergen/frontend"

func (b *ThompsonsConstruction) fromCharClass(regexNode *frontend.CharClass, ruleIdx int, states []State) []State {
	states = append(states,
		// start state
		State{
			RuleIdx: ruleIdx,
		},
		// accepting state
		State{
			RuleIdx: ruleIdx,
			Accept:  true,
		},
	)
	startStateIdx := len(states) - 2

	// invert the character ranges if needed
	characterRanges := regexNode.Ranges
	if regexNode.Negate {
		negatedRanges := frontend.NegateCharRanges(characterRanges)
		characterRanges = negatedRanges
	}

	// add transitions from the start state to the accepting state
	for _, characterRange := range characterRanges {
		states[startStateIdx].Transitions = append(states[startStateIdx].Transitions, Transition{
			CharRange:    characterRange,
			NextStateIdx: startStateIdx + 1,
		})
	}
	return states
}
