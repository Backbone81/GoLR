package generator_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/generator"
	"github.com/backbone81/golr/internal/utils"
)

// renderDecode returns the decode functions generated from the meta model.
func renderDecode(model *generator.MetaModel) string {
	var builder strings.Builder
	Expect(generator.DecodeFromMetaModel(&builder, model)).To(Succeed())
	return builder.String()
}

// renderPropertyDecode returns the decode functions generated for a structure Item with the property label.
func renderPropertyDecode(input generator.Type, optional bool) string {
	return renderDecode(&generator.MetaModel{Structures: []generator.Structure{{
		Name:       "Item",
		Properties: []generator.Property{{Name: "label", Type: input, Optional: optional}},
	}}})
}

var _ = Describe("DecodeFromMetaModel", func() {
	It("should write the decode function of a structure", func() {
		model := &generator.MetaModel{Structures: []generator.Structure{
			{
				Name: "Parent",
				Properties: []generator.Property{
					{Name: "uri", Type: generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeDocumentUri}},
				},
			},
			{
				Name:    "Child",
				Extends: []generator.Type{{Kind: generator.TypeKindReference, Name: "Parent"}},
				Properties: []generator.Property{
					{
						Name:     "label",
						Type:     generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
						Optional: true,
					},
				},
			},
		}}
		Expect(renderDecode(model)).To(ContainSubstring(utils.HereDoc(`
			func DecodeChild(value any) (Child, error) {
				var result Child
				object, err := decodeObject(value)
				if err != nil {
					return result, err
				}
				if property, ok := object["uri"]; ok {
					if result.Uri, err = DecodeDocumentUri(property); err != nil {
						return Child{}, withPathSegment(err, "uri")
					}
				}
				if property, ok := object["label"]; ok && property != nil {
					if result.Label, err = optionalOf(decodeString[string])(property); err != nil {
						return Child{}, withPathSegment(err, "label")
					}
				}
				return result, nil
			}
		`)))
	})

	It("should write the decode function of a structure without properties", func() {
		Expect(renderDecode(&generator.MetaModel{Structures: []generator.Structure{{Name: "Empty"}}})).To(
			ContainSubstring(utils.HereDoc(`
				func DecodeEmpty(value any) (Empty, error) {
					var result Empty
					_, err := decodeObject(value)
					return result, err
				}
			`)),
		)
	})

	It("should not write a decode function for a structure starting with _", func() {
		Expect(renderDecode(&generator.MetaModel{Structures: []generator.Structure{{Name: "_Hidden"}}})).
			ToNot(ContainSubstring("_Hidden"))
	})

	DescribeTable("should decode a required property",
		func(input generator.Type, expected string) {
			Expect(renderPropertyDecode(input, false)).To(ContainSubstring(
				`if property, ok := object["label"]; ok {` + "\n\t\tif result.Label, err = " + expected + "(property);",
			))
		},
		Entry("string", generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
			"decodeString[string]"),
		Entry("boolean", generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeBoolean},
			"decodeBoolean"),
		Entry("uinteger", generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeUinteger},
			"decodeInteger[int]"),
		Entry("decimal", generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeDecimal},
			"decodeDecimal"),
		Entry("URI", generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeURI}, "DecodeURI"),
		Entry("unsupported base type", generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeRegExp},
			"RegExp"),
		Entry("reference", generator.Type{Kind: generator.TypeKindReference, Name: "Range"}, "DecodeRange"),
		Entry("array", generator.Type{
			Kind:    generator.TypeKindArray,
			Element: &generator.Type{Kind: generator.TypeKindReference, Name: "TextEdit"},
		}, "arrayOf(DecodeTextEdit)"),
		Entry("map", generator.Type{
			Kind: generator.TypeKindMap,
			Key:  &generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeDocumentUri},
			Value: &generator.Type{
				Kind:    generator.TypeKindArray,
				Element: &generator.Type{Kind: generator.TypeKindReference, Name: "TextEdit"},
			},
		}, "mapOf[DocumentUri](arrayOf(DecodeTextEdit))"),
		Entry("pair", generator.Type{
			Kind: generator.TypeKindTuple,
			Items: []generator.Type{
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeUinteger},
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeUinteger},
			},
		}, "pairOf(decodeInteger[int])"),
		Entry("tuple of three", generator.Type{
			Kind: generator.TypeKindTuple,
			Items: []generator.Type{
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeUinteger},
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeUinteger},
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeUinteger},
			},
		}, "tuple"),
		Entry("empty literal", generator.Type{
			Kind:    generator.TypeKindLiteral,
			Literal: &generator.StructureLiteral{},
		}, "decodeEmptyObject"),
		Entry("string literal", generator.Type{Kind: generator.TypeKindStringLiteral, StringValue: "create"},
			`stringLiteralOf("create")`),
		Entry("unsupported kind", generator.Type{Kind: generator.TypeKindBooleanLiteral}, "booleanLiteral"),
		Entry("union", generator.Type{
			Kind: generator.TypeKindOr,
			Items: []generator.Type{
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
				{Kind: generator.TypeKindReference, Name: "MarkupContent"},
			},
		}, "DecodeItemLabel"),
		Entry("nullable", generator.Type{
			Kind: generator.TypeKindOr,
			Items: []generator.Type{
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeInteger},
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeNull},
			},
		}, "pointerOf(decodeInteger[int])"),
		Entry("nullable array", generator.Type{
			Kind: generator.TypeKindOr,
			Items: []generator.Type{
				{Kind: generator.TypeKindArray, Element: &generator.Type{Kind: generator.TypeKindReference, Name: "Range"}},
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeNull},
			},
		}, "arrayOf(DecodeRange)"),
	)

	DescribeTable("should decode an optional property",
		func(input generator.Type, condition string, expected string) {
			Expect(renderPropertyDecode(input, true)).To(ContainSubstring(
				`if property, ok := object["label"]; ` + condition + " {\n\t\tif result.Label, err = " + expected +
					"(property);",
			))
		},
		Entry("string", generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
			"ok && property != nil", "optionalOf(decodeString[string])"),
		Entry("array", generator.Type{
			Kind:    generator.TypeKindArray,
			Element: &generator.Type{Kind: generator.TypeKindReference, Name: "Range"},
		}, "ok && property != nil", "arrayOf(DecodeRange)"),
		Entry("union", generator.Type{
			Kind: generator.TypeKindOr,
			Items: []generator.Type{
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
				{Kind: generator.TypeKindReference, Name: "MarkupContent"},
			},
		}, "ok && property != nil", "DecodeItemLabel"),
		Entry("nullable", generator.Type{
			Kind: generator.TypeKindOr,
			Items: []generator.Type{
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeNull},
			},
		}, "ok", "optionalOf(pointerOf(decodeString[string]))"),
		Entry("nullable array", generator.Type{
			Kind: generator.TypeKindOr,
			Items: []generator.Type{
				{Kind: generator.TypeKindArray, Element: &generator.Type{Kind: generator.TypeKindReference, Name: "Range"}},
				{Kind: generator.TypeKindBase, Name: generator.BaseTypeNull},
			},
		}, "ok", "optionalOf(arrayOf(DecodeRange))"),
		Entry("reference to its own structure", generator.Type{Kind: generator.TypeKindReference, Name: "Item"},
			"ok && property != nil", "pointerOf(DecodeItem)"),
	)

	It("should write the decode function of an enumeration", func() {
		model := &generator.MetaModel{Enumerations: []generator.Enumeration{
			{Name: "MarkupKind", Type: generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeString}},
			{Name: "SymbolKind", Type: generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeUinteger}},
			{Name: "Odd", Type: generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeDecimal}},
		}}
		Expect(renderDecode(model)).To(And(
			ContainSubstring(utils.HereDoc(`
				func DecodeMarkupKind(value any) (MarkupKind, error) {
					return decodeString[MarkupKind](value)
				}
			`)),
			ContainSubstring("return decodeInteger[SymbolKind](value)"),
			ContainSubstring("return decimal[Odd](value)"),
		))
	})

	It("should write the decode function of a type alias", func() {
		model := &generator.MetaModel{TypeAliases: []generator.TypeAlias{
			{Name: "Pattern", Type: generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeString}},
			{Name: "DefinitionLink", Type: generator.Type{Kind: generator.TypeKindReference, Name: "LocationLink"}},
			{Name: "LSPAny", Type: generator.Type{
				Kind: generator.TypeKindOr,
				Items: []generator.Type{
					{Kind: generator.TypeKindReference, Name: "LSPObject"},
					{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
				},
			}},
		}}
		Expect(renderDecode(model)).To(And(
			ContainSubstring(utils.HereDoc(`
				func DecodePattern(value any) (Pattern, error) {
					decoded, err := decodeString[string](value)
					return Pattern(decoded), err
				}
			`)),
			ContainSubstring(utils.HereDoc(`
				func DecodeDefinitionLink(value any) (DefinitionLink, error) {
					return DecodeLocationLink(value)
				}
			`)),
			ContainSubstring(utils.HereDoc(`
				func DecodeLSPAny(value any) (LSPAny, error) {
					return decodeAny(value)
				}
			`)),
		))
	})

	It("should try the arms of a union by their number of required properties", func() {
		model := &generator.MetaModel{
			Structures: []generator.Structure{
				{
					Name: "TextEdit",
					Properties: []generator.Property{
						{Name: "range", Type: generator.Type{Kind: generator.TypeKindReference, Name: "Range"}},
						{Name: "newText", Type: generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeString}},
					},
				},
				{
					Name:    "AnnotatedTextEdit",
					Extends: []generator.Type{{Kind: generator.TypeKindReference, Name: "TextEdit"}},
					Properties: []generator.Property{
						{Name: "annotationId", Type: generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeString}},
						{
							Name:     "label",
							Type:     generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
							Optional: true,
						},
					},
				},
			},
			TypeAliases: []generator.TypeAlias{{Name: "Edit", Type: generator.Type{
				Kind: generator.TypeKindOr,
				Items: []generator.Type{
					{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
					{Kind: generator.TypeKindReference, Name: "TextEdit"},
					{
						Kind:    generator.TypeKindArray,
						Element: &generator.Type{Kind: generator.TypeKindReference, Name: "AnnotatedTextEdit"},
					},
					{Kind: generator.TypeKindBase, Name: generator.BaseTypeNull},
				},
			}}},
		}
		Expect(renderDecode(model)).To(ContainSubstring(utils.HereDoc(`
			func DecodeEdit(value any) (Edit, error) {
				if value == nil {
					return Edit{}, nil
				}
				if arm, ok := decodeArrayArm(value, DecodeAnnotatedTextEdit, "range", "newText", "annotationId"); ok {
					return NewEdit(arm), nil
				}
				if arm, ok := decodeArm(value, DecodeTextEdit, "range", "newText"); ok {
					return NewEdit(arm), nil
				}
				if arm, ok := decodeArm(value, decodeString[string]); ok {
					return NewEdit(arm), nil
				}
				return Edit{}, kindError("string, TextEdit or []AnnotatedTextEdit", value)
			}
		`)))
	})
})

var _ = Describe("UnionArmsTestFromMetaModel", func() {
	It("should write an entry with a sample for every arm of a union", func() {
		model := &generator.MetaModel{
			Structures: []generator.Structure{
				{
					Name: "CreateFile",
					Properties: []generator.Property{
						{Name: "kind", Type: generator.Type{Kind: generator.TypeKindStringLiteral, StringValue: "create"}},
						{Name: "uri", Type: generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeDocumentUri}},
						{Name: "severity", Type: generator.Type{Kind: generator.TypeKindReference, Name: "Severity"}},
						{
							Name:     "label",
							Type:     generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeString},
							Optional: true,
						},
					},
				},
			},
			Enumerations: []generator.Enumeration{{
				Name:   "Severity",
				Type:   generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeUinteger},
				Values: []generator.EnumerationEntry{{Name: "Error", Value: []byte("1")}, {Name: "Warning", Value: []byte("2")}},
			}},
			TypeAliases: []generator.TypeAlias{{Name: "Change", Type: generator.Type{
				Kind: generator.TypeKindOr,
				Items: []generator.Type{
					{Kind: generator.TypeKindBase, Name: generator.BaseTypeBoolean},
					{
						Kind:    generator.TypeKindArray,
						Element: &generator.Type{Kind: generator.TypeKindReference, Name: "CreateFile"},
					},
				},
			}}},
		}
		var builder strings.Builder
		Expect(generator.UnionArmsTestFromMetaModel(&builder, model)).To(Succeed())
		Expect(builder.String()).To(And(
			ContainSubstring("package protocol_test"),
			ContainSubstring(withTags(`
				var _ = Describe("Decode", func() {
					DescribeTable("should decode every arm of a union as itself", expectArm,
						Entry("Change []CreateFile", unionValue(protocol.DecodeChange),
							'[{"kind":"create","severity":1,"uri":""}]', reflect.TypeFor[[]protocol.CreateFile]()),
						Entry("Change bool", unionValue(protocol.DecodeChange),
							'true', reflect.TypeFor[bool]()),
					)
				})
			`)),
		))
	})
})
