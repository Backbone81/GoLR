package nfa

import (
	"github.com/backbone81/golr/internal/scannergen/frontend"
)

func (b *ThompsonsConstruction) fromAny(_ *frontend.Any, ruleIdx int, states []State) []State {
	continuation1ByteStateIdx := len(states) + 1  // expecting 1 more continuation byte
	continuation2BytesStateIdx := len(states) + 2 // expecting 2 more continuation bytes
	continuation3BytesStateIdx := len(states) + 3 // expecting 3 more continuation bytes
	acceptStateIdx := len(states) + 4

	states = append(states,
		State{ // start
			RuleIdx: ruleIdx,
			Transitions: []Transition{
				{
					CharRange:    frontend.CharRange{Low: 0x00, High: '\n' - 1},
					NextStateIdx: acceptStateIdx,
				},
				{
					CharRange:    frontend.CharRange{Low: '\n' + 1, High: 0x7F},
					NextStateIdx: acceptStateIdx,
				},
				{
					CharRange:    frontend.CharRange{Low: 0xC2, High: 0xDF},
					NextStateIdx: continuation1ByteStateIdx,
				},
				{
					CharRange:    frontend.CharRange{Low: 0xE0, High: 0xEF},
					NextStateIdx: continuation2BytesStateIdx,
				},
				{
					CharRange:    frontend.CharRange{Low: 0xF0, High: 0xF4},
					NextStateIdx: continuation3BytesStateIdx,
				},
			},
		},
		State{ // 1 continuation byte
			RuleIdx: ruleIdx,
			Transitions: []Transition{
				{
					CharRange:    frontend.CharRange{Low: 0x80, High: 0xBF},
					NextStateIdx: acceptStateIdx,
				},
			},
		},
		State{ // 2 continuation bytes
			RuleIdx: ruleIdx,
			Transitions: []Transition{
				{
					CharRange:    frontend.CharRange{Low: 0x80, High: 0xBF},
					NextStateIdx: continuation1ByteStateIdx,
				},
			},
		},
		State{ // 3 continuation bytes
			RuleIdx: ruleIdx,
			Transitions: []Transition{
				{
					CharRange:    frontend.CharRange{Low: 0x80, High: 0xBF},
					NextStateIdx: continuation2BytesStateIdx,
				},
			},
		},
		State{ // accept — always last
			RuleIdx: ruleIdx,
			Accept:  true,
		},
	)
	return states
}
