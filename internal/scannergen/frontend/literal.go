package frontend

import (
	"encoding/json"
	"errors"
)

// Literal is a regular expression matching its text as a literal.
type Literal struct {
	Text string `json:"text" yaml:"text"`
}

// String returns a string representation of this regular expression.
func (l *Literal) String() string {
	return l.Text
}

// IsSingleNode reports if this regular expression can be represented as a single node when converted to a string.
// Regular expressions which cannot be represented as a single node need to have parenthesis placed around them to
// form a subexpression.
func (l *Literal) IsSingleNode() bool {
	return len(l.Text) == 1
}

// Validate reports if the regular expression satisfies the required conditions to be considered valid.
// A nil return value indicates that the regular expression is valid.
// An error return value provides details about the unmet condition.
// In situations where the regular expression has children, all children are checked for validity recursively. If
// any child is not valid, this regular expression is also not valid.
func (l *Literal) Validate() error {
	if len(l.Text) == 0 {
		return errors.New("literal must have at least one character")
	}
	return nil
}

// MarshalYAML encodes the Literal as YAML. It uses a JSON-encoded (double-quoted) scalar for the Text field to
// avoid block scalar notation, which goccy/go-yaml uses for strings containing newlines or other special characters.
// Block scalars inside sequence items cause parsing errors in goccy/go-yaml when the value contains newlines.
func (l Literal) MarshalYAML() ([]byte, error) {
	textJSON, err := json.Marshal(l.Text)
	if err != nil {
		return nil, err
	}
	return []byte("text: " + string(textJSON) + "\n"), nil
}
