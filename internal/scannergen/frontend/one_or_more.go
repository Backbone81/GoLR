package frontend

import (
	"errors"
	"strings"
)

// OneOrMore is a regular expression matching one or more instances of its child.
// The child needs to implement the [Node] interface.
type OneOrMore struct {
	Child *Node `json:"child" yaml:"child"`
}

// String returns a string representation of this regular expression.
func (o *OneOrMore) String() string {
	var result strings.Builder
	if !o.Child.IsSingleNode() {
		result.WriteString("(")
	}
	result.WriteString(o.Child.String())
	if !o.Child.IsSingleNode() {
		result.WriteString(")")
	}
	result.WriteString("+")
	return result.String()
}

// IsSingleNode reports if this regular expression can be represented as a single node when converted to a string.
// Regular expressions which cannot be represented as a single node need to have parenthesis placed around them to
// form a subexpression.
func (o *OneOrMore) IsSingleNode() bool {
	return false
}

// Validate reports if the regular expression satisfies the required conditions to be considered valid.
// A nil return value indicates that the regular expression is valid.
// An error return value provides details about the unmet condition.
// In situations where the regular expression has children, all children are checked for validity recursively. If
// any child is not valid, this regular expression is also not valid.
func (o *OneOrMore) Validate() error {
	if o.Child == nil {
		return errors.New("the regular expression requires a child")
	}
	return o.Child.Validate()
}

// NewNodeOneOrMore creates a new node for one or more
func NewNodeOneOrMore(child *Node) *Node {
	return &Node{
		Kind: KindOneOrMore,
		OneOrMore: OneOrMore{
			Child: child,
		},
	}
}
