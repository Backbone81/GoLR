package backend

import (
	"fmt"
	"strings"

	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// Parser is a parser.
type Parser struct {
	// Grammar is the grammar which was used to generate the parser. This is the augmented grammar, where a new start
	// symbol derives the old one followed by the end of input marker, not the grammar a frontend produced: it is the
	// grammar the terminal, nonterminal and production indexes of the states refer to. Whatever reads those indexes has
	// to be bound to this grammar, a conflict policy for instance, or it reads the neighbouring symbol of every index it
	// looks up.
	Grammar frontend.Grammar `json:"grammar" yaml:"grammar"`

	// States is the list of parser states.
	States []State `json:"states" yaml:"states"`
}

// Parser implements fmt.Stringer.
var _ fmt.Stringer = (*Parser)(nil)

// String returns a string representation.
func (p Parser) String() string {
	var builder strings.Builder
	builder.WriteString(p.Grammar.String())
	for i := range p.States {
		fmt.Fprintf(&builder, "state %d:\n", i)
		builder.WriteString(p.States[i].String())
		builder.WriteString("\n")
	}
	return builder.String()
}
