//go:build windows

package server

import (
	"syscall"
)

// clientProcess is a handle of the client process, which is signaled when the process exits. The open handle keeps
// the id from being reused.
type clientProcess struct {
	handle syscall.Handle
}

// openClientProcess opens a handle of the process with the id, which must be at least 1.
func openClientProcess(pid int) (clientProcess, error) {
	//nolint:gosec // The caller checks the id, and process ids fit into uint32.
	handle, err := syscall.OpenProcess(syscall.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		return clientProcess{}, err
	}
	return clientProcess{handle: handle}, nil
}

// exited reports whether the process is gone, without waiting for it.
func (p clientProcess) exited() bool {
	event, _ := syscall.WaitForSingleObject(p.handle, 0)
	return event == syscall.WAIT_OBJECT_0
}

// close closes the handle of the process.
func (p clientProcess) close() {
	// The error of a close only means the handle was invalid, which leaves nothing to close.
	_ = syscall.CloseHandle(p.handle)
}
