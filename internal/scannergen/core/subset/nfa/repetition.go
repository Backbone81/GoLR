package nfa

import "github.com/backbone81/golr/internal/scannergen/frontend"

func (b *ThompsonsConstruction) fromRepetition(regexNode *frontend.Repetition, ruleIdx int, states []State) []State {
	// create the start state
	states = append(states, State{
		RuleIdx: ruleIdx,
	})
	currStateIdx := len(states) - 1

	acceptingStateIdxs := make([]int, regexNode.Maximum-regexNode.Minimum+1)
	for i := range regexNode.Maximum {
		states = b.buildNFAFromRegexValidated(regexNode.Child, ruleIdx, states)
		states[len(states)-1].Accept = false
		if i+1 >= regexNode.Minimum {
			// record the child accepting state for later when we know where our final accepting state lands at
			acceptingStateIdxs[i+1-regexNode.Minimum] = len(states) - 1
		}
		states[currStateIdx].Transitions = append(states[currStateIdx].Transitions, Transition{
			Empty:        true,
			NextStateIdx: currStateIdx + 1,
		})
		currStateIdx = len(states) - 1
	}

	// create accepting state
	states = append(states, State{
		RuleIdx: ruleIdx,
		Accept:  true,
	})

	// add a transition to the accepting state
	for _, acceptingStateIdx := range acceptingStateIdxs {
		states[acceptingStateIdx].Transitions = append(states[acceptingStateIdx].Transitions, Transition{
			Empty:        true,
			NextStateIdx: len(states) - 1,
		})
	}

	return states
}
