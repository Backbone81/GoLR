package lsp

import (
	"io"
	"os"
)

type stdioTransport struct{}

func (stdioTransport) Streams() (io.Reader, io.Writer) {
	return os.Stdin, os.Stdout
}
