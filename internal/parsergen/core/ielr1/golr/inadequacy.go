package golr

import (
	"github.com/backbone81/golr/internal/parsergen/conflict"
)

// Inadequacy describes a single grammar-relative inadequacy of the LALR(1) parser tables. An inadequacy is identified
// by the conflict by which it manifests, which is the conflicted state, the conflicted terminal and all the
// contributions the conflict has within that state. This is definition 3.27 of IELR(1) and named
// "inadequacy_lists" there.
type Inadequacy struct {
	// StateIdx is the state index of the conflicted state.
	StateIdx int

	// TerminalIdx is the terminal index of the conflicted terminal.
	TerminalIdx int

	// Contributions are all the contributions the conflict has within the conflicted state. This is the value of the
	// contributions function of definition 2.17 of IELR(1). The order of the contributions is significant, because the
	// contribution matrix of an annotation is indexed by the index of the contribution within this set.
	Contributions conflict.ContributionSet
}
