package nfa

import "github.com/backbone81/golr/internal/scannergen/frontend"

func (b *ThompsonsConstruction) fromOr(regexNode *frontend.Or, ruleIdx int, states []State) []State {
	// create the start state and remember the index for later
	states = append(states, State{
		RuleIdx: ruleIdx,
	})
	startStateIdx := len(states) - 1

	acceptingStateIdxs := make([]int, len(regexNode.Children))
	for i, child := range regexNode.Children {
		// add the transition from our start state to the child start state which is written next
		states[startStateIdx].Transitions = append(states[startStateIdx].Transitions, Transition{
			Empty:        true,
			NextStateIdx: len(states),
		})
		states = b.buildNFAFromRegexValidated(child, ruleIdx, states)

		// record the child accepting state for later when we know where our final accepting state lands at
		acceptingStateIdxs[i] = len(states) - 1
	}

	// create the accepting state
	states = append(states, State{
		RuleIdx: ruleIdx,
		Accept:  true,
	})

	// disable the child accepting states and add a transition to the accepting state
	for _, acceptingStateIdx := range acceptingStateIdxs {
		states[acceptingStateIdx].Accept = false
		states[acceptingStateIdx].Transitions = append(states[acceptingStateIdx].Transitions, Transition{
			Empty:        true,
			NextStateIdx: len(states) - 1,
		})
	}
	return states
}
