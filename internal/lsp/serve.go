package lsp

import (
	"bufio"
	"context"
	"errors"
	"runtime/debug"

	"github.com/backbone81/golr/internal/lsp/server"
)

// Serve runs a language server session until the client exits or the channel ends. It returns an error if the channel
// could not be opened, a message could not be handled or the session ended without shutdown.
func Serve(ctx context.Context, options ...Option) error {
	config := ConfigFromOptions(options...)
	channel, err := config.Transport.Open(ctx)
	if err != nil {
		return err
	}

	srv := server.New(config.Logger, version(), config.Parallel, config.ClientProcessID)
	err = srv.Run(ctx, bufio.NewReader(channel), channel)
	return errors.Join(err, channel.Close())
}

// version returns the module version of the golr binary, empty when it is unknown.
func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "(devel)" {
		return ""
	}
	return info.Main.Version
}
