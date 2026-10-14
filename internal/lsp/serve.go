package lsp

import (
	"bufio"
	"context"
	"runtime/debug"

	"github.com/backbone81/golr/internal/lsp/server"
)

// Serve runs a language server session until the client exits or the channel ends. It returns an error if a message
// could not be handled or the session ended without shutdown.
func Serve(ctx context.Context, options ...Option) error {
	config := ConfigFromOptions(options...)
	reader, writer := config.Transport.Streams()

	srv := server.New(config.Logger, version(), config.Parallel, config.ClientProcessID)
	return srv.Run(ctx, bufio.NewReader(reader), writer)
}

// version returns the module version of the golr binary, empty when it is unknown.
func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "(devel)" {
		return ""
	}
	return info.Main.Version
}
