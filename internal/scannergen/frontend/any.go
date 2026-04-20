package frontend

// Any is a regular expression matching any character.
// The character must be in the range of valid Unicode characters from 0x00 to [unicode.MaxRune].
type Any struct{}

// String returns a string representation of this regular expression.
func (a *Any) String() string {
	return "."
}

// IsSingleNode reports if this regular expression can be represented as a single node when converted to a string.
// Regular expressions which cannot be represented as a single node need to have parenthesis placed around them to
// form a subexpression.
func (a *Any) IsSingleNode() bool {
	return true
}

// Validate reports if the regular expression satisfies the required conditions to be considered valid.
// A nil return value indicates that the regular expression is valid.
// An error return value provides details about the unmet condition.
// In situations where the regular expression has children, all children are checked for validity recursively. If
// any child is not valid, this regular expression is also not valid.
func (a *Any) Validate() error {
	return nil
}

// NewNodeAny returns a node for any character.
func NewNodeAny() *Node {
	return &Node{
		Kind: KindAny,
	}
}
