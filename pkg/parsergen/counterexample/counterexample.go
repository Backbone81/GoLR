// Package counterexample provides the public API re-exports for the counterexamples of the conflicts a parser core
// reports.
package counterexample

import (
	intcounterexample "github.com/backbone81/golr/internal/parsergen/counterexample"
)

type (
	// Counterexample shows where a conflict between two items comes from, with one derivation per item. A unifying
	// counterexample proves the grammar ambiguous, a nonunifying one shows two derivations which share a prefix up to
	// the conflict.
	Counterexample = intcounterexample.Counterexample

	// Option modifies the configuration of Find.
	Option = intcounterexample.Option
)

var (
	// DefaultConfig provides the configuration of Find without options.
	DefaultConfig = intcounterexample.DefaultConfig

	// Find returns the counterexamples of the conflicts a GoLR core returned together with the parser tables, one slice
	// per conflict in the same order, with one counterexample per pair of competing items. The report writers take the
	// result next to the conflicts.
	Find = intcounterexample.Find

	// WithTimeLimit limits the search for a unifying counterexample of a pair of conflict items, after which the pair
	// gets a nonunifying counterexample. The default is 5 seconds.
	WithTimeLimit = intcounterexample.WithTimeLimit

	// WithTotalTimeLimit limits the searches for unifying counterexamples of all conflicts together, after which the
	// remaining pairs of conflict items get nonunifying counterexamples. The default is 2 minutes.
	WithTotalTimeLimit = intcounterexample.WithTotalTimeLimit
)
