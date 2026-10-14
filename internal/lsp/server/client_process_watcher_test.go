package server_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/base"
	"github.com/backbone81/golr/internal/lsp/server"
)

var _ = Describe("Server", func() {
	It("ends the session when the client process of New is gone", func(ctx SpecContext) {
		input, inputWriter := io.Pipe()
		DeferCleanup(inputWriter.Close)
		srv := server.New(slog.New(slog.DiscardHandler), "", false, exitedProcess(ctx))
		done := make(chan error, 1)
		go func() {
			done <- srv.Run(context.Background(), input, io.Discard)
		}()
		Eventually(done).Should(Receive(MatchError(server.ErrClientProcessGone)))
	})

	It("ends the session when the client process of initialize exits", func(ctx SpecContext) {
		pending, pendingWriter := io.Pipe()
		DeferCleanup(pendingWriter.Close)
		input := io.MultiReader(strings.NewReader(initializeWithProcess(startProcess(ctx, "1"))), pending)
		srv := server.New(slog.New(slog.DiscardHandler), "", false, 0)
		done := make(chan error, 1)
		go func() {
			done <- srv.Run(context.Background(), input, io.Discard)
		}()
		// The process exits after a second, which the watcher notices within its interval of three seconds.
		Eventually(done, 10*time.Second).Should(Receive(MatchError(server.ErrClientProcessGone)))
	})

	It("watches the client process of New instead of the one of initialize", func(ctx SpecContext) {
		pending, pendingWriter := io.Pipe()
		input := io.MultiReader(strings.NewReader(initializeWithProcess(exitedProcess(ctx))), pending)
		srv := server.New(slog.New(slog.DiscardHandler), "", false, startProcess(ctx, "60"))
		done := make(chan error, 1)
		go func() {
			done <- srv.Run(context.Background(), input, io.Discard)
		}()
		Consistently(done, 500*time.Millisecond).ShouldNot(Receive())
		Expect(pendingWriter.Close()).To(Succeed())
		Eventually(done).Should(Receive(MatchError(server.ErrSessionEndedWithoutShutdown)))
	})
})

// initializeWithProcess returns the initialize request naming the client process as message.
func initializeWithProcess(pid int) string {
	var message strings.Builder
	content := fmt.Sprintf(
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"processId":%d,"capabilities":{}}}`, pid)
	Expect(base.WriteMessage(&message, []byte(content))).To(Succeed())
	return message.String()
}

// exitedProcess returns the id of a process which has exited.
func exitedProcess(ctx context.Context) int {
	cmd := exec.CommandContext(ctx, "true")
	Expect(cmd.Run()).To(Succeed())
	return cmd.Process.Pid
}

// startProcess starts a process which runs for the seconds and returns its id. The process is reaped when it exits,
// so it is gone then, and killed when the context ends.
func startProcess(ctx context.Context, seconds string) int {
	cmd := exec.CommandContext(ctx, "sleep", seconds)
	Expect(cmd.Start()).To(Succeed())
	go func() {
		// Waiting reaps the process, a zombie would still count as running. The exit status does not matter.
		_ = cmd.Wait()
	}()
	return cmd.Process.Pid
}
