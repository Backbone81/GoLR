package table

import (
	"fmt"

	"github.com/backbone81/golr/internal/utils"
)

// Action is a single entry of the action table: what the parser does when the scanner delivers a terminal while the
// parser is in a state. It is a single integer, because that is the form the row displacement packs and the form the
// generated driver reads with a single lookup.
//
// The kind sits in the low bits of the integer and the value it carries above them. That keeps the magnitude of an
// action proportional to the number of states and productions of the grammar, which is what lets a backend pick a
// narrow integer type for the emitted table. Putting the kind into fixed high bits instead would push every reduce
// action past the range of a 16 bit integer, however small the grammar is.
type Action int

// NoAction is the entry for a terminal which a state has no action of its own for. The parser takes the default action
// of the state then, see NewDefaultActions. Encoded actions are never negative, so this can not be confused with one.
//
// This is not the same as the action NewErrorAction returns. An absent entry falls through to the default action of the
// state, an error action beats it.
const NoAction Action = -1

const (
	// ActionKindBits is the number of low bits of an Action which hold its ActionKind. A generated parser shifts an
	// action down by this much to get at the value it carries, which is why it is exported.
	ActionKindBits = 2

	// ActionKindMask selects the ActionKind out of an Action. A generated parser masks an action with it to find out
	// what to do, which is why it is exported.
	ActionKindMask = (1 << ActionKindBits) - 1

	// actionValueMax is the largest value an Action can carry next to its kind. Both values it ever carries are
	// capped at 16 bits by the parser tables themselves, the state index by backend.TransitionActionMaxState and the
	// production index by backend.NewReduceAction.
	actionValueMax = (1 << 16) - 1

	// acceptProductionIdx is the production index of `$accept -> Start $end` in the augmented grammar. Reducing by it
	// means the parse is finished, which is encoded as ActionKindAccept instead of as a reduce.
	acceptProductionIdx = 0
)

// ActionKind is what an Action does.
type ActionKind int

const (
	// ActionKindShift shifts the terminal and continues in the state the Action carries.
	ActionKindShift ActionKind = iota

	// ActionKindReduce reduces by the production the Action carries.
	ActionKindReduce

	// ActionKindAccept ends the parse successfully, because the input has been reduced to the start symbol.
	ActionKindAccept

	// ActionKindError rejects the terminal as a syntax error. It is only ever written for a terminal which the state
	// rejects on purpose, see backend.State.RejectedTerminals, because that rejection has to survive the default
	// action of the state. Every other syntax error is an entry the action table does not have at all and a default
	// action which does not cover it either.
	ActionKindError
)

// NewShiftAction returns the action which shifts the terminal and continues in the given state.
func NewShiftAction(stateIdx int) Action {
	utils.AssertValidIndex(stateIdx, actionValueMax)
	return Action(stateIdx)<<ActionKindBits | Action(ActionKindShift)
}

// NewReduceAction returns the action which reduces by the given production.
//
// Reducing by the accept production of the augmented grammar finishes the parse instead of pushing a nonterminal, so
// the accept action comes back for it. Encoding the accept here, in the one place which turns a production index into
// an action, keeps every caller from having to know about it.
func NewReduceAction(productionIdx int) Action {
	utils.AssertValidIndex(productionIdx, actionValueMax)
	if productionIdx == acceptProductionIdx {
		return NewAcceptAction()
	}
	return Action(productionIdx)<<ActionKindBits | Action(ActionKindReduce)
}

// NewAcceptAction returns the action which ends the parse successfully. It carries no value next to its kind, so
// nothing is shifted in above it.
func NewAcceptAction() Action {
	return Action(ActionKindAccept)
}

// NewErrorAction returns the action which rejects the terminal. It carries no value next to its kind, so nothing is
// shifted in above it.
func NewErrorAction() Action {
	return Action(ActionKindError)
}

// Kind returns what the action does.
func (a Action) Kind() ActionKind {
	return ActionKind(a & ActionKindMask)
}

// StateIdx returns the state a shift action continues in.
func (a Action) StateIdx() int {
	utils.DebugAssert(func() error {
		if a.Kind() != ActionKindShift {
			return fmt.Errorf("action %s carries no state index", a)
		}
		return nil
	})
	return int(a >> ActionKindBits)
}

// ProductionIdx returns the production a reduce action reduces by.
func (a Action) ProductionIdx() int {
	utils.DebugAssert(func() error {
		if a.Kind() != ActionKindReduce {
			return fmt.Errorf("action %s carries no production index", a)
		}
		return nil
	})
	return int(a >> ActionKindBits)
}

// Action implements fmt.Stringer.
var _ fmt.Stringer = (*Action)(nil)

// String returns a string representation.
func (a Action) String() string {
	if a == NoAction {
		return "(no action)"
	}
	switch a.Kind() {
	case ActionKindShift:
		return fmt.Sprintf("(shift, state %d)", a.StateIdx())
	case ActionKindReduce:
		return fmt.Sprintf("(reduce, production %d)", a.ProductionIdx())
	case ActionKindAccept:
		return "(accept)"
	case ActionKindError:
		return "(error)"
	default:
		return fmt.Sprintf("(unknown action %d)", int(a))
	}
}
