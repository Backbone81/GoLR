package lsp

import (
	"context"
	"io"
)

// streamsTransport talks over a reader and a writer which the caller owns, so closing it closes neither.
type streamsTransport struct {
	reader io.Reader
	writer io.Writer
}

func (t streamsTransport) Open(context.Context) (io.ReadWriteCloser, error) {
	return t, nil
}

func (t streamsTransport) Read(p []byte) (int, error) {
	return t.reader.Read(p)
}

func (t streamsTransport) Write(p []byte) (int, error) {
	return t.writer.Write(p)
}

func (streamsTransport) Close() error {
	return nil
}
