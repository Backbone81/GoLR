//go:build !windows

package lsp_test

import (
	"context"
	"log/slog"
	"net"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp"
)

var _ = Describe("WithPipe", func() {
	It("formats a document over the Unix domain socket the client listens on", func() {
		name := filepath.Join(GinkgoT().TempDir(), "lsp.sock")
		listener, err := net.Listen("unix", name)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(listener.Close)

		responses, err := serveOver(listener, lsp.WithPipe(name), formattingSession...)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(ConsistOf(formattingSessionResponses))
	})

	It("fails when no client listens on the pipe", func() {
		name := filepath.Join(GinkgoT().TempDir(), "lsp.sock")

		err := lsp.Serve(context.Background(), lsp.WithPipe(name), lsp.WithLogger(slog.New(slog.DiscardHandler)))
		Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
	})
})
