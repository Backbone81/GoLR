package frontend

import (
	"errors"
	"strings"
)

// Or is a regular expression matching one of its children.
// The children need to implement the [Node] interface.
type Or struct {
	Children []*Node
}

// String returns a string representation of this regular expression.
func (o *Or) String() string {
	var result strings.Builder
	for i := range o.Children {
		if i > 0 {
			result.WriteString("|")
		}
		result.WriteString(o.Children[i].String())
	}
	return result.String()
}

// IsSingleNode reports if this regular expression can be represented as a single node when converted to a string.
// Regular expressions which cannot be represented as a single node need to have parenthesis placed around them to
// form a subexpression.
func (o *Or) IsSingleNode() bool {
	return len(o.Children) == 1
}

// Validate reports if the regular expression satisfies the required conditions to be considered valid.
// A nil return value indicates that the regular expression is valid.
// An error return value provides details about the unmet condition.
// In situations where the regular expression has children, all children are checked for validity recursively. If
// any child is not valid, this regular expression is also not valid.
func (o *Or) Validate() error {
	if len(o.Children) < 2 {
		return errors.New("the regular expression requires at least two children")
	}
	for _, child := range o.Children {
		if err := child.Validate(); err != nil {
			return err
		}
	}
	return nil
}
