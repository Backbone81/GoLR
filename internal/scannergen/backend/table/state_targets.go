package table

import (
	"github.com/backbone81/golr/internal/scannergen/backend"
)

// NoTransition is the entry used for a byte or byte class a state has no transition on. State indices are never
// negative, so this can not be confused with a transition to state 0.
const NoTransition = -1

// ByteValueCount is the number of distinct values a byte can take, and therefore the number of columns a transition
// row would need if it were indexed by the input byte itself.
const ByteValueCount = 256

// fillTargets fills targets with the state each byte transitions to in the given state, using NoTransition for the
// bytes the state has no transition on. This expands the byte ranges of a state into one entry per byte, which is the
// form both the equivalence classes and the transition rows are derived from.
func fillTargets(targets *[ByteValueCount]int, state *backend.State) {
	for byteValue := range targets {
		targets[byteValue] = NoTransition
	}
	for _, transition := range state.Transitions {
		// The loop counter must be wider than a byte, because a byte range ending at 0xFF would otherwise wrap
		// around instead of terminating the loop.
		for byteValue := int(transition.ByteRange.Low); byteValue <= int(transition.ByteRange.High); byteValue++ {
			targets[byteValue] = transition.StateIdx
		}
	}
}
