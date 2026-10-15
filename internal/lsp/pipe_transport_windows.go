//go:build windows

package lsp

import (
	"context"
	"io"
	"os"
	"syscall"
)

// Open opens the client side of the named pipe. The pipe is opened for overlapped I/O, as a synchronous handle would
// block every write while the read of the next message waits.
func (t pipeTransport) Open(context.Context) (io.ReadWriteCloser, error) {
	return os.OpenFile(t.name, os.O_RDWR|syscall.FILE_FLAG_OVERLAPPED, 0)
}
