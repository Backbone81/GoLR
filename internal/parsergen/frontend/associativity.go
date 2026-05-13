package frontend

import "fmt"

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
