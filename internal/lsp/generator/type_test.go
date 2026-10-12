package generator_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/generator"
)

var _ = Describe("Type", func() {
	stringType := generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeString}

	DescribeTable("should decode every kind",
		func(input string, expected generator.Type) {
			var actual generator.Type
			Expect(json.Unmarshal([]byte(input), &actual)).To(Succeed())
			Expect(actual).To(Equal(expected))
		},
		Entry("base", `{"kind": "base", "name": "string"}`, stringType),
		Entry("reference", `{"kind": "reference", "name": "Range"}`,
			generator.Type{Kind: generator.TypeKindReference, Name: "Range"}),
		Entry("array", `{"kind": "array", "element": {"kind": "base", "name": "string"}}`,
			generator.Type{Kind: generator.TypeKindArray, Element: &stringType}),
		Entry("map",
			`{"kind": "map", "key": {"kind": "base", "name": "string"}, "value": {"kind": "base", "name": "string"}}`,
			generator.Type{Kind: generator.TypeKindMap, Key: &stringType, Value: &stringType}),
		Entry("and", `{"kind": "and", "items": [{"kind": "base", "name": "string"}]}`,
			generator.Type{Kind: generator.TypeKindAnd, Items: []generator.Type{stringType}}),
		Entry("or", `{"kind": "or", "items": [{"kind": "base", "name": "string"}]}`,
			generator.Type{Kind: generator.TypeKindOr, Items: []generator.Type{stringType}}),
		Entry("tuple", `{"kind": "tuple", "items": [{"kind": "base", "name": "string"}]}`,
			generator.Type{Kind: generator.TypeKindTuple, Items: []generator.Type{stringType}}),
		Entry("literal",
			`{"kind": "literal", "value": {"properties": [{"name": "a", "type": {"kind": "base", "name": "string"}}]}}`,
			generator.Type{Kind: generator.TypeKindLiteral, Literal: &generator.StructureLiteral{
				Properties: []generator.Property{{Name: "a", Type: stringType}},
			}}),
		Entry("string literal", `{"kind": "stringLiteral", "value": "create"}`,
			generator.Type{Kind: generator.TypeKindStringLiteral, StringValue: "create"}),
		Entry("integer literal", `{"kind": "integerLiteral", "value": -3}`,
			generator.Type{Kind: generator.TypeKindIntegerLiteral, IntegerValue: -3}),
		Entry("boolean literal", `{"kind": "booleanLiteral", "value": true}`,
			generator.Type{Kind: generator.TypeKindBooleanLiteral, BooleanValue: true}),
	)

	DescribeTable("should fail on invalid types",
		func(input string) {
			var actual generator.Type
			Expect(json.Unmarshal([]byte(input), &actual)).ToNot(Succeed())
		},
		Entry("unknown kind", `{"kind": "unknown"}`),
		Entry("unknown property", `{"kind": "base", "name": "string", "unknown": 1}`),
		Entry("map without value", `{"kind": "map", "key": {"kind": "base", "name": "string"}}`),
		Entry("string literal with a number", `{"kind": "stringLiteral", "value": 1}`),
		Entry("unknown property in a nested type", `{"kind": "array", "element": {"kind": "base", "unknown": 1}}`),
	)
})
