package core

import (
	intcore "github.com/backbone81/golr/internal/parsergen/core"
)

// Option configures how a core's GrammarToParser builds the parser tables.
type Option = intcore.Option

var (
	// FailOnConflicts makes GrammarToParser fail if a shift/reduce or reduce/reduce conflict is not resolved by
	// precedence or associativity.
	FailOnConflicts = intcore.FailOnConflicts

	// FailOnShiftReduceConflicts makes GrammarToParser fail if a shift/reduce conflict is not resolved by precedence or
	// associativity.
	FailOnShiftReduceConflicts = intcore.FailOnShiftReduceConflicts

	// FailOnReduceReduceConflicts makes GrammarToParser fail if a reduce/reduce conflict is not resolved by precedence
	// or associativity.
	FailOnReduceReduceConflicts = intcore.FailOnReduceReduceConflicts

	// FailOnWarnings makes GrammarToParser fail if there are warnings. The warnings are joined into the error instead
	// of being returned on their own.
	FailOnWarnings = intcore.FailOnWarnings
)
