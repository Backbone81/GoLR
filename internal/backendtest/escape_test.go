package backendtest_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/backendtest"
)

var _ = Describe("EscapeLexeme", func() {
	DescribeTable("escapes a lexeme to the canonical form",
		func(lexeme string, expected string) {
			Expect(backendtest.EscapeLexeme(lexeme)).To(Equal(expected))
		},
		Entry("empty", "", ""),
		Entry("printable ASCII passes through", "abc XYZ 0-9 !~", "abc XYZ 0-9 !~"),
		// A lexeme which is a single space is the reason the event writes quotes around this. Without them it would be
		// indistinguishable from an empty lexeme, and the trace line would end in trailing whitespace.
		Entry("a space", " ", " "),
		Entry("quote", `"`, `\"`),
		Entry("backslash", `\`, `\\`),
		Entry("newline", "\n", `\n`),
		Entry("carriage return", "\r", `\r`),
		Entry("tab", "\t", `\t`),
		Entry("the lowest byte", "\x00", `\x00`),
		Entry("the byte below the printable range", "\x1f", `\x1f`),
		Entry("the byte above the printable range", "\x7f", `\x7f`),
		Entry("the lowest byte with the high bit set", "\x80", `\x80`),
		Entry("the highest byte", "\xff", `\xff`),
		// A lexeme is a sequence of bytes and not a sequence of runes, so a multi byte UTF-8 sequence becomes one
		// escape per byte. That is what every target language can reproduce without knowing about UTF-8, and it is
		// where Go's own %q would differ by writing the character instead.
		Entry("a two byte UTF-8 sequence", "ä", `\xc3\xa4`),
		Entry("a three byte UTF-8 sequence", "€", `\xe2\x82\xac`),
		Entry("a four byte UTF-8 sequence", "𝄞", `\xf0\x9d\x84\x9e`),
		Entry("an invalid UTF-8 sequence is still escaped byte by byte", "\xc3\x28", `\xc3(`),
		Entry("hexadecimal digits are lower case", "\xab\xcd\xef", `\xab\xcd\xef`),
		// The bytes Go's %q would write as \a, \b, \f and \v get the same \xHH treatment as any other byte, which
		// keeps the set of escape sequences a runner has to reproduce as small as possible.
		Entry("the bytes with a Go specific escape sequence", "\a\b\f\v", `\x07\x08\x0c\x0b`),
		Entry("escapes mixed with printable bytes", "a\tb\nc\\d\x00e", `a\tb\nc\\d\x00e`),
	)

	It("escapes no two bytes to the same text", func() {
		// A trace has to say unambiguously which bytes a rule matched, so no two lexemes may escape to the same text.
		// Every escape sequence is self delimiting, which reduces that to the 256 single bytes being distinct.
		seen := make(map[string]int, 256)
		for value := range 256 {
			escaped := backendtest.EscapeLexeme(string([]byte{byte(value)}))

			Expect(seen).ToNot(HaveKey(escaped), "bytes %#02x and %#02x both escape to %q", seen[escaped], value, escaped)
			seen[escaped] = value
		}
	})

	It("escapes a lexeme byte by byte", func() {
		// Escaping a whole lexeme has to give the same result as escaping its bytes one at a time. A driver which
		// looped over runes instead of bytes would pass the table above and fail here.
		allBytes := make([]byte, 0, 256)
		var byteByByte strings.Builder
		for value := range 256 {
			allBytes = append(allBytes, byte(value))
			byteByByte.WriteString(backendtest.EscapeLexeme(string([]byte{byte(value)})))
		}

		Expect(backendtest.EscapeLexeme(string(allBytes))).To(Equal(byteByByte.String()))
	})
})
