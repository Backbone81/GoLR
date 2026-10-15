package lsp

import (
	intlsp "github.com/backbone81/golr/internal/lsp"
)

// Option configures the language server.
type Option = intlsp.Option

// Serve runs a language server session over stdin and stdout, unless an option names another channel. It returns an
// error if the session ended without shutdown.
var Serve = intlsp.Serve

var (
	// WithPipe talks to the client over the pipe with the given name, which the client listens on. The name is the
	// path of a Unix domain socket on Linux and macOS and of a named pipe on Windows.
	WithPipe = intlsp.WithPipe

	// WithSocket talks to the client over the TCP port on 127.0.0.1 the client listens on.
	WithSocket = intlsp.WithSocket

	// WithClientProcessID ends the session when the client process with the id is gone. It takes precedence over the
	// process id of initialize.
	WithClientProcessID = intlsp.WithClientProcessID
)
