package base_test

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/backbone81/golr/internal/lsp/base"
)

// FuzzProtocolParser reads arbitrary input. Every message the parser returns has to read back the same after writing
// it again, and its content has the length of its Content-Length.
func FuzzProtocolParser(f *testing.F) {
	for _, seed := range []string{
		"",
		"Content-Length: 2\r\n\r\n{}",
		"Content-Length: 2\r\nContent-Type: application/vscode-jsonrpc; charset=utf-8\r\n\r\n{}Content-Length: 4\r\n\r\nnull",
		"Content-Length: 2\r\n\r\n{",
		"Content-Length: two\r\n\r\n{}",
		"Content-Length: -1\r\n\r\n",
		"Content-Length:2\r\n\r\n{}",
		"Content-Length: 2\rX\n\r\n{}",
		"Content-Length: 2\r\nX-Name: \xc3\xa4\r\n\r\n{}",
		"\r\n{}",
	} {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, input []byte) {
		parser := base.NewProtocolParser(bytes.NewReader(input))
		var contents [][]byte
		for parser.Next() {
			content := bytes.Clone(parser.Content())
			length, err := strconv.Atoi(parser.Header()[base.HeaderContentLength])
			if err != nil || length != len(content) {
				t.Fatalf("content of %d bytes for Content-Length %q", len(content), parser.Header()[base.HeaderContentLength])
			}
			contents = append(contents, content)
		}

		var written bytes.Buffer
		for _, content := range contents {
			if err := base.WriteMessage(&written, content); err != nil {
				t.Fatal(err)
			}
		}
		reread := base.NewProtocolParser(&written)
		count := 0
		for ; reread.Next(); count++ {
			if count >= len(contents) || !bytes.Equal(reread.Content(), contents[count]) {
				t.Fatalf("message %d reads back differently", count)
			}
		}
		if err := reread.Err(); err != nil {
			t.Fatalf("reading back: %v", err)
		}
		if count != len(contents) {
			t.Fatalf("%d messages read back, %d written", count, len(contents))
		}
	})
}
