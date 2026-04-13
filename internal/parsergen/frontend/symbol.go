package frontend

import (
	"fmt"
)

// Symbol is the textual representation of either a terminal or a nonterminal.
type Symbol struct {
	// Name is the technical name for that symbol.
	Name string `json:"name" yaml:"name"`

	// Alias is an alternative name for that symbol which might be less technical. For example while the technical name
	// for a terminal might be OP_PLUS the alias might be "+" to make it easier to read.
	Alias string `json:"alias" yaml:"alias"`
}

// Symbol implements fmt.Stringer.
var _ fmt.Stringer = (*Symbol)(nil)

// String returns the alias for the symbol if an alias is set. If no alias is set the name is returned.
func (s Symbol) String() string {
	if s.Alias == "" {
		return s.Name
	}
	return s.Alias
}

var (
	// SymbolEOF is the end of file symbol which marks the end of the parse.
	SymbolEOF = Symbol{
		Name: "$end",
	}
)
