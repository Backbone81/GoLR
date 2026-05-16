package nfa

import "github.com/backbone81/golr/internal/scannergen/frontend"

func (b *ThompsonsConstruction) fromConcat(regexNode *frontend.Concat, ruleIdx int, states []State) []State {
	for i, child := range regexNode.Children {
		acceptingStateIdx := len(states) - 1
		states = b.buildNFAFromRegexValidated(child, ruleIdx, states)
		if i == 0 {
			// the first child does not have a predecessor which we can connect to
			continue
		}
		states[acceptingStateIdx].Accept = false
		states[acceptingStateIdx].Transitions = append(states[acceptingStateIdx].Transitions, Transition{
			Empty:        true,
			NextStateIdx: acceptingStateIdx + 1,
		})
	}
	return states
}
