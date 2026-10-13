package server_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/protocol"
	"github.com/backbone81/golr/internal/lsp/server"
)

var _ = Describe("LineTable", func() {
	// "ä" is two bytes and one UTF-16 unit, "😀" four bytes and two UTF-16 units.
	text := []byte("ab\nä😀c\n\nlast")

	DescribeTable("converts in both directions",
		func(offset int, encoding protocol.PositionEncodingKind, line int, character int) {
			table := server.NewLineTable(text)
			position := protocol.Position{Line: line, Character: character}
			Expect(table.Position(offset, encoding)).To(Equal(position))
			Expect(table.Offset(position, encoding)).To(Equal(offset))
		},
		Entry("start of text", 0, protocol.PositionEncodingKindUTF16, 0, 0),
		Entry("end of first line", 2, protocol.PositionEncodingKindUTF16, 0, 2),
		Entry("start of second line", 3, protocol.PositionEncodingKindUTF16, 1, 0),
		Entry("after a two byte character in UTF-16", 5, protocol.PositionEncodingKindUTF16, 1, 1),
		Entry("after a two byte character in UTF-8", 5, protocol.PositionEncodingKindUTF8, 1, 2),
		Entry("after a surrogate pair in UTF-16", 9, protocol.PositionEncodingKindUTF16, 1, 3),
		Entry("after a four byte character in UTF-8", 9, protocol.PositionEncodingKindUTF8, 1, 6),
		Entry("empty line", 11, protocol.PositionEncodingKindUTF16, 2, 0),
		Entry("end of text without final newline", 16, protocol.PositionEncodingKindUTF16, 3, 4),
	)

	It("clamps an offset outside the text", func() {
		table := server.NewLineTable(text)
		Expect(table.Position(-1, protocol.PositionEncodingKindUTF16)).To(Equal(protocol.Position{}))
		Expect(table.Position(100, protocol.PositionEncodingKindUTF16)).To(Equal(protocol.Position{Line: 3, Character: 4}))
	})

	It("has an empty last line after a final newline", func() {
		table := server.NewLineTable([]byte("a\n"))
		Expect(table.Position(2, protocol.PositionEncodingKindUTF16)).To(Equal(protocol.Position{Line: 1, Character: 0}))
	})

	DescribeTable("falls back on positions which are no character boundary",
		func(position protocol.Position, encoding protocol.PositionEncodingKind, offset int) {
			Expect(server.NewLineTable(text).Offset(position, encoding)).To(Equal(offset))
		},
		Entry("beyond the line end", protocol.Position{Line: 0, Character: 10}, protocol.PositionEncodingKindUTF16, 2),
		Entry("beyond the line end in UTF-8",
			protocol.Position{Line: 0, Character: 10}, protocol.PositionEncodingKindUTF8, 2),
		Entry("beyond the last line", protocol.Position{Line: 10, Character: 0}, protocol.PositionEncodingKindUTF16, 16),
		Entry("inside a surrogate pair", protocol.Position{Line: 1, Character: 2}, protocol.PositionEncodingKindUTF16, 5),
		Entry("inside a UTF-8 sequence", protocol.Position{Line: 1, Character: 4}, protocol.PositionEncodingKindUTF8, 5),
		Entry("negative line", protocol.Position{Line: -1, Character: 1}, protocol.PositionEncodingKindUTF16, 1),
		Entry("negative character", protocol.Position{Line: 1, Character: -1}, protocol.PositionEncodingKindUTF16, 3),
		Entry("negative character in UTF-8",
			protocol.Position{Line: 1, Character: -1}, protocol.PositionEncodingKindUTF8, 3),
	)
})
