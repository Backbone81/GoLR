package lsp_test

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"

	"github.com/onsi/gomega/format"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp"
	"github.com/backbone81/golr/internal/lsp/base"
	"github.com/backbone81/golr/internal/lsp/lsptest"
)

func TestSuite(t *testing.T) {
	format.MaxLength = 0
	RegisterFailHandler(Fail)
	RunSpecs(t, "LSP Suite")
}

// serveOver runs a session with a client listening on the listener, which the transport option makes the server
// connect to, and returns the bodies the server wrote.
func serveOver(listener net.Listener, transport lsp.Option, messages ...string) ([]string, error) {
	done := make(chan error, 1)
	go func() {
		done <- lsp.Serve(context.Background(), transport, lsp.WithLogger(slog.New(slog.DiscardHandler)))
	}()

	conn, err := listener.Accept()
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(conn.Close)
	for _, message := range messages {
		Expect(base.WriteMessage(conn, []byte(message))).To(Succeed())
	}

	// The server closes the channel when the session ends, which ends the reading.
	bodies := readBodies(conn)
	lsptest.ExpectDecodableResults(messages, bodies)
	return bodies, <-done
}

// readBodies returns the bodies of all messages until the end of the reader.
func readBodies(reader io.Reader) []string {
	parser := base.NewProtocolParser(reader)
	var bodies []string
	for parser.Next() {
		bodies = append(bodies, string(parser.Content()))
	}
	Expect(parser.Err()).NotTo(HaveOccurred())
	return bodies
}
