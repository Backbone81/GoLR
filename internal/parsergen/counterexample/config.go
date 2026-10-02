package counterexample

import "time"

// Config configures the search of Find.
type Config struct {
	// TimeLimit limits the unifying search of a pair of conflict items.
	TimeLimit time.Duration

	// TotalTimeLimit limits the unifying searches of all pairs together.
	TotalTimeLimit time.Duration
}

// DefaultConfig holds the time limits of the paper (section 6, "Constructing nonunifying counterexamples").
var DefaultConfig = Config{
	TimeLimit:      5 * time.Second,
	TotalTimeLimit: 2 * time.Minute,
}

// ConfigFromOptions returns the default configuration with the options applied.
func ConfigFromOptions(options ...Option) Config {
	result := DefaultConfig
	for _, option := range options {
		option(&result)
	}
	return result
}
