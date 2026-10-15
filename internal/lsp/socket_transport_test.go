package lsp_test

import (
	"context"
	"log/slog"
	"net"
	"net/netip"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp"
)

var _ = Describe("WithSocket", func() {
	It("formats a document over the port the client listens on", func() {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(listener.Close)
		port := netip.MustParseAddrPort(listener.Addr().String()).Port()

		responses, err := serveOver(listener, lsp.WithSocket(int(port)), formattingSession...)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(ConsistOf(formattingSessionResponses))
	})

	It("fails when no client listens on the port", func() {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred())
		port := netip.MustParseAddrPort(listener.Addr().String()).Port()
		Expect(listener.Close()).To(Succeed())

		err = lsp.Serve(context.Background(), lsp.WithSocket(int(port)), lsp.WithLogger(slog.New(slog.DiscardHandler)))
		Expect(err).To(MatchError(ContainSubstring("connection refused")))
	})
})
