//go:build !windows

package server

import (
	"errors"
	"os"
	"syscall"
)

// clientProcess is the client process, checked with signal 0, which tests whether a signal could be sent.
type clientProcess struct {
	process *os.Process
}

// openClientProcess returns the process with the id.
func openClientProcess(pid int) (clientProcess, error) {
	process, err := os.FindProcess(pid)
	if err != nil {
		return clientProcess{}, err
	}
	return clientProcess{process: process}, nil
}

// exited reports whether the process is gone. A process of another user still runs, signal 0 then fails with a
// permission error.
func (p clientProcess) exited() bool {
	return errors.Is(p.process.Signal(syscall.Signal(0)), os.ErrProcessDone)
}

// close releases the resources of the process.
func (p clientProcess) close() {
	// The error of a release only means there was nothing left to release.
	_ = p.process.Release()
}
