package frontend

import (
	"errors"
	"fmt"
	"strings"
)

// Repetition is a regular expression matching its child for a specific number of times.
// The child needs to implement the [Node] interface.
// Minimum must always be smaller or equal to Maximum.
// Set Minimum and Maximum to the same value to have an exact number of repetitions.
type Repetition struct {
	Minimum int
	Maximum int
	Child   *Node
}

// String returns a string representation of this regular expression.
func (r *Repetition) String() string {
	var result strings.Builder
	if !r.Child.IsSingleNode() {
		result.WriteString("(")
	}
	result.WriteString(r.Child.String())
	if !r.Child.IsSingleNode() {
		result.WriteString(")")
	}
	result.WriteString(fmt.Sprintf("{%d", r.Minimum))
	if r.Minimum != r.Maximum {
		result.WriteString(fmt.Sprintf(",%d", r.Maximum))
	}
	result.WriteString("}")
	return result.String()
}

// IsSingleNode reports if this regular expression can be represented as a single node when converted to a string.
// Regular expressions which cannot be represented as a single node need to have parenthesis placed around them to
// form a subexpression.
func (r *Repetition) IsSingleNode() bool {
	return false
}

// Validate reports if the regular expression satisfies the required conditions to be considered valid.
// A nil return value indicates that the regular expression is valid.
// An error return value provides details about the unmet condition.
// In situations where the regular expression has children, all children are checked for validity recursively. If
// any child is not valid, this regular expression is also not valid.
func (r *Repetition) Validate() error {
	if r.Child == nil {
		return errors.New("the regular expression requires a child")
	}
	if r.Minimum < 0 {
		return errors.New("the minimum must be non-negative")
	}
	if r.Maximum < 0 {
		return errors.New("the maximum must be non-negative")
	}
	if r.Minimum > r.Maximum {
		return errors.New("minimum must always be smaller or equal to maximum")
	}
	if r.Minimum == 0 && r.Maximum == 0 {
		return errors.New("minimum and maximum must not be both zero")
	}
	return r.Child.Validate()
}
