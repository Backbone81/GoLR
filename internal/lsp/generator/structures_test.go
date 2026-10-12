package generator_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/generator"
	"github.com/backbone81/golr/internal/utils"
)

// withTags returns the snippet in HereDoc form with ' replaced by `, so struct tags can be written in it.
func withTags(snippet string) string {
	return strings.ReplaceAll(utils.HereDoc(snippet), "'", "`")
}

// renderStructures returns the structures and type aliases generated from the meta model.
func renderStructures(model *generator.MetaModel) string {
	var builder strings.Builder
	Expect(generator.StructuresFromMetaModel(&builder, model)).To(Succeed())
	return builder.String()
}

// renderProperty returns the structures generated for a structure Item with the property label.
func renderProperty(input generator.Type, optional bool) string {
	return renderStructures(&generator.MetaModel{Structures: []generator.Structure{{
		Name:       "Item",
		Properties: []generator.Property{{Name: "label", Type: input, Optional: optional}},
	}}})
}

var _ = Describe("StructuresFromMetaModel", func() {
	DescribeTable("should write the field of a required property",
		func(input generator.Type, expected string) {
			Expect(renderProperty(input, false)).To(ContainSubstring(expected))
		},
		Entry("string", generator.Type{
			Kind: generator.TypeKindBase,
			Name: generator.BaseTypeString,
		}, "Label string `json:\"label\"`"),
		Entry("boolean", generator.Type{
			Kind: generator.TypeKindBase,
			Name: generator.BaseTypeBoolean,
		}, "Label bool `json:\"label\"`"),
		Entry("integer", generator.Type{
			Kind: generator.TypeKindBase,
			Name: generator.BaseTypeInteger,
		}, "Label int `json:\"label\"`"),
		Entry("uinteger", generator.Type{
			Kind: generator.TypeKindBase,
			Name: generator.BaseTypeUinteger,
		}, "Label int `json:\"label\"`"),
		Entry("decimal", generator.Type{
			Kind: generator.TypeKindBase,
			Name: generator.BaseTypeDecimal,
		}, "Label float64 `json:\"label\"`"),
		Entry("DocumentUri", generator.Type{
			Kind: generator.TypeKindBase,
			Name: generator.BaseTypeDocumentUri,
		}, "Label DocumentUri `json:\"label\"`"),
		Entry("unsupported base type", generator.Type{
			Kind: generator.TypeKindBase,
			Name: generator.BaseTypeRegExp,
		}, "Label RegExp `json:\"label\"`"),
		Entry("reference", generator.Type{
			Kind: generator.TypeKindReference,
			Name: "Range",
		}, "Label Range `json:\"label\"`"),
		Entry("array", generator.Type{
			Kind:    generator.TypeKindArray,
			Element: &generator.Type{Kind: generator.TypeKindReference, Name: "Range"},
		}, "Label []Range `json:\"label\"`"),
		Entry("map", generator.Type{
			Kind: generator.TypeKindMap,
			Key:  &generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeDocumentUri},
			Value: &generator.Type{
				Kind:    generator.TypeKindArray,
				Element: &generator.Type{Kind: generator.TypeKindReference, Name: "TextEdit"},
			},
		}, "Label map[DocumentUri][]TextEdit `json:\"label\"`"),
		Entry("tuple of one type", generator.Type{
			Kind: generator.TypeKindTuple,
			Items: []generator.Type{
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeUinteger},
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeUinteger},
			},
		}, "Label [2]int `json:\"label\"`"),
		Entry("tuple of mixed types", generator.Type{
			Kind: generator.TypeKindTuple,
			Items: []generator.Type{
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeUinteger},
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
			},
		}, "Label tuple `json:\"label\"`"),
		Entry("empty literal", generator.Type{
			Kind:    generator.TypeKindLiteral,
			Literal: &generator.StructureLiteral{},
		}, "Label struct{} `json:\"label\"`"),
		Entry("literal with properties", generator.Type{
			Kind: generator.TypeKindLiteral,
			Literal: &generator.StructureLiteral{Properties: []generator.Property{
				{Name: "line", Type: generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeInteger}},
			}},
		}, "Label literal `json:\"label\"`"),
		Entry("string literal", generator.Type{
			Kind:        generator.TypeKindStringLiteral,
			StringValue: "create",
		}, "Label string `json:\"label\"`"),
		Entry("unsupported kind", generator.Type{
			Kind: generator.TypeKindAnd,
			Items: []generator.Type{
				{Kind: generator.TypeKindReference, Name: "Range"},
				{Kind: generator.TypeKindReference, Name: "Position"},
			},
		}, "Label and `json:\"label\"`"),
		Entry("union", generator.Type{
			Kind: generator.TypeKindOr,
			Items: []generator.Type{
				{Kind: generator.TypeKindReference, Name: "Range"},
				{Kind: generator.TypeKindReference, Name: "Position"},
			},
		}, "Label ItemLabel `json:\"label\"`"),
		Entry("nullable", generator.Type{
			Kind: generator.TypeKindOr,
			Items: []generator.Type{
				{Kind: generator.TypeKindReference, Name: "Range"},
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeNull},
			},
		}, "Label *Range `json:\"label\"`"),
		Entry("nullable array", generator.Type{
			Kind: generator.TypeKindOr,
			Items: []generator.Type{
				{Kind: generator.TypeKindArray, Element: &generator.Type{Kind: generator.TypeKindReference, Name: "Range"}},
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeNull},
			},
		}, "Label []Range `json:\"label\"`"),
		Entry("nullable union", generator.Type{
			Kind: generator.TypeKindOr,
			Items: []generator.Type{
				{Kind: generator.TypeKindReference, Name: "Range"},
				{Kind: generator.TypeKindReference, Name: "Position"},
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeNull},
			},
		}, "Label ItemLabel `json:\"label\"`"),
	)

	DescribeTable("should write the field of an optional property",
		func(input generator.Type, expected string) {
			Expect(renderProperty(input, true)).To(ContainSubstring(expected))
		},
		Entry("string", generator.Type{
			Kind: generator.TypeKindBase,
			Name: generator.BaseTypeString,
		}, "Label Optional[string] `json:\"label,omitzero\"`"),
		Entry("array", generator.Type{
			Kind:    generator.TypeKindArray,
			Element: &generator.Type{Kind: generator.TypeKindReference, Name: "Range"},
		}, "Label []Range `json:\"label,omitzero\"`"),
		Entry("union", generator.Type{
			Kind: generator.TypeKindOr,
			Items: []generator.Type{
				{Kind: generator.TypeKindReference, Name: "Range"},
				{Kind: generator.TypeKindReference, Name: "Position"},
			},
		}, "Label ItemLabel `json:\"label,omitzero\"`"),
		Entry("nullable", generator.Type{
			Kind: generator.TypeKindOr,
			Items: []generator.Type{
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeNull},
			},
		}, "Label Optional[*string] `json:\"label,omitzero\"`"),
		Entry("nullable array", generator.Type{
			Kind: generator.TypeKindOr,
			Items: []generator.Type{
				{Kind: generator.TypeKindArray, Element: &generator.Type{Kind: generator.TypeKindReference, Name: "Range"}},
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeNull},
			},
		}, "Label Optional[[]Range] `json:\"label,omitzero\"`"),
		Entry("reference to its own structure", generator.Type{
			Kind: generator.TypeKindReference,
			Name: "Item",
		}, "Label *Item `json:\"label,omitzero\"`"),
	)

	It("should write the absent value of a reference by the type behind a type alias", func() {
		model := &generator.MetaModel{
			Structures: []generator.Structure{{Name: "Item", Properties: []generator.Property{
				{
					Name:     "selector",
					Type:     generator.Type{Kind: generator.TypeKindReference, Name: "Selector"},
					Optional: true,
				},
				{
					Name:     "filter",
					Type:     generator.Type{Kind: generator.TypeKindReference, Name: "Filter"},
					Optional: true,
				},
				{
					Name:     "pattern",
					Type:     generator.Type{Kind: generator.TypeKindReference, Name: "Pattern"},
					Optional: true,
				},
			}}},
			TypeAliases: []generator.TypeAlias{
				{Name: "Selector", Type: generator.Type{
					Kind:    generator.TypeKindArray,
					Element: &generator.Type{Kind: generator.TypeKindReference, Name: "Filter"},
				}},
				{Name: "Filter", Type: generator.Type{
					Kind: generator.TypeKindOr,
					Items: []generator.Type{
						{Kind: generator.TypeKindReference, Name: "Range"},
						{Kind: generator.TypeKindReference, Name: "Position"},
					},
				}},
				{Name: "Pattern", Type: generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeString}},
			},
		}
		Expect(renderStructures(model)).To(ContainSubstring(withTags(`
			type Item struct {
				Selector Selector          'json:"selector,omitzero"'
				Filter   Filter            'json:"filter,omitzero"'
				Pattern  Optional[Pattern] 'json:"pattern,omitzero"'
			}
		`)))
	})

	It("should flatten inherited properties, replacing a redeclared one at its place", func() {
		model := &generator.MetaModel{Structures: []generator.Structure{
			{Name: "Operation", Properties: []generator.Property{
				{Name: "kind", Type: generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeString}},
				{
					Name:     "annotationId",
					Type:     generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
					Optional: true,
				},
			}},
			{Name: "Progress", Properties: []generator.Property{
				{
					Name:     "token",
					Type:     generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
					Optional: true,
				},
			}},
			{
				Name:    "Create",
				Extends: []generator.Type{{Kind: generator.TypeKindReference, Name: "Operation"}},
				Mixins:  []generator.Type{{Kind: generator.TypeKindReference, Name: "Progress"}},
				Properties: []generator.Property{
					{Name: "kind", Type: generator.Type{Kind: generator.TypeKindStringLiteral, StringValue: "create"}},
					{Name: "uri", Type: generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeDocumentUri}},
				},
			},
		}}
		Expect(renderStructures(model)).To(ContainSubstring(withTags(`
			type Create struct {
				Kind         string           'json:"kind"'
				AnnotationId Optional[string] 'json:"annotationId,omitzero"'
				Token        Optional[string] 'json:"token,omitzero"'
				Uri          DocumentUri      'json:"uri"'
			}
		`)))
	})

	It("should name a union of an inherited property after the declaring structure", func() {
		model := &generator.MetaModel{Structures: []generator.Structure{
			{Name: "Parent", Properties: []generator.Property{
				{Name: "id", Type: generator.Type{
					Kind: generator.TypeKindOr,
					Items: []generator.Type{
						{Kind: generator.TypeKindBase, Name: generator.BaseTypeInteger},
						{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
					},
				}},
			}},
			{Name: "Child", Extends: []generator.Type{{Kind: generator.TypeKindReference, Name: "Parent"}}},
		}}
		Expect(renderStructures(model)).To(ContainSubstring(withTags(`
			type Child struct {
				Id ParentId 'json:"id"'
			}
		`)))
	})

	It("should only flatten a structure whose name starts with _", func() {
		model := &generator.MetaModel{Structures: []generator.Structure{
			{Name: "_Params", Properties: []generator.Property{
				{Name: "processId", Type: generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeInteger}},
			}},
			{Name: "Params", Extends: []generator.Type{{Kind: generator.TypeKindReference, Name: "_Params"}}},
		}}
		output := renderStructures(model)
		Expect(output).NotTo(ContainSubstring("type _Params"))
		Expect(output).To(ContainSubstring(withTags(`
			type Params struct {
				ProcessId int 'json:"processId"'
			}
		`)))
	})

	It("should mark deprecated structures and properties", func() {
		model := &generator.MetaModel{Structures: []generator.Structure{{
			Annotations: generator.Annotations{Deprecated: "use something else"},
			Name:        "Old",
			Properties: []generator.Property{{
				Annotations: generator.Annotations{Deprecated: "use something else"},
				Name:        "label",
				Type:        generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
			}},
		}}}
		Expect(renderStructures(model)).To(ContainSubstring(withTags(`
			// Deprecated: deprecated by the Language Server Protocol.
			type Old struct {
				// Deprecated: deprecated by the Language Server Protocol.
				Label string 'json:"label"'
			}
		`)))
	})

	DescribeTable("should write a type alias",
		func(input generator.TypeAlias, expected string) {
			model := &generator.MetaModel{TypeAliases: []generator.TypeAlias{input}}
			Expect(renderStructures(model)).To(ContainSubstring(expected))
		},
		Entry("base type", generator.TypeAlias{
			Name: "Pattern",
			Type: generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
		}, "type Pattern string\n"),
		Entry("array", generator.TypeAlias{
			Name: "Selector",
			Type: generator.Type{
				Kind:    generator.TypeKindArray,
				Element: &generator.Type{Kind: generator.TypeKindReference, Name: "Filter"},
			},
		}, "type Selector []Filter\n"),
		Entry("reference", generator.TypeAlias{
			Name: "DefinitionLink",
			Type: generator.Type{Kind: generator.TypeKindReference, Name: "LocationLink"},
		}, "type DefinitionLink = LocationLink\n"),
		Entry("LSPAny", generator.TypeAlias{
			Name: "LSPAny",
			Type: generator.Type{
				Kind: generator.TypeKindOr,
				Items: []generator.Type{
					{Kind: generator.TypeKindReference, Name: "LSPObject"},
					{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
					{Kind: generator.TypeKindBase, Name: generator.BaseTypeNull},
				},
			},
		}, "type LSPAny = any\n"),
		Entry("LSPArray", generator.TypeAlias{
			Name: "LSPArray",
			Type: generator.Type{
				Kind:    generator.TypeKindArray,
				Element: &generator.Type{Kind: generator.TypeKindReference, Name: "LSPAny"},
			},
		}, "type LSPArray = []LSPAny\n"),
	)

	It("should leave a union type alias to the union types", func() {
		model := &generator.MetaModel{TypeAliases: []generator.TypeAlias{
			{Name: "Token", Type: generator.Type{
				Kind: generator.TypeKindOr,
				Items: []generator.Type{
					{Kind: generator.TypeKindBase, Name: generator.BaseTypeInteger},
					{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
				},
			}},
		}}
		Expect(renderStructures(model)).NotTo(ContainSubstring("Token"))
	})
})
