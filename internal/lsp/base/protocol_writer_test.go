package base_test

import (
	"bytes"
	"errors"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/base"
)

var _ = Describe("WriteMessage", func() {
	It("writes the header and the content", func() {
		var output bytes.Buffer
		Expect(base.WriteMessage(&output, []byte("{}"))).To(Succeed())
		Expect(output.String()).To(Equal("Content-Length: 2\r\n\r\n{}"))
	})

	It("writes empty content", func() {
		var output bytes.Buffer
		Expect(base.WriteMessage(&output, nil)).To(Succeed())
		Expect(output.String()).To(Equal("Content-Length: 0\r\n\r\n"))
	})

	It("writes messages the parser reads back", func() {
		large := strings.Repeat("x", 5000)
		var output bytes.Buffer
		Expect(base.WriteMessage(&output, []byte("{}"))).To(Succeed())
		Expect(base.WriteMessage(&output, []byte(large))).To(Succeed())
		messages, err := ParseAll(&output)
		Expect(err).NotTo(HaveOccurred())
		Expect(messages).To(Equal([]Message{
			{Header: map[string]string{"Content-Length": "2"}, Content: "{}"},
			{Header: map[string]string{"Content-Length": "5000"}, Content: large},
		}))
	})

	It("returns the error of the writer", func() {
		writeErr := errors.New("broken pipe")
		Expect(base.WriteMessage(failingWriter{err: writeErr}, []byte("{}"))).To(MatchError(writeErr))
	})
})

// failingWriter fails every write.
type failingWriter struct {
	err error
}

func (w failingWriter) Write([]byte) (int, error) {
	return 0, w.err
}
