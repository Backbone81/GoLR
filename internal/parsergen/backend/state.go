package backend

import (
	"fmt"
	"strings"
)

// State represents a LR(1) state. The structure of this state is derived from definition 3.1 of IELR(1).
type State struct {
	// KernelItems provides a set of all kernel items.
	KernelItems CoreSet `json:"kernelItems" yaml:"kernelItems"`

	// TransitionActions provides a set of all transition actions.
	TransitionActions TransitionActionSet `json:"transitionActions" yaml:"transitionActions"`

	// ReduceActions provides a set of all reduce actions.
	ReduceActions ReduceActionSet `json:"reduceActions" yaml:"reduceActions"`

	// DefaultReduceProductionIdx provides the production index for a default reduce action on any lookahead. Is nil
	// if not set.
	//nolint:lll // Go tag lines cannot be broken onto multiple lines.
	DefaultReduceProductionIdx *int `json:"defaultReduceProductionIdx,omitempty" yaml:"defaultReduceProductionIdx,omitempty"`
}

// State implements fmt.Stringer.
var _ fmt.Stringer = (*State)(nil)

// String returns a string representation.
func (s *State) String() string {
	var builder strings.Builder

	builder.WriteString("\tkernel items: ")
	builder.WriteString(s.KernelItems.String())
	builder.WriteString("\n")

	builder.WriteString("\ttransition actions: ")
	builder.WriteString(s.TransitionActions.String())
	builder.WriteString("\n")

	builder.WriteString("\treduce actions: ")
	builder.WriteString(s.ReduceActions.String())
	builder.WriteString("\n")

	return builder.String()
}
