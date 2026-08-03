package table

import (
	"github.com/backbone81/golr/internal/parsergen/backend"
)

// NewDefaultActions returns the action every state takes for a terminal which its row of the action table has no entry
// for.
//
// This is where the default reductions of backend.ApplyDefaultReductions arrive. A state which reduces by the same
// production on most of its lookaheads carries that reduction once, as its default action, instead of once per
// lookahead, which is what leaves the rows of the action table sparse enough for the row displacement to pack them
// tightly.
//
// The accept arrives here as well, because it is the action of its state for any lookahead. The cores encode it
// differently: the Bison backed ones report it as `$default accept`, which reaches us as
// backend.State.DefaultReduceProductionIdx, while the native GoLR ones keep it a reduce of the accept production with
// the empty reduction lookahead set the DeRemer-Pennello computation yields for it. Both forms are recognized here, the
// same way the directly coded backend renders both into the default arm of the state.
//
// A state which has neither gets NoAction, which makes every terminal its row has no entry for a syntax error.
func NewDefaultActions(parser backend.Parser) []Action {
	result := make([]Action, len(parser.States))
	for stateIdx := range parser.States {
		result[stateIdx] = defaultAction(&parser.States[stateIdx])
	}
	return result
}

// defaultAction returns the action a single state takes for any terminal it has no action of its own for.
func defaultAction(state *backend.State) Action {
	if state.DefaultReduceProductionIdx != nil {
		return NewReduceAction(*state.DefaultReduceProductionIdx)
	}
	for _, reduceAction := range state.ReduceActions.All() {
		if reduceAction.ProductionIdx == acceptProductionIdx {
			return NewAcceptAction()
		}
	}
	return NoAction
}
