package server

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"
)

// clientProcessPollInterval is the time between two checks whether the client process still runs.
const clientProcessPollInterval = 3 * time.Second

// clientProcessWatcher reports when the client process is gone. It watches only the first process it is given: the
// one of the command line, otherwise the one of initialize.
type clientProcessWatcher struct {
	stop    <-chan struct{}
	logger  *slog.Logger
	started atomic.Bool
	gone    chan struct{}
}

// newClientProcessWatcher returns a watcher which stops watching when the context ends.
func newClientProcessWatcher(ctx context.Context, logger *slog.Logger) *clientProcessWatcher {
	return &clientProcessWatcher{
		stop:   ctx.Done(),
		logger: logger,
		gone:   make(chan struct{}),
	}
}

// Watch starts watching the process unless a process is watched already. An id below 1 names no process.
func (w *clientProcessWatcher) Watch(pid int) {
	if pid < 1 {
		w.logger.Error("Invalid client process id.", "pid", pid)
		return
	}
	if !w.started.CompareAndSwap(false, true) {
		return
	}
	go w.poll(pid)
}

// Gone returns a channel which is closed when the client process is gone.
func (w *clientProcessWatcher) Gone() <-chan struct{} {
	return w.gone
}

// poll checks the process right away and then in intervals, until it is gone or the context ends. A process which
// cannot be opened counts as gone.
func (w *clientProcessWatcher) poll(pid int) {
	process, err := openClientProcess(pid)
	if err != nil {
		w.logger.Error("Client process gone.", "pid", pid, "error", err)
		close(w.gone)
		return
	}
	defer process.close()

	ticker := time.NewTicker(clientProcessPollInterval)
	defer ticker.Stop()
	for {
		if process.exited() {
			w.logger.Error("Client process gone.", "pid", pid)
			close(w.gone)
			return
		}
		select {
		case <-w.stop:
			return
		case <-ticker.C:
		}
	}
}
