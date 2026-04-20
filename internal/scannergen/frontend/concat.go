package frontend

import (
	"errors"
	"strings"
)

// Concat is a regular expression matching all its children in sequence.
// The children need to implement the [Node] interface.
type Concat struct {
	Children []*Node `json:"children" yaml:"children"`
}

// String returns a string representation of this regular expression.
func (c *Concat) String() string {
	var result strings.Builder
	for i := range c.Children {
		result.WriteString(c.Children[i].String())
	}
	return result.String()
}

// IsSingleNode reports if this regular expression can be represented as a single node when converted to a string.
// Regular expressions which cannot be represented as a single node need to have parenthesis placed around them to
// form a subexpression.
func (c *Concat) IsSingleNode() bool {
	if len(c.Children) == 1 {
		return c.Children[0].IsSingleNode()
	}
	return false
}

// Validate reports if the regular expression satisfies the required conditions to be considered valid.
// A nil return value indicates that the regular expression is valid.
// An error return value provides details about the unmet condition.
// In situations where the regular expression has children, all children are checked for validity recursively. If
// any child is not valid, this regular expression is also not valid.
func (c *Concat) Validate() error {
	if len(c.Children) < 2 {
		return errors.New("the regular expression requires at least two children")
	}
	for _, child := range c.Children {
		if err := child.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func NewNodeConcat(nodes ...*Node) *Node {
	return &Node{
		Kind: KindConcat,
		Concat: Concat{
			Children: nodes,
		},
	}
}
