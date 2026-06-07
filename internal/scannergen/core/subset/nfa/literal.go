package nfa

import "github.com/backbone81/golr/internal/scannergen/frontend"

func (b *ThompsonsConstruction) fromLiteral(regexNode *frontend.Literal, ruleIdx int, states []State) []State {
	// add the start state
	states = append(states, State{
		RuleIdx: ruleIdx,
	})
	currStateIdx := len(states) - 1

	for _, character := range []byte(regexNode.Text) {
		// add the next state
		states = append(states, State{
			RuleIdx: ruleIdx,
		})

		// add a transition from the previous to the next state
		states[currStateIdx].Transitions = []Transition{
			{
				CharRange: frontend.CharRange{
					Low:  rune(character),
					High: rune(character),
				},
				NextStateIdx: currStateIdx + 1,
			},
		}
		currStateIdx++
	}

	// mark the last state as an accepting state
	states[currStateIdx].Accept = true
	return states
}
