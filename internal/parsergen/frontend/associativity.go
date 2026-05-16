package frontend

import (
	"encoding/json"
	"fmt"

	"github.com/goccy/go-yaml"
)

// Associativity describes how terminals associate with each other.
type Associativity int

const (
	// AssociativityUndeclared means that no associativity is declared. This is the default for every terminal.
	AssociativityUndeclared Associativity = iota

	// AssociativityLeft introduces a left associativity.
	AssociativityLeft

	// AssociativityRight introduces a right associativity.
	AssociativityRight

	// AssociativityNone describes that the terminal should not associate at all and should trigger an error if some
	// association is needed.
	AssociativityNone
)

// Associativity implements fmt.Stringer.
var _ fmt.Stringer = (*Associativity)(nil)

// String returns a string for the associativity.
func (a Associativity) String() string {
	switch a {
	case AssociativityUndeclared:
		return "undeclared"
	case AssociativityLeft:
		return "left"
	case AssociativityRight:
		return "right"
	case AssociativityNone:
		return "none"
	default:
		return "unknown"
	}
}

func (a Associativity) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *Associativity) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	return a.fromString(s)
}

func (a Associativity) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(a.String())
}

func (a *Associativity) UnmarshalYAML(data []byte) error {
	var s string
	if err := yaml.Unmarshal(data, &s); err != nil {
		return err
	}
	return a.fromString(s)
}

func (a *Associativity) fromString(s string) error {
	switch s {
	case "undeclared":
		*a = AssociativityUndeclared
	case "left":
		*a = AssociativityLeft
	case "right":
		*a = AssociativityRight
	case "none":
		*a = AssociativityNone
	default:
		return fmt.Errorf("unknown associativity %q", s)
	}
	return nil
}
