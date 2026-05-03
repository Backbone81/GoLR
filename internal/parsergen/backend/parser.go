package backend

import (
	"fmt"
	"strings"

	"golr/internal/parsergen/frontend"
)

// Parser is a parser.
type Parser struct {
	// Grammar is the grammar which was used to generate the parser.
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
