package lsp

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// Config configures the language server.
type Config struct {
	// Transport is the channel to the client.
	Transport Transport

	Logger *slog.Logger

	// ClientProcessID is the process of the client, the session ends when it is gone. 0 names no process, the one of
	// initialize is watched then.
	ClientProcessID int

	// Parallel runs feature requests in their own goroutine.
	Parallel bool
}

// DefaultConfig provides the standard configuration which can be modified by options.
var DefaultConfig = Config{
	Transport: streamsTransport{reader: os.Stdin, writer: os.Stdout},
	Logger:    slog.New(slog.NewTextHandler(os.Stderr, nil)),
	Parallel:  true,
}

// Transport is the channel to the client.
type Transport interface {
	// Open connects to the client. Serve closes the channel when the session ends.
	Open(ctx context.Context) (io.ReadWriteCloser, error)
}
