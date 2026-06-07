package backend

// State is a single DFA state.
type State struct {
	// RuleIdx is the index for the rule this state is part of. As the state of a DFA can be part of multiple rules
	// at the same time, this is the rule which has the lowest index and therefore the highest priority.
	RuleIdx int `json:"ruleIdx" yaml:"ruleIdx"`

	// Accept reports if this state is an accepting state for the rule given with RuleIdx.
	Accept bool `json:"accept,omitempty" yaml:"accept,omitempty"`

	// Transitions are the transitions to other DFA states.
	Transitions []Transition `json:"transitions,omitempty" yaml:"transitions,omitempty"`
}

func (s *State) GetTransition(byteRange ByteRange) *Transition {
	for i := range s.Transitions {
		if s.Transitions[i].ByteRange.Low == byteRange.Low && s.Transitions[i].ByteRange.High == byteRange.High {
			return &s.Transitions[i]
		}
	}
	return nil
}
