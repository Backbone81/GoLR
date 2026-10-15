//go:build !windows

package lsp

import (
	"context"
	"io"
	"net"
)

func (t pipeTransport) Open(ctx context.Context) (io.ReadWriteCloser, error) {
	var dialer net.Dialer
	return dialer.DialContext(ctx, "unix", t.name)
}
