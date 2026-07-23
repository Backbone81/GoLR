package core

// Config configures how a core's GrammarToParser builds the parser tables.
type Config struct {
	// DefaultReductions enables the default-reduction table compaction (backend.ApplyDefaultReductions) after the
	// conflicts have been resolved.
	DefaultReductions bool
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
