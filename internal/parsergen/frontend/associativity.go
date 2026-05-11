package frontend

// Associativity describes how terminals associate.
type Associativity int

const (
	AssociativityUndeclared Associativity = iota
	AssociativityLeft
	AssociativityRight
	AssociativityNone
)

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
