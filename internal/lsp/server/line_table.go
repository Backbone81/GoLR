package server

import (
	"slices"

	"github.com/backbone81/golr/internal/lsp/protocol"
)

// LineTable converts between byte offsets of a text with "\n" line endings and LSP positions. Columns skip UTF-8
// continuation bytes like the golr scanner does.
type LineTable struct {
	text       []byte
	lineStarts []int
}

// NewLineTable returns the line table of text.
func NewLineTable(text []byte) *LineTable {
	lineStarts := []int{0}
	for offset, value := range text {
		if value == '\n' {
			lineStarts = append(lineStarts, offset+1)
		}
	}
	return &LineTable{
		text:       text,
		lineStarts: lineStarts,
	}
}

// Position returns the position of the byte offset, which is clamped to the text.
func (t *LineTable) Position(offset int, encoding protocol.PositionEncodingKind) protocol.Position {
	offset = min(max(offset, 0), len(t.text))
	line, found := slices.BinarySearch(t.lineStarts, offset)
	if !found {
		line--
	}
	lineStart := t.lineStarts[line]

	character := offset - lineStart
	if encoding != protocol.PositionEncodingKindUTF8 {
		character = 0
		for _, value := range t.text[lineStart:offset] {
			character += t.utf16Units(value)
		}
	}
	return protocol.Position{
		Line:      uint32(line), //nolint:gosec // A line index is below the text length.
		Character: uint32(character),
	}
}

// Offset returns the byte offset of the position. A position beyond the end of its line is the end of the line, one
// inside a character is the start of the character, and a line beyond the text is the end of the text.
func (t *LineTable) Offset(position protocol.Position, encoding protocol.PositionEncodingKind) int {
	line := int(position.Line)
	if line >= len(t.lineStarts) {
		return len(t.text)
	}
	lineStart := t.lineStarts[line]
	lineEnd := len(t.text)
	if line+1 < len(t.lineStarts) {
		lineEnd = t.lineStarts[line+1] - 1
	}

	character := int(position.Character)
	if encoding == protocol.PositionEncodingKindUTF8 {
		offset := min(lineStart+character, lineEnd)
		for offset > lineStart && t.isContinuationByte(t.text[offset]) {
			offset--
		}
		return offset
	}

	offset := lineStart
	units := 0
	for offset < lineEnd {
		width := t.utf16Units(t.text[offset])
		if units+width > character {
			break
		}
		units += width
		offset++
		for offset < lineEnd && t.isContinuationByte(t.text[offset]) {
			offset++
		}
	}
	return offset
}

// utf16Units returns the UTF-16 code units the byte adds to a column: none for a continuation byte, two for the lead
// byte of a four byte sequence, which is a surrogate pair.
func (t *LineTable) utf16Units(value byte) int {
	switch {
	case t.isContinuationByte(value):
		return 0
	case value >= 0xf0:
		return 2
	default:
		return 1
	}
}

func (t *LineTable) isContinuationByte(value byte) bool {
	return value&0xc0 == 0x80
}
