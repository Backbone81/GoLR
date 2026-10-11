package lsp

import (
	intlsp "github.com/backbone81/golr/internal/lsp"
)

// Option configures the language server.
type Option = intlsp.Option

// Serve runs a language server session. It returns an error if the session ended without shutdown.
var Serve = intlsp.Serve
