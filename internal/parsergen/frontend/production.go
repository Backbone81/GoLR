package frontend

import (
	"fmt"
	"strconv"
	"strings"
)

// Production is a production of a context-free grammar. The Nonterminal is the left hand side of the production and
// the Symbols are the right hand side of the production.
type Production struct {
	// NonterminalIdx is the nonterminal index which makes up the left hand side of the production.
	NonterminalIdx int `json:"nonterminalIdx" yaml:"nonterminalIdx"`

	// SymbolRefs is the list of symbols which make up the right hand side of the production.
	SymbolRefs []SymbolRef `json:"symbolRefs" yaml:"symbolRefs"`

	// PrecedenceTerminalIdx provides the terminal index to use for deriving the precedence for this production.
	// If this is nil, the default behavior is used which is the precedence of the rightmost terminal.
	PrecedenceTerminalIdx *int `json:"precedenceTerminalIdx,omitempty" yaml:"precedenceTerminalIdx,omitempty"`
}

// Production implements fmt.Stringer.
var _ fmt.Stringer = (*Production)(nil)

// String returns a string representation of the production.
func (p Production) String() string {
	var builder strings.Builder
	builder.WriteString(strconv.Itoa(p.NonterminalIdx))
	builder.WriteString(" -> ")
	for idx, symbolRef := range p.SymbolRefs {
		if idx > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(symbolRef.String())
	}
	return builder.String()
}
