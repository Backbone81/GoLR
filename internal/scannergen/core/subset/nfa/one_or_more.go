package nfa

import "golr/internal/scannergen/frontend"

func (b *ThompsonsConstruction) fromOneOrMore(regexNode *frontend.OneOrMore, ruleIdx int, states []State) []State {
	startStateIdx := len(states)
	states = b.buildNFAFromRegexValidated(regexNode.Child, ruleIdx, states)
	states[len(states)-1].Transitions = append(states[len(states)-1].Transitions, Transition{
		Empty:        true,
		NextStateIdx: startStateIdx,
	})
	return states
}
