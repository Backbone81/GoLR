package core

import (
	"errors"
	"fmt"
	"strings"
)

// Config configures how a core's GrammarToParser builds the parser tables.
type Config struct {
	// DefaultReductions enables the default-reduction table compaction (backend.ApplyDefaultReductions) after the
	// conflicts have been resolved.
	DefaultReductions bool

	// FailOnShiftReduceConflicts leaves a shift/reduce conflict unresolved.
	FailOnShiftReduceConflicts bool

	// FailOnReduceReduceConflicts leaves a reduce/reduce conflict unresolved.
	FailOnReduceReduceConflicts bool
}

// DefaultConfig provides the standard configuration which can be modified by options.
var DefaultConfig = Config{
	DefaultReductions: true,
}

// Option is the type required by all options modifying Config.
type Option func(*Config)

// ConfigFromOptions returns the default options with the given overrides applied.
func ConfigFromOptions(options ...Option) Config {
	config := DefaultConfig
	for _, opt := range options {
		opt(&config)
	}
	return config
}

// WithoutDefaultReductions disables the default-reduction compaction, so GrammarToParser returns the canonical resolved
// table where every reduce action is still keyed on its own lookahead terminals rather than one being moved into the
// state's default arm. This is useful for testing situations.
func WithoutDefaultReductions() Option {
	return func(options *Config) {
		options.DefaultReductions = false
	}
}

// FailOnShiftReduceConflicts makes GrammarToParser fail if a shift/reduce conflict is not resolved by precedence or
// associativity.
func FailOnShiftReduceConflicts() Option {
	return func(options *Config) {
		options.FailOnShiftReduceConflicts = true
	}
}

// FailOnReduceReduceConflicts makes GrammarToParser fail if a reduce/reduce conflict is not resolved by precedence or
// associativity.
func FailOnReduceReduceConflicts() Option {
	return func(options *Config) {
		options.FailOnReduceReduceConflicts = true
	}
}

// FailOnConflicts makes GrammarToParser fail if a shift/reduce or reduce/reduce conflict is not resolved by precedence
// or associativity.
func FailOnConflicts() Option {
	return func(options *Config) {
		options.FailOnShiftReduceConflicts = true
		options.FailOnReduceReduceConflicts = true
	}
}

// ErrOptionNotSupported reports an option which a core cannot apply.
var ErrOptionNotSupported = errors.New("option not supported")

// RejectFailOnConflicts returns an ErrOptionNotSupported naming the conflict kinds the config asks to fail on, or nil
// when it asks for none. A core which does not resolve the conflicts itself calls it, because silently ignoring these
// options would let a build pass which was meant to fail.
func RejectFailOnConflicts(config Config, coreName string) error {
	var kinds []string
	if config.FailOnShiftReduceConflicts {
		kinds = append(kinds, "shift/reduce")
	}
	if config.FailOnReduceReduceConflicts {
		kinds = append(kinds, "reduce/reduce")
	}
	if len(kinds) == 0 {
		return nil
	}
	return fmt.Errorf(
		"%w: failing on %s conflicts is not supported by %s",
		ErrOptionNotSupported,
		strings.Join(kinds, " and "),
		coreName,
	)
}
