package nfa

import "github.com/backbone81/golr/internal/scannergen/frontend"

func (b *ThompsonsConstruction) fromZeroOrMore(regexNode *frontend.ZeroOrMore, ruleIdx int, states []State) []State {
	// create the start state and remember the index for later
	states = append(states, State{
		RuleIdx: ruleIdx,
	})
	startStateIdx := len(states) - 1

	// add the child regex
	states = b.buildNFAFromRegexValidated(regexNode.Child, ruleIdx, states)
	states[len(states)-1].Accept = false

	// transition from child accepting state to child start state
	states[len(states)-1].Transitions = append(states[len(states)-1].Transitions, Transition{
		Empty:        true,
		NextStateIdx: startStateIdx + 1,
	})

	// transition from start state to child start state
	states[startStateIdx].Transitions = append(states[startStateIdx].Transitions, Transition{
		Empty:        true,
		NextStateIdx: startStateIdx + 1,
	})

	// create accepting state
	states = append(states, State{
		RuleIdx: ruleIdx,
		Accept:  true,
	})

	// transition from start state to accepting state
	states[startStateIdx].Transitions = append(states[startStateIdx].Transitions, Transition{
		Empty:        true,
		NextStateIdx: len(states) - 1,
	})

	// transition from child accepting state to accepting state
	states[len(states)-2].Transitions = append(states[len(states)-2].Transitions, Transition{
		Empty:        true,
		NextStateIdx: len(states) - 1,
	})

	return states
}
