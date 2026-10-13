package protocol_test

import (
	"bytes"
	"encoding/json"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/protocol"
)

// tree returns the JSON text as the tree the decode functions take.
func tree(text string) any {
	decoder := json.NewDecoder(bytes.NewReader([]byte(text)))
	decoder.UseNumber()
	var value any
	Expect(decoder.Decode(&value)).To(Succeed())
	return value
}

// expectRoundTrip decodes the JSON text with the decode function and expects the encoded result to be equal to it.
func expectRoundTrip[T any](decode func(any) (T, error), text string) {
	decoded, err := decode(tree(text))
	Expect(err).ToNot(HaveOccurred())
	Expect(json.Marshal(decoded)).To(MatchJSON(text))
}

var _ = Describe("Decode", func() {
	It("should decode a structure", func() {
		Expect(protocol.DecodeTextEdit(tree(`{
			"range": {"start": {"line": 1, "character": 2}, "end": {"line": 3, "character": 4}},
			"newText": "text"
		}`))).To(Equal(protocol.TextEdit{
			Range: protocol.Range{
				Start: protocol.Position{Line: 1, Character: 2},
				End:   protocol.Position{Line: 3, Character: 4},
			},
			NewText: "text",
		}))
	})

	It("should decode the inherited properties", func() {
		Expect(protocol.DecodeTextDocumentPositionParams(tree(`{
			"textDocument": {"uri": "file:///a.golr"},
			"position": {"line": 1, "character": 2}
		}`))).To(Equal(protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{Uri: "file:///a.golr"},
			Position:     protocol.Position{Line: 1, Character: 2},
		}))
	})

	DescribeTable("should report the path to a value of the wrong kind",
		func(decode func(any) error, text string, expected string) {
			Expect(decode(tree(text))).To(MatchError(expected))
		},
		Entry("value", decodeError(protocol.DecodePosition), `[]`, "expected object, got array"),
		Entry("property", decodeError(protocol.DecodeTextEdit), `{"range": {"start": {"line": "1"}}}`,
			"range.start.line: expected integer, got string"),
		Entry("array element", decodeError(protocol.DecodeTextDocumentEdit), `{"edits": [
			{"range": {"start": {"line": 0, "character": 0}, "end": {"line": 0, "character": 0}}, "newText": ""},
			{"range": {"start": {"line": 0, "character": 0}, "end": {"line": 0, "character": 0}}, "newText": 1}
		]}`, "edits[1]: expected TextEdit, AnnotatedTextEdit or "+
			"SnippetTextEdit, got object"),
		Entry("map entry", decodeError(protocol.DecodeWorkspaceEdit), `{"changes": {"file:///a.golr": [{"newText": true}]}}`,
			`changes["file:///a.golr"][0].newText: expected string, got boolean`),
		Entry("fraction", decodeError(protocol.DecodePosition), `{"line": 1.5}`, "line: expected integer, got number 1.5"),
		Entry("literal value", decodeError(protocol.DecodeCreateFile), `{"kind": "delete", "uri": "file:///a.golr"}`,
			`kind: expected "create", got "delete"`),
		Entry("tuple length", decodeError(protocol.DecodeParameterInformation), `{"label": [1, 2, 3]}`,
			"label: expected string or [2]int, got array"),
	)

	It("should decode a decimal", func() {
		Expect(protocol.DecodeColor(tree(`{"red": 0.5, "green": 1, "blue": 0, "alpha": 0.25}`))).To(Equal(
			protocol.Color{Red: 0.5, Green: 1, Blue: 0, Alpha: 0.25},
		))
	})

	It("should decode a tuple", func() {
		information, err := protocol.DecodeParameterInformation(tree(`{"label": [1, 2]}`))
		Expect(err).ToNot(HaveOccurred())
		Expect(information.Label.Value()).To(Equal([2]int{1, 2}))
	})

	It("should keep an LSPAny as decoded", func() {
		Expect(protocol.DecodeExecuteCommandParams(tree(`{"command": "c", "arguments": [1, "a", {"b": null}]}`))).To(Equal(
			protocol.ExecuteCommandParams{
				Command:   "c",
				Arguments: []protocol.LSPAny{json.Number("1"), "a", map[string]any{"b": nil}},
			},
		))
	})

	Describe("leniency", func() {
		It("should accept unknown enumeration values", func() {
			Expect(protocol.DecodeMarkupKind(tree(`"html"`))).To(Equal(protocol.MarkupKind("html")))
			Expect(protocol.DecodeCompletionItemKind(tree(`99`))).To(Equal(protocol.CompletionItemKind(99)))
		})

		It("should treat null in an optional property which is not nullable as absent", func() {
			hover, err := protocol.DecodeHover(tree(`{"contents": "text", "range": null}`))
			Expect(err).ToNot(HaveOccurred())
			Expect(hover.Range.IsZero()).To(BeTrue())
		})

		It("should leave the zero value for a missing required property", func() {
			Expect(protocol.DecodePosition(tree(`{"line": 1}`))).To(Equal(protocol.Position{Line: 1}))
		})

		It("should decode null in a required array to nil", func() {
			Expect(protocol.DecodeConfigurationParams(tree(`{"items": null}`))).To(Equal(protocol.ConfigurationParams{}))
		})

		It("should ignore unknown properties", func() {
			Expect(protocol.DecodePosition(tree(`{"line": 1, "character": 2, "offset": 3}`))).To(Equal(
				protocol.Position{Line: 1, Character: 2},
			))
		})
	})

	Describe("optional and nullable properties", func() {
		DescribeTable("should keep absent, null and a value apart",
			func(text string, expectedSet bool, expected []protocol.WorkspaceFolder) {
				params, err := protocol.DecodeInitializeParams(tree(text))
				Expect(err).ToNot(HaveOccurred())
				folders, set := params.WorkspaceFolders.Get()
				Expect(set).To(Equal(expectedSet))
				Expect(folders).To(Equal(expected))
			},
			Entry("absent", `{}`, false, nil),
			Entry("null", `{"workspaceFolders": null}`, true, nil),
			Entry("value", `{"workspaceFolders": [{"uri": "file:///a", "name": "a"}]}`, true,
				[]protocol.WorkspaceFolder{{Uri: "file:///a", Name: "a"}}),
		)

		It("should decode a nullable value to a pointer", func() {
			params, err := protocol.DecodeInitializeParams(tree(`{"processId": null, "rootUri": "file:///a"}`))
			Expect(err).ToNot(HaveOccurred())
			Expect(params.ProcessId).To(BeNil())
			Expect(params.RootUri).To(HaveValue(Equal(protocol.DocumentUri("file:///a"))))
		})

		It("should decode the structure containing itself", func() {
			selectionRange, err := protocol.DecodeSelectionRange(tree(`{
				"range": {"start": {"line": 1, "character": 0}, "end": {"line": 1, "character": 1}},
				"parent": {"range": {"start": {"line": 0, "character": 0}, "end": {"line": 2, "character": 0}}}
			}`))
			Expect(err).ToNot(HaveOccurred())
			Expect(selectionRange.Parent).To(HaveValue(Equal(protocol.SelectionRange{Range: protocol.Range{
				End: protocol.Position{Line: 2},
			}})))
		})
	})

	Describe("unions", func() {
		DescribeTable("should choose the arm",
			func(decode func(any) (any, error), text string, expected any) {
				Expect(decode(tree(text))).To(Equal(expected))
			},
			Entry("base type", unionValue(protocol.DecodeProgressToken), `7`, 7),
			Entry("other base type", unionValue(protocol.DecodeProgressToken), `"token"`, "token"),
			Entry("more required properties first", unionValue(protocol.DecodeTextDocumentEditEdits),
				`{"range": {"start": {"line": 0, "character": 0}, "end": {"line": 0, "character": 0}},
				"newText": "", "annotationId": "a"}`,
				protocol.AnnotatedTextEdit{NewText: "", AnnotationId: "a"}),
			Entry("missing required property rules out the arm", unionValue(protocol.DecodeWorkspaceSymbolLocation),
				`{"uri": "file:///a.golr"}`, protocol.LocationUriOnly{Uri: "file:///a.golr"}),
			Entry("literal value rules out the arm", unionValue(protocol.DecodeWorkspaceEditDocumentChanges),
				`{"kind": "delete", "uri": "file:///a.golr"}`,
				protocol.DeleteFile{Kind: "delete", Uri: "file:///a.golr"}),
			Entry("kind of a property rules out the arm", unionValue(protocol.DecodeCodeActionRequestResult),
				`{"title": "t", "command": {"title": "c", "command": "run"}}`,
				protocol.CodeAction{Title: "t", Command: protocol.NewOptional(protocol.Command{
					Title:   "c",
					Command: "run",
				})}),
			Entry("first element of an array", unionValue(protocol.DecodeDefinitionRequestPartialResult),
				`[{"uri": "file:///a.golr", "range": {"start": {"line": 0, "character": 0},
				"end": {"line": 0, "character": 0}}}]`,
				[]protocol.Location{{Uri: "file:///a.golr"}}),
			Entry("empty array", unionValue(protocol.DecodeDefinitionRequestPartialResult),
				`[]`, []protocol.DefinitionLink{}),
			Entry("empty literal", unionValue(protocol.DecodeSemanticTokensOptionsRange), `{"unknown": 1}`, struct{}{}),
		)

		It("should decode null to the zero value", func() {
			Expect(protocol.DecodeDefinitionRequestResult(nil)).To(Equal(protocol.DefinitionRequestResult{}))
		})

		It("should report a value which fits no arm", func() {
			_, err := protocol.DecodeProgressToken(tree(`true`))
			Expect(err).To(MatchError("expected int or string, got boolean"))
		})
	})

	DescribeTable("should encode a decoded message as it was",
		expectRoundTrip[protocol.ReferenceParams],
		Entry("tokens", protocol.DecodeReferenceParams, `{
			"textDocument": {"uri": "file:///folder/file.golr"},
			"position": {"line": 9, "character": 5},
			"context": {"includeDeclaration": true},
			"workDoneToken": "1d546990-40a3-4b77-b134-46622995f6ae",
			"partialResultToken": 7
		}`),
		Entry("no tokens", protocol.DecodeReferenceParams, `{
			"textDocument": {"uri": "file:///folder/file.golr"},
			"position": {"line": 0, "character": 0},
			"context": {"includeDeclaration": false}
		}`),
	)

	It("should encode decoded unions as they were", func() {
		expectRoundTrip(protocol.DecodeWorkspaceEdit, `{
			"documentChanges": [
				{"textDocument": {"uri": "file:///a.golr", "version": null}, "edits": [
					{"range": {"start": {"line": 0, "character": 0}, "end": {"line": 0, "character": 1}}, "newText": "a"},
					{"range": {"start": {"line": 1, "character": 0}, "end": {"line": 1, "character": 1}}, "newText": "b",
						"annotationId": "x"}
				]},
				{"kind": "create", "uri": "file:///b.golr", "options": {"overwrite": false}},
				{"kind": "rename", "oldUri": "file:///b.golr", "newUri": "file:///c.golr"},
				{"kind": "delete", "uri": "file:///c.golr"}
			],
			"changeAnnotations": {"x": {"label": "rename", "needsConfirmation": true}}
		}`)
		expectRoundTrip(protocol.DecodeServerCapabilities, `{
			"positionEncoding": "utf-16",
			"textDocumentSync": 1,
			"referencesProvider": {"workDoneProgress": true},
			"definitionProvider": true,
			"semanticTokensProvider": {"legend": {"tokenTypes": [], "tokenModifiers": []}, "range": {}, "full": true}
		}`)
		expectRoundTrip(protocol.DecodeHover, `{"contents": [{"language": "golr", "value": "a"}, "b"]}`)
	})

	It("should keep a required null apart from an absent optional value", func() {
		expectRoundTrip(protocol.DecodeVersionedTextDocumentIdentifier, `{"uri": "file:///a.golr", "version": 1}`)
		expectRoundTrip(protocol.DecodeOptionalVersionedTextDocumentIdentifier, `{"uri": "file:///a.golr", "version": null}`)
		expectRoundTrip(protocol.DecodeInitializeParams, `{"processId": null, "rootUri": null, "capabilities": {},
			"workspaceFolders": null}`)
	})
})

// unionValue returns a decode function giving the value of the arm of a union.
func unionValue[T interface{ Value() any }](decode func(any) (T, error)) func(any) (any, error) {
	return func(value any) (any, error) {
		decoded, err := decode(value)
		if err != nil {
			return nil, err
		}
		return decoded.Value(), nil
	}
}

// decodeError returns a decode function giving only the error.
func decodeError[T any](decode func(any) (T, error)) func(any) error {
	return func(value any) error {
		_, err := decode(value)
		return err
	}
}

// ambiguousArm is an arm of a union which an earlier arm with the same number of required properties takes. A server
// sends these values, it never decodes them.
type ambiguousArm struct {
	// entry is the description of the entry of the arm.
	entry string

	// taken is the Go type of the arm taking the value.
	taken reflect.Type
}

var ambiguousArms = []ambiguousArm{
	{
		entry: "ServerCapabilitiesNotebookDocumentSync NotebookDocumentSyncRegistrationOptions",
		taken: reflect.TypeFor[protocol.NotebookDocumentSyncOptions](),
	},
	{
		entry: "WorkspaceOptionsTextDocumentContent TextDocumentContentRegistrationOptions",
		taken: reflect.TypeFor[protocol.TextDocumentContentOptions](),
	},
	{
		entry: "WorkspaceSymbolRequestResult []WorkspaceSymbol",
		taken: reflect.TypeFor[[]protocol.SymbolInformation](),
	},
	{
		entry: "WorkspaceSymbolRequestPartialResult []WorkspaceSymbol",
		taken: reflect.TypeFor[[]protocol.SymbolInformation](),
	},
}

// expectArm decodes the JSON text with the decode function of a union and expects an arm of the Go type, or the arm
// taking it after ambiguousArms.
func expectArm(decode func(any) (any, error), text string, expected reflect.Type) {
	for _, arm := range ambiguousArms {
		if arm.entry == CurrentSpecReport().LeafNodeText {
			expected = arm.taken
		}
	}
	decoded, err := decode(tree(text))
	Expect(err).ToNot(HaveOccurred())
	Expect(reflect.TypeOf(decoded).String()).To(Equal(expected.String()))
}
