package backend

import (
	"fmt"
	"golr/internal/parsergen/frontend"
	"golr/internal/utils"
	"slices"

	"github.com/goccy/go-yaml"
)

// TransitionAction is a transition action of an LR(1) item consisting of a symbol index representing a terminal or
// nonterminal and a state index for the target state. The values for the symbol index and the state index must be in
// the range of [0, 65535]. The symbol index is the index of a terminal or the index of a nonterminal added to the index
// of the last terminal + 1.
//
// It is implemented as a single unsigned integer to allow for a more compact representation and to enable easy
// sorting when dealing with a slice of TransitionAction.
type TransitionAction uint32

// NewTransitionAction creates a new transition action with the given symbol index and the target state index.
func NewTransitionAction(symbolRef frontend.SymbolRef, stateIdx int) TransitionAction {
	utils.AssertValidIndex(stateIdx, transitionActionMaxState)
	// NOTE: We want to have the symbol index in the upper half of the TransitionAction and the state index in the lower
	// half. That way we automatically get a sensible order when sorting by the value of the TransitionAction (i.e.
	// first by symbol and second by state).
	//nolint:gosec // no integer overflow on correct usage
	return TransitionAction(symbolRef)<<transitionActionStateBits | TransitionAction(stateIdx)
}

const (
	transitionActionSymbolBits = 16
	transitionActionMaxSymbol  = (1 << transitionActionSymbolBits) - 1

	transitionActionStateBits = 16
	transitionActionMaxState  = (1 << transitionActionStateBits) - 1
	transitionActionStateMask = transitionActionMaxState
)

// SymbolRef returns the symbol index of the TransitionAction.
func (a TransitionAction) SymbolRef() frontend.SymbolRef {
	return frontend.SymbolRef(a >> transitionActionStateBits)
}

// StateIdx returns the state index of the TransitionAction.
func (a TransitionAction) StateIdx() int {
	return int(a & transitionActionStateMask)
}

// TransitionAction implements fmt.Stringer
var _ fmt.Stringer = (*TransitionAction)(nil)

// String returns a string representation.
func (a TransitionAction) String() string {
	return fmt.Sprintf("(symbol %d, state %d)", a.SymbolRef(), a.StateIdx())
}

// transitionActionMarshal is a helper struct which is only used for marshaling
type transitionActionMarshal struct {
	SymbolRef frontend.SymbolRef `json:"symbolRef" yaml:"symbol_ref"`
	StateIdx  int                `json:"stateIdx" yaml:"state_idx"`
}

// MarshalYAML implements the yaml.Marshaler interface.
func (a TransitionAction) MarshalYAML() ([]byte, error) {
	repr := transitionActionMarshal{
		SymbolRef: a.SymbolRef(),
		StateIdx:  a.StateIdx(),
	}
	return yaml.Marshal(repr)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (a *TransitionAction) UnmarshalYAML(b []byte) error {
	if slices.Equal(b, []byte("null")) {
		return nil
	}
	var repr transitionActionMarshal
	err := yaml.Unmarshal(b, &repr)
	if err != nil {
		return err
	}
	*a = NewTransitionAction(repr.SymbolRef, repr.StateIdx)
	return nil
}
