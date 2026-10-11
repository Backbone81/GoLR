package lsp

import (
	"io"
	"log/slog"
)

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

// WithStreams talks to the client over the given streams instead of stdin and stdout.
func WithStreams(reader io.Reader, writer io.Writer) Option {
	return func(config *Config) {
		config.Transport = streamsTransport{reader: reader, writer: writer}
	}
}

// WithLogger logs to the given logger instead of stderr.
func WithLogger(logger *slog.Logger) Option {
	return func(config *Config) {
		config.Logger = logger
	}
}

// WithoutParallelism handles every message before the next one is read, for debugging and deterministic tests.
func WithoutParallelism() Option {
	return func(config *Config) {
		config.Parallel = false
	}
}
