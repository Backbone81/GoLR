package nfa

// Merge combines multiple NFAs into a single one.
func Merge(nfas ...[]State) []State {
	if len(nfas) < 1 {
		panic("at least one NFA required as parameter")
	}

	states := []State{
		{},
	}

	for _, nfa := range nfas {
		offset := len(states)

		// patch up transition destinations
		for stateIdx := range nfa {
			for transitionIdx := range nfa[stateIdx].Transitions {
				nfa[stateIdx].Transitions[transitionIdx].NextStateIdx += offset
			}
		}

		states[0].Transitions = append(states[0].Transitions, Transition{
			Empty:        true,
			NextStateIdx: len(states),
		})
		states = append(states, nfa...)
	}
	return states
}
