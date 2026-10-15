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

// WithPipe talks to the client over the pipe with the given name, which the client listens on. The name is the path
// of a Unix domain socket on Linux and macOS and of a named pipe on Windows.
func WithPipe(name string) Option {
	return func(config *Config) {
		config.Transport = pipeTransport{name: name}
	}
}

// WithSocket talks to the client over the TCP port on 127.0.0.1 the client listens on.
func WithSocket(port int) Option {
	return func(config *Config) {
		config.Transport = socketTransport{port: port}
	}
}

// WithLogger logs to the given logger instead of stderr.
func WithLogger(logger *slog.Logger) Option {
	return func(config *Config) {
		config.Logger = logger
	}
}

// WithClientProcessID ends the session when the client process with the id is gone. It takes precedence over the
// process id of initialize.
func WithClientProcessID(pid int) Option {
	return func(config *Config) {
		config.ClientProcessID = pid
	}
}

// WithoutParallelism handles every message before the next one is read, for debugging and deterministic tests.
func WithoutParallelism() Option {
	return func(config *Config) {
		config.Parallel = false
	}
}
