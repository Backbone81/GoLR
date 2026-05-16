package nfa

import (
	"unicode"

	"github.com/backbone81/golr/internal/scannergen/frontend"
)

func (b *ThompsonsConstruction) fromAny(_ *frontend.Any, ruleIdx int, states []State) []State {
	states = append(states,
		// start state
		State{
			RuleIdx: ruleIdx,
			Transitions: []Transition{
				{
					CharRange: frontend.CharRange{
						Low:  0,
						High: unicode.MaxRune,
					},
					NextStateIdx: len(states) + 1,
				},
			},
		},
		// accepting state
		State{
			RuleIdx: ruleIdx,
			Accept:  true,
		},
	)
	return states
}
