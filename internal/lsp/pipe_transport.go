package lsp

// pipeTransport connects to the pipe the client listens on: a Unix domain socket on Linux and macOS, a named pipe on
// Windows.
type pipeTransport struct {
	name string
}
