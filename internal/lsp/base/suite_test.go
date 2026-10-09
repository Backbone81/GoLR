package base_test

import (
	"io"
	"maps"
	"testing"

	"github.com/backbone81/golr/internal/lsp/base"
	"github.com/onsi/gomega/format"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	format.MaxLength = 0
	RegisterFailHandler(Fail)
	RunSpecs(t, "LSP Base Protocol Suite")
}

// Message is a copy of one parsed message, as the parser reuses its storage.
type Message struct {
	Header  map[string]string
	Content string
}

// ParseAll reads every message from reader and returns them with the final error.
func ParseAll(reader io.Reader) ([]Message, error) {
	parser := base.NewProtocolParser(reader)
	var messages []Message
	for parser.Next() {
		messages = append(messages, Message{
			Header:  maps.Clone(parser.Header()),
			Content: string(parser.Content()),
		})
	}
	return messages, parser.Err()
}
