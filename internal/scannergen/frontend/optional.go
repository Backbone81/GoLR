package frontend

import (
	"errors"
	"strings"
)

// Optional is a regular expression matching zero or one instances of its child.
// The child needs to implement the [Node] interface.
type Optional struct {
	Child *Node
}

// String returns a string representation of this regular expression.
func (o *Optional) String() string {
	var result strings.Builder
	if !o.Child.IsSingleNode() {
		result.WriteString("(")
	}
	result.WriteString(o.Child.String())
	if !o.Child.IsSingleNode() {
		result.WriteString(")")
	}
	result.WriteString("?")
	return result.String()
}

// IsSingleNode reports if this regular expression can be represented as a single node when converted to a string.
// Regular expressions which cannot be represented as a single node need to have parenthesis placed around them to
// form a subexpression.
func (o *Optional) IsSingleNode() bool {
	return false
}

// Validate reports if the regular expression satisfies the required conditions to be considered valid.
// A nil return value indicates that the regular expression is valid.
// An error return value provides details about the unmet condition.
// In situations where the regular expression has children, all children are checked for validity recursively. If
// any child is not valid, this regular expression is also not valid.
func (o *Optional) Validate() error {
	if o.Child == nil {
		return errors.New("the regular expression requires a child")
	}
	return o.Child.Validate()
}
