package lsp

import "io"

type streamsTransport struct {
	reader io.Reader
	writer io.Writer
}

func (t streamsTransport) Streams() (io.Reader, io.Writer) {
	return t.reader, t.writer
}
