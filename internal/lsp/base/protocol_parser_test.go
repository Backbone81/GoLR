package base_test

import (
	"errors"
	"io"
	"strings"
	"testing/iotest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/base"
)

var _ = Describe("ProtocolParser", func() {
	It("ends without error on an empty stream", func() {
		messages, err := ParseAll(strings.NewReader(""))
		Expect(err).NotTo(HaveOccurred())
		Expect(messages).To(BeEmpty())
	})

	It("reads one message", func() {
		messages, err := ParseAll(strings.NewReader("Content-Length: 2\r\n\r\n{}"))
		Expect(err).NotTo(HaveOccurred())
		Expect(messages).To(Equal([]Message{
			{Header: map[string]string{"Content-Length": "2"}, Content: "{}"},
		}))
	})

	It("reads several messages and clears the header between them", func() {
		messages, err := ParseAll(strings.NewReader(
			"Content-Length: 2\r\nContent-Type: application/vscode-jsonrpc; charset=utf-8\r\n\r\n{}" +
				"Content-Length: 4\r\n\r\nnull",
		))
		Expect(err).NotTo(HaveOccurred())
		Expect(messages).To(Equal([]Message{
			{
				Header: map[string]string{
					"Content-Length": "2",
					"Content-Type":   "application/vscode-jsonrpc; charset=utf-8",
				},
				Content: "{}",
			},
			{
				Header:  map[string]string{"Content-Length": "4"},
				Content: "null",
			},
		}))
	})

	It("reads a message delivered one byte per read", func() {
		messages, err := ParseAll(iotest.OneByteReader(strings.NewReader("Content-Length: 7\r\n\r\n[1,2,3]")))
		Expect(err).NotTo(HaveOccurred())
		Expect(messages).To(HaveLen(1))
		Expect(messages[0].Content).To(Equal("[1,2,3]"))
	})

	It("reads a message with empty content", func() {
		messages, err := ParseAll(strings.NewReader("Content-Length: 0\r\n\r\nContent-Length: 2\r\n\r\n{}"))
		Expect(err).NotTo(HaveOccurred())
		Expect(messages).To(HaveLen(2))
		Expect(messages[0].Content).To(BeEmpty())
		Expect(messages[1].Content).To(Equal("{}"))
	})

	It("reads content larger than the initial buffer followed by smaller content", func() {
		large := strings.Repeat("x", 5000)
		messages, err := ParseAll(strings.NewReader(
			"Content-Length: 5000\r\n\r\n" + large + "Content-Length: 2\r\n\r\n{}",
		))
		Expect(err).NotTo(HaveOccurred())
		Expect(messages).To(HaveLen(2))
		Expect(messages[0].Content).To(Equal(large))
		Expect(messages[1].Content).To(Equal("{}"))
	})

	It("canonicalizes header names", func() {
		messages, err := ParseAll(strings.NewReader("content-length: 2\r\nCONTENT-TYPE: text/plain\r\n\r\n{}"))
		Expect(err).NotTo(HaveOccurred())
		Expect(messages).To(HaveLen(1))
		Expect(messages[0].Header).To(Equal(map[string]string{
			"Content-Length": "2",
			"Content-Type":   "text/plain",
		}))
	})

	It("keeps unknown headers and values containing colons and spaces", func() {
		messages, err := ParseAll(strings.NewReader("X-Custom: a: b c\r\nX-Empty: \r\nContent-Length: 2\r\n\r\n{}"))
		Expect(err).NotTo(HaveOccurred())
		Expect(messages).To(HaveLen(1))
		Expect(messages[0].Header).To(Equal(map[string]string{
			"X-Custom":       "a: b c",
			"X-Empty":        "",
			"Content-Length": "2",
		}))
	})

	It("does not continue after an error", func() {
		parser := base.NewProtocolParser(strings.NewReader("Content-Length: x\r\n\r\nContent-Length: 2\r\n\r\n{}"))
		Expect(parser.Next()).To(BeFalse())
		Expect(parser.Err()).To(HaveOccurred())
		Expect(parser.Next()).To(BeFalse())
		Expect(parser.Err()).To(HaveOccurred())
	})

	It("returns errors of the underlying reader", func() {
		readErr := errors.New("broken pipe")
		Expect(ParseAll(iotest.ErrReader(readErr))).Error().To(MatchError(readErr))
	})

	It("returns errors of the underlying reader while reading content", func() {
		readErr := errors.New("broken pipe")
		_, err := ParseAll(io.MultiReader(
			strings.NewReader("Content-Length: 2\r\n\r\n{"),
			iotest.ErrReader(readErr),
		))
		Expect(err).To(MatchError(readErr))
	})

	DescribeTable("fails on an unexpected end of the stream",
		func(input string) {
			messages, err := ParseAll(strings.NewReader(input))
			Expect(err).To(MatchError(io.ErrUnexpectedEOF))
			Expect(messages).To(BeEmpty())
		},
		Entry("in a header name", "Content-Len"),
		Entry("after the colon", "Content-Length:"),
		Entry("in a header value", "Content-Length: 2"),
		Entry("after the carriage return of a header field", "Content-Length: 2\r"),
		Entry("before the header/content separator", "Content-Length: 2\r\n"),
		Entry("in the header/content separator", "Content-Length: 2\r\n\r"),
		Entry("before the content", "Content-Length: 2\r\n\r\n"),
		Entry("in the content", "Content-Length: 2\r\n\r\n{"),
	)

	It("fails on an unexpected end of the stream after a complete message", func() {
		messages, err := ParseAll(strings.NewReader("Content-Length: 2\r\n\r\n{}Content"))
		Expect(err).To(MatchError(io.ErrUnexpectedEOF))
		Expect(messages).To(HaveLen(1))
	})

	DescribeTable("fails on a malformed header",
		func(input string, expected string) {
			messages, err := ParseAll(strings.NewReader(input))
			Expect(err).To(MatchError(ContainSubstring(expected)))
			Expect(messages).To(BeEmpty())
		},
		Entry("without Content-Length", "Content-Type: text/plain\r\n\r\n{}", `required header "Content-Length" is missing`),
		Entry("without any header field", "\r\n{}", `required header "Content-Length" is missing`),
		Entry("with a non-integer Content-Length", "Content-Length: two\r\n\r\n{}", "did not provide an integer"),
		Entry("with an empty Content-Length", "Content-Length: \r\n\r\n{}", "did not provide an integer"),
		Entry("with a negative Content-Length", "Content-Length: -1\r\n\r\n", "out of range"),
		Entry("with a Content-Length above 10 MB", "Content-Length: 10485761\r\n\r\n", "out of range"),
		Entry("with an empty header name", ": 2\r\n\r\n{}", "header field has an empty name"),
		Entry("without a separator", "Content-Length\r\n\r\n{}", `header field "Content-Length" has no ': ' separator`),
		Entry("without a space after the colon", "Content-Length:2\r\n\r\n{}", "expected ' ' after ':'"),
		Entry("with a carriage return not followed by a line feed", "Content-Length: 2\rX\n\r\n{}", `expected '\n' after '\r'`),
		Entry("with a malformed header/content separator", "Content-Length: 2\r\n\rX{}", "header/content separator"),
		Entry("with a non-ASCII byte", "Content-Length: 2\r\nX-Name: \xc3\xa4\r\n\r\n{}", "non-ASCII byte 0xc3 in header"),
	)
})
