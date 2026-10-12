package server

import (
	"bytes"
	"sync"

	"github.com/backbone81/golr/internal/lsp/protocol"
	"github.com/backbone81/golr/internal/lsp/stopgap"
)

// Snapshot is an immutable version of a document. Its derived data is built on first use.
type Snapshot struct {
	URI     stopgap.DocumentUri
	Version int

	// Text has "\n" line endings only.
	Text []byte

	encoding      protocol.PositionEncodingKind
	lineTableOnce sync.Once
	lineTable     *LineTable
}

func newSnapshot(
	uri stopgap.DocumentUri,
	version int,
	text string,
	encoding protocol.PositionEncodingKind,
) *Snapshot {
	return &Snapshot{
		URI:      uri,
		Version:  version,
		Text:     normalizeLineEndings([]byte(text)),
		encoding: encoding,
	}
}

// Position returns the position of the byte offset in the encoding negotiated with the client.
func (s *Snapshot) Position(offset int) stopgap.Position {
	return s.LineTable().Position(offset, s.encoding)
}

// Offset returns the byte offset of the position in the encoding negotiated with the client.
func (s *Snapshot) Offset(position stopgap.Position) int {
	return s.LineTable().Offset(position, s.encoding)
}

// LineTable returns the line table of the text.
func (s *Snapshot) LineTable() *LineTable {
	s.lineTableOnce.Do(func() {
		s.lineTable = NewLineTable(s.Text)
	})
	return s.lineTable
}

// normalizeLineEndings turns "\r\n" and a lone "\r" into "\n". LSP ends a line at all three, so every position stays
// the same.
func normalizeLineEndings(text []byte) []byte {
	if bytes.IndexByte(text, '\r') < 0 {
		return text
	}
	text = bytes.ReplaceAll(text, []byte("\r\n"), []byte("\n"))
	return bytes.ReplaceAll(text, []byte("\r"), []byte("\n"))
}
