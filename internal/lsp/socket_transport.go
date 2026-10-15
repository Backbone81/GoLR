package lsp

import (
	"context"
	"io"
	"net"
	"strconv"
)

// socketTransport connects to the TCP port the client listens on.
type socketTransport struct {
	port int
}

func (t socketTransport) Open(ctx context.Context) (io.ReadWriteCloser, error) {
	var dialer net.Dialer
	return dialer.DialContext(ctx, "tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(t.port)))
}
