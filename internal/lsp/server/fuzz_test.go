package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/backbone81/golr/internal/lsp/base"
	"github.com/backbone81/golr/internal/lsp/jsonrpc"
	"github.com/backbone81/golr/internal/lsp/lsptest"
	"github.com/backbone81/golr/internal/lsp/protocol"
	"github.com/backbone81/golr/internal/lsp/server"
)

// contentSeparator separates the message contents in the input of FuzzSession. The fuzzer works on the contents, the
// test writes the header with the correct Content-Length.
const contentSeparator = "\x1e"

// fuzzSessionTimeout is how long a fuzzed session may run before it counts as hanging.
const fuzzSessionTimeout = 10 * time.Second

// FuzzSession runs a session on arbitrary message contents. No handler may panic, no response may be an internal
// error, the output has to be well formed, and every result has to decode as the result type of its method.
func FuzzSession(f *testing.F) {
	for _, session := range [][]string{
		{
			initializeRequest, `{"jsonrpc":"2.0","method":"initialized","params":{}}`, didOpenUnformattedNotification,
			formattingRequest, shutdownRequest, exitNotification,
		},
		{
			initializeRequest, didOpenUnformattedNotification, didChangeToFormattedNotification, formattingRequest,
			didCloseNotification, formattingRequest, shutdownRequest, exitNotification,
		},
		{
			`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"processId":null,"capabilities":{
				"general":{"positionEncodings":["utf-16","utf-8"]}}}}`,
			`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{
				"uri":"file:///a.golr","languageId":"golr","version":1,
				"text":"@scanner{A:\"a\";}@parser{s:A;} // \ud83d\ude00"}}}`,
			formattingRequest, shutdownRequest, exitNotification,
		},
		{
			formattingRequest, initializeRequest, initializeRequest, didOpenUnformattedNotification,
			`{"jsonrpc":"2.0","method":"$/cancelRequest","params":{"id":2}}`,
			`{"jsonrpc":"2.0","id":"x","method":"$/unknown"}`,
			`{"jsonrpc":"2.0","method":"textDocument/formatting","params":{"textDocument":{"uri":"file:///a.golr"}}}`,
			unknownRequest, shutdownRequest, formattingRequest, exitNotification,
		},
		{
			initializeRequest,
			`{"jsonrpc":"2.0","id":3,"method":"textDocument/didOpen","params":{"textDocument":{
				"uri":"file:///a.golr","version":1,"text":""}}}`,
			`{"jsonrpc":"2.0","method":"textDocument/didChange","params":{"textDocument":{"uri":"file:///a.golr","version":2},
				"contentChanges":[{"range":{"start":{"line":0,"character":0},"end":{"line":0,"character":0}},"text":"a"}]}}`,
			shutdownRequest, exitNotification,
		},
	} {
		f.Add([]byte(strings.Join(session, contentSeparator)))
	}

	f.Fuzz(func(t *testing.T, input []byte) {
		contents := strings.Split(string(input), contentSeparator)
		responses := fuzzRun(t, contents)

		if uniqueIDs(contents) {
			if err := lsptest.DecodableResults(contents, responses); err != nil {
				t.Fatal(err)
			}
		}
	})
}

// FuzzFormatting formats arbitrary documents. A document with a syntax error is refused, any other gets an edit
// whose text formats to no edit and has no syntax error.
func FuzzFormatting(f *testing.F) {
	paths, err := filepath.Glob("../../backendtest/golden/*/spec.golr")
	if err != nil {
		f.Fatal(err)
	}
	paths = append(paths, "../../../examples/calculator/calculator.golr")
	for _, path := range paths {
		text, err := os.ReadFile(path)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(string(text))
	}
	f.Add("@scanner{\r\nPLUS:\"+\"; // 😀\r\nINTEGER:/[0-9]+/;}@parser{s:INTEGER;}")
	f.Add("@scanner {\n    A: \"a;\n    B: \"b\";\n}\n\n@parser {\n    s\n        : A\n        ;\n}\n")

	f.Fuzz(func(t *testing.T, text string) {
		if !utf8.ValidString(text) {
			// JSON carries no invalid UTF-8, a client cannot send it.
			return
		}

		edits, refused := formatText(t, text)
		if refused || len(edits) == 0 {
			return
		}
		if len(edits) != 1 {
			t.Fatalf("%d edits", len(edits))
		}

		again, refused := formatText(t, edits[0].NewText)
		if refused {
			t.Fatalf("formatted text has a syntax error:\n%s", edits[0].NewText)
		}
		if len(again) != 0 {
			t.Fatalf("formatting is not idempotent:\n%s\nformats to\n%s", edits[0].NewText, again[0].NewText)
		}
	})
}

// FuzzLineTable converts between offsets and positions of arbitrary texts. Positions have to match the ones counted
// with the standard library, and the offset of a position is a character start within the text.
func FuzzLineTable(f *testing.F) {
	f.Add("", 0, 0)
	f.Add("a\nbc\n", 1, 1)
	f.Add("ä😀\nx😀y", 1, 2)
	f.Add("😀\n", 0, 1)
	f.Add("a\n\n", 5, -1)

	f.Fuzz(func(t *testing.T, text string, line int, character int) {
		if !utf8.ValidString(text) || len(text) > 4096 {
			return
		}
		table := server.NewLineTable([]byte(text))
		for _, encoding := range []protocol.PositionEncodingKind{
			protocol.PositionEncodingKindUTF8,
			protocol.PositionEncodingKindUTF16,
		} {
			for offset := 0; offset <= len(text); offset++ {
				if offset < len(text) && !utf8.RuneStart(text[offset]) {
					continue
				}
				want := expectedPosition(text, offset, encoding)
				if got := table.Position(offset, encoding); got != want {
					t.Fatalf("%s: position of offset %d is %+v, want %+v", encoding, offset, got, want)
				}
				if got := table.Offset(want, encoding); got != offset {
					t.Fatalf("%s: offset of %+v is %d, want %d", encoding, want, got, offset)
				}
			}

			position := protocol.Position{Line: line, Character: character}
			offset := table.Offset(position, encoding)
			if offset < 0 || offset > len(text) || (offset < len(text) && !utf8.RuneStart(text[offset])) {
				t.Fatalf("%s: offset %d of %+v is no character start", encoding, offset, position)
			}
		}
	})
}

// expectedPosition returns the position of the offset counted with the standard library.
func expectedPosition(text string, offset int, encoding protocol.PositionEncodingKind) protocol.Position {
	before := text[:offset]
	lineStart := strings.LastIndexByte(before, '\n') + 1
	character := len(before) - lineStart
	if encoding != protocol.PositionEncodingKindUTF8 {
		character = len(utf16.Encode([]rune(before[lineStart:])))
	}
	return protocol.Position{Line: strings.Count(before, "\n"), Character: character}
}

// formatText opens the text in a session and formats it. It returns the edits, or true when the server refused the
// document for a syntax error.
func formatText(t *testing.T, text string) ([]protocol.TextEdit, bool) {
	t.Helper()
	quoted, err := json.Marshal(text)
	if err != nil {
		t.Fatal(err)
	}
	responses := fuzzRun(t, []string{
		initializeRequest,
		`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{
			"uri":"file:///a.golr","languageId":"golr","version":1,"text":` + string(quoted) + `}}}`,
		formattingRequest,
		shutdownRequest,
		exitNotification,
	})
	if len(responses) != 3 {
		t.Fatalf("%d responses: %v", len(responses), responses)
	}

	var response struct {
		Result []protocol.TextEdit  `json:"result"`
		Error  *jsonrpc.ErrorObject `json:"error"`
	}
	if err := json.Unmarshal([]byte(responses[1]), &response); err != nil {
		t.Fatal(err)
	}
	if response.Error != nil {
		if response.Error.Code != int(protocol.LSPErrorCodesRequestFailed) {
			t.Fatalf("formatting failed: %s", responses[1])
		}
		return nil, true
	}
	return response.Result, false
}

// fuzzRun runs a session inline on the message contents and returns the contents the server wrote. It fails on a
// panicking handler, an internal error, broken output and a session which does not end.
func fuzzRun(t *testing.T, contents []string) []string {
	t.Helper()
	var input bytes.Buffer
	for _, content := range contents {
		if err := base.WriteMessage(&input, []byte(content)); err != nil {
			t.Fatal(err)
		}
	}

	logs := &panicDetector{}
	var output bytes.Buffer
	srv := server.New(slog.New(logs), "", false, 0)
	done := make(chan struct{})
	go func() {
		defer close(done)
		// The error only says how the session ended, which every input may do.
		_ = srv.Run(context.Background(), &input, &output)
	}()
	select {
	case <-done:
	case <-time.After(fuzzSessionTimeout):
		t.Fatalf("session did not end within %s", fuzzSessionTimeout)
	}
	if logs.panicked.Load() {
		t.Fatal("handler panicked")
	}

	parser := base.NewProtocolParser(&output)
	var responses []string
	for parser.Next() {
		content := string(parser.Content())
		var response struct {
			Error *jsonrpc.ErrorObject `json:"error"`
		}
		if err := json.Unmarshal(parser.Content(), &response); err != nil {
			t.Fatalf("response %s: %v", content, err)
		}
		if response.Error != nil && response.Error.Code == jsonrpc.ErrorCodeInternalError {
			t.Fatalf("internal error: %s", content)
		}
		responses = append(responses, content)
	}
	if err := parser.Err(); err != nil {
		t.Fatalf("output: %v", err)
	}
	return responses
}

// uniqueIDs reports whether the requests among the contents have distinct ids, without which a response cannot be
// matched to its request.
func uniqueIDs(contents []string) bool {
	seen := map[string]bool{}
	for _, content := range contents {
		var request map[string]json.RawMessage
		if json.Unmarshal([]byte(content), &request) != nil || request["id"] == nil {
			continue
		}
		if seen[string(request["id"])] {
			return false
		}
		seen[string(request["id"])] = true
	}
	return true
}

// panicDetector is a slog handler which notes a panicking handler and discards the records.
type panicDetector struct {
	panicked atomic.Bool
}

func (d *panicDetector) Enabled(context.Context, slog.Level) bool { return true }

func (d *panicDetector) Handle(_ context.Context, record slog.Record) error {
	if record.Message == "Handler panicked." {
		d.panicked.Store(true)
	}
	return nil
}

func (d *panicDetector) WithAttrs([]slog.Attr) slog.Handler { return d }

func (d *panicDetector) WithGroup(string) slog.Handler { return d }
