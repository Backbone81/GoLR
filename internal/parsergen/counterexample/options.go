package counterexample

import "time"

// Option modifies the configuration of Find.
type Option func(*Config)

// WithTimeLimit limits the search for a unifying counterexample of a pair of conflict items, after which the pair gets
// a nonunifying counterexample. The default is 5 seconds.
func WithTimeLimit(limit time.Duration) Option {
	return func(c *Config) {
		c.TimeLimit = limit
	}
}

// WithConfigurationLimit limits the configurations the search for a unifying counterexample of a pair of conflict items
// processes, after which the pair gets a nonunifying counterexample. Unlike the time limits, it does not depend on the
// speed of the machine. The default is no limit.
func WithConfigurationLimit(limit int) Option {
	return func(c *Config) {
		c.ConfigurationLimit = limit
	}
}

// WithTotalTimeLimit limits the searches for unifying counterexamples of all conflicts together, after which the
// remaining pairs of conflict items get nonunifying counterexamples. The default is 2 minutes.
func WithTotalTimeLimit(limit time.Duration) Option {
	return func(c *Config) {
		c.TotalTimeLimit = limit
	}
}
