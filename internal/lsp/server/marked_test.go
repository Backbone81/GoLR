package server_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/protocol"
)

const (
	markedURI    = "file:///marked.golr"
	caretMarker  = "<caret>"
	rangeStart   = "«"
	rangeEnd     = "»"
	markedFormat = `{"jsonrpc":"2.0","id":2,"method":"%s","params":{"textDocument":{"uri":"` + markedURI + `"},` +
		`"position":%s%s}}`
)

// marked is a document with a caret and marked ranges, written as "<caret>" and "«...»" in its source. Positions are
// in UTF-16, the position encoding of initializeRequest.
type marked struct {
	text   string
	caret  protocol.Position
	ranges []protocol.Range
}

// newMarked returns the document of the source with the markers removed.
func newMarked(source string) marked {
	var result marked
	var text strings.Builder
	var position protocol.Position
	var start protocol.Position
	for len(source) > 0 {
		switch {
		case strings.HasPrefix(source, caretMarker):
			result.caret = position
			source = source[len(caretMarker):]
		case strings.HasPrefix(source, rangeStart):
			start = position
			source = source[len(rangeStart):]
		case strings.HasPrefix(source, rangeEnd):
			result.ranges = append(result.ranges, protocol.Range{Start: start, End: position})
			source = source[len(rangeEnd):]
		default:
			character, size := utf8.DecodeRuneInString(source)
			text.WriteString(source[:size])
			source = source[size:]
			if character == '\n' {
				position = protocol.Position{Line: position.Line + 1}
				continue
			}
			position.Character += utf16.RuneLen(character)
		}
	}
	result.text = text.String()
	return result
}

// didOpen returns the didOpen notification of the document.
func (m marked) didOpen() string {
	text, err := json.Marshal(m.text)
	Expect(err).NotTo(HaveOccurred())
	return `{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"` + markedURI +
		`","languageId":"golr","version":1,"text":` + string(text) + `}}}`
}

// positionRequest returns a request with id 2 of the method at the caret of the document.
func (m marked) positionRequest(method string) string {
	return m.request(method, "")
}

// referencesRequest returns a references request with id 2 at the caret of the document.
func (m marked) referencesRequest(includeDeclaration bool) string {
	return m.request(protocol.MethodTextDocumentReferences,
		fmt.Sprintf(`,"context":{"includeDeclaration":%t}`, includeDeclaration))
}

// renameRequest returns a rename request with id 2 at the caret of the document.
func (m marked) renameRequest(newName string) string {
	name, err := json.Marshal(newName)
	Expect(err).NotTo(HaveOccurred())
	return m.request(protocol.MethodTextDocumentRename, `,"newName":`+string(name))
}

// request returns a request with id 2 of the method at the caret of the document, with the extra members appended to
// its params.
func (m marked) request(method string, members string) string {
	position, err := json.Marshal(m.caret)
	Expect(err).NotTo(HaveOccurred())
	return fmt.Sprintf(markedFormat, method, position, members)
}

// locationsResponse returns the response with id 2 which has a location for every marked range.
func (m marked) locationsResponse() string {
	locations := make([]protocol.Location, 0, len(m.ranges))
	for _, markedRange := range m.ranges {
		locations = append(locations, protocol.Location{Uri: markedURI, Range: markedRange})
	}
	return m.response(locations)
}

// rangeResponse returns the response with id 2 which has the only marked range.
func (m marked) rangeResponse() string {
	Expect(m.ranges).To(HaveLen(1))
	return m.response(m.ranges[0])
}

// renameResponse returns the response with id 2 which replaces every marked range with the new name.
func (m marked) renameResponse(newName string) string {
	edits := make([]protocol.TextEdit, 0, len(m.ranges))
	for _, markedRange := range m.ranges {
		edits = append(edits, protocol.TextEdit{Range: markedRange, NewText: newName})
	}
	return m.response(protocol.WorkspaceEdit{
		Changes: map[protocol.DocumentUri][]protocol.TextEdit{markedURI: edits},
	})
}

// response returns the response with id 2 which has the result.
func (m marked) response(result any) string {
	content, err := json.Marshal(result)
	Expect(err).NotTo(HaveOccurred())
	return `{"jsonrpc":"2.0","id":2,"result":` + string(content) + `}`
}
