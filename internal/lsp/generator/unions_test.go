package generator_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/generator"
	"github.com/backbone81/golr/internal/utils"
)

// renderUnions returns the union types generated from the meta model.
func renderUnions(model *generator.MetaModel) string {
	var builder strings.Builder
	Expect(generator.UnionsFromMetaModel(&builder, model)).To(Succeed())
	return builder.String()
}

var _ = Describe("UnionsFromMetaModel", func() {
	It("should write the union type with constructor, value and encoding", func() {
		model := &generator.MetaModel{Structures: []generator.Structure{{
			Name: "CompletionItem",
			Properties: []generator.Property{
				{
					Name: "textEdit",
					Type: generator.Type{
						Kind: generator.TypeKindOr,
						Items: []generator.Type{
							{Kind: generator.TypeKindReference, Name: "TextEdit"},
							{Kind: generator.TypeKindReference, Name: "InsertReplaceEdit"},
						},
					},
					Optional: true,
				},
			},
		}}}
		Expect(renderUnions(model)).To(ContainSubstring(utils.HereDoc(`
			// CompletionItemTextEdit holds a value of type TextEdit or InsertReplaceEdit.
			type CompletionItemTextEdit struct{ value any }

			func NewCompletionItemTextEdit[T TextEdit | InsertReplaceEdit](value T) CompletionItemTextEdit {
				return CompletionItemTextEdit{value: value}
			}

			func (u CompletionItemTextEdit) Value() any { return u.value }

			func (u CompletionItemTextEdit) MarshalJSON() ([]byte, error) { return json.Marshal(u.value) }
		`)))
	})

	DescribeTable("should name the union after its place",
		func(model *generator.MetaModel, expected string) {
			Expect(renderUnions(model)).To(ContainSubstring(expected))
		},
		Entry("property", &generator.MetaModel{Structures: []generator.Structure{{
			Name: "CancelParams",
			Properties: []generator.Property{
				{Name: "id", Type: generator.Type{
					Kind: generator.TypeKindOr,
					Items: []generator.Type{
						{Kind: generator.TypeKindBase, Name: generator.BaseTypeInteger},
						{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
					},
				}},
			},
		}}}, "func NewCancelParamsId[T int | string](value T) CancelParamsId {"),
		Entry("type alias", &generator.MetaModel{TypeAliases: []generator.TypeAlias{{
			Name: "ProgressToken",
			Type: generator.Type{
				Kind: generator.TypeKindOr,
				Items: []generator.Type{
					{Kind: generator.TypeKindBase, Name: generator.BaseTypeInteger},
					{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
				},
			},
		}}}, "func NewProgressToken[T int | string](value T) ProgressToken {"),
		Entry("result", &generator.MetaModel{Requests: []generator.Request{{
			TypeName: "DefinitionRequest",
			Result: &generator.Type{
				Kind: generator.TypeKindOr,
				Items: []generator.Type{
					{Kind: generator.TypeKindReference, Name: "Definition"},
					{
						Kind:    generator.TypeKindArray,
						Element: &generator.Type{Kind: generator.TypeKindReference, Name: "DefinitionLink"},
					},
					{Kind: generator.TypeKindBase, Name: generator.BaseTypeNull},
				},
			},
		}}}, "func NewDefinitionRequestResult[T Definition | []DefinitionLink](value T) DefinitionRequestResult {"),
		Entry("partial result", &generator.MetaModel{Requests: []generator.Request{{
			TypeName: "DefinitionRequest",
			PartialResult: &generator.Type{
				Kind: generator.TypeKindOr,
				Items: []generator.Type{
					{Kind: generator.TypeKindArray, Element: &generator.Type{Kind: generator.TypeKindReference, Name: "Location"}},
					{
						Kind:    generator.TypeKindArray,
						Element: &generator.Type{Kind: generator.TypeKindReference, Name: "DefinitionLink"},
					},
				},
			},
		}}}, "func NewDefinitionRequestPartialResult[T []Location | []DefinitionLink](value T) "+
			"DefinitionRequestPartialResult {"),
		Entry("element of a nullable array", &generator.MetaModel{Requests: []generator.Request{{
			TypeName: "CodeActionRequest",
			Result: &generator.Type{
				Kind: generator.TypeKindOr,
				Items: []generator.Type{
					{Kind: generator.TypeKindArray, Element: &generator.Type{
						Kind: generator.TypeKindOr,
						Items: []generator.Type{
							{Kind: generator.TypeKindReference, Name: "Command"},
							{Kind: generator.TypeKindReference, Name: "CodeAction"},
						},
					}},
					{Kind: generator.TypeKindBase, Name: generator.BaseTypeNull},
				},
			},
		}}}, "func NewCodeActionRequestResult[T Command | CodeAction](value T) CodeActionRequestResult {"),
	)

	It("should link the section of a union type alias and mark it deprecated", func() {
		model := &generator.MetaModel{MetaData: generator.MetaData{Version: "3.18.0"}, TypeAliases: []generator.TypeAlias{{
			Annotations: generator.Annotations{Deprecated: "use MarkupContent instead"},
			Name:        "MarkedString",
			Type: generator.Type{
				Kind: generator.TypeKindOr,
				Items: []generator.Type{
					{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
					{Kind: generator.TypeKindReference, Name: "MarkedStringWithLanguage"},
				},
			},
		}}}
		Expect(renderUnions(model)).To(ContainSubstring(utils.HereDoc(`
			// MarkedString holds a value of type string or MarkedStringWithLanguage.
			//
			// See https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification/#markedString.
			//
			// Deprecated: deprecated by the Language Server Protocol.
			type MarkedString struct{ value any }
		`)))
	})

	It("should write no union type for a single item besides null or for LSPAny", func() {
		model := &generator.MetaModel{
			Structures: []generator.Structure{{
				Name: "Hover",
				Properties: []generator.Property{
					{Name: "range", Type: generator.Type{
						Kind: generator.TypeKindOr,
						Items: []generator.Type{
							{Kind: generator.TypeKindReference, Name: "Range"},
							{Kind: generator.TypeKindBase, Name: generator.BaseTypeNull},
						},
					}},
				},
			}},
			TypeAliases: []generator.TypeAlias{{
				Name: "LSPAny",
				Type: generator.Type{
					Kind: generator.TypeKindOr,
					Items: []generator.Type{
						{Kind: generator.TypeKindReference, Name: "LSPObject"},
						{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
						{Kind: generator.TypeKindBase, Name: generator.BaseTypeNull},
					},
				},
			}},
		}
		Expect(renderUnions(model)).NotTo(ContainSubstring("type "))
	})
})
