package nfa

import "github.com/backbone81/golr/internal/scannergen/frontend"

func (b *ThompsonsConstruction) fromOptional(regexNode *frontend.Optional, ruleIdx int, states []State) []State {
	startStateIdx := len(states)
	states = b.buildNFAFromRegexValidated(regexNode.Child, ruleIdx, states)
	states[startStateIdx].Transitions = append(states[startStateIdx].Transitions, Transition{
		Empty:        true,
		NextStateIdx: len(states) - 1,
	})
	return states
}
