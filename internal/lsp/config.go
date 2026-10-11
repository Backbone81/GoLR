package lsp

import (
	"io"
	"log/slog"
	"os"
)

// Config configures the language server.
type Config struct {
	// Transport is the channel to the client.
	Transport Transport

	Logger *slog.Logger

	// Parallel runs feature requests in their own goroutine.
	Parallel bool
}

// DefaultConfig provides the standard configuration which can be modified by options.
var DefaultConfig = Config{
	Transport: stdioTransport{},
	Logger:    slog.New(slog.NewTextHandler(os.Stderr, nil)),
	Parallel:  true,
}

// Transport is the channel to the client.
type Transport interface {
	Streams() (io.Reader, io.Writer)
}
