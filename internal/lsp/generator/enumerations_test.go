package generator_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/generator"
	"github.com/backbone81/golr/internal/utils"
)

// enumeration returns an enumeration over the base type with a value per name and JSON value.
func enumeration(name string, baseType string, values ...string) generator.Enumeration {
	result := generator.Enumeration{
		Name: name,
		Type: generator.Type{Kind: generator.TypeKindBase, Name: baseType},
	}
	for i := 0; i < len(values); i += 2 {
		result.Values = append(result.Values, generator.EnumerationEntry{
			Name:  values[i],
			Value: []byte(values[i+1]),
		})
	}
	return result
}

var _ = Describe("EnumerationsFromMetaModel", func() {
	DescribeTable("should define the type and its constants",
		func(input generator.Enumeration, expected string) {
			model := &generator.MetaModel{Enumerations: []generator.Enumeration{input}}
			var builder strings.Builder
			Expect(generator.EnumerationsFromMetaModel(&builder, model)).To(Succeed())
			Expect(builder.String()).To(ContainSubstring(expected))
		},
		Entry("string", enumeration("Kind", generator.BaseTypeString, "namespace", `"namespace"`), utils.HereDoc(`
			type Kind string

			const (
				KindNamespace Kind = "namespace"
			)
		`)),
		Entry("integer", enumeration("Code", generator.BaseTypeInteger, "ParseError", `-32700`), utils.HereDoc(`
			type Code int

			const (
				CodeParseError Code = -32700
			)
		`)),
		Entry("uinteger", enumeration("Severity", generator.BaseTypeUinteger, "Error", `1`), utils.HereDoc(`
			type Severity int

			const (
				SeverityError Severity = 1
			)
		`)),
	)

	It("should write an unsupported type with its name", func() {
		model := &generator.MetaModel{Enumerations: []generator.Enumeration{
			enumeration("Color", generator.BaseTypeDecimal, "Half", `0.5`),
		}}
		var builder strings.Builder
		Expect(generator.EnumerationsFromMetaModel(&builder, model)).To(Succeed())
		Expect(builder.String()).To(ContainSubstring(utils.HereDoc(`
			type Color decimal

			const (
				ColorHalf Color = 0.5
			)
		`)))
	})

	It("should link the section of the specification", func() {
		model := &generator.MetaModel{
			MetaData:     generator.MetaData{Version: "3.18.0"},
			Enumerations: []generator.Enumeration{enumeration("SymbolKind", generator.BaseTypeUinteger, "File", `1`)},
		}
		var builder strings.Builder
		Expect(generator.EnumerationsFromMetaModel(&builder, model)).To(Succeed())
		Expect(builder.String()).To(ContainSubstring(utils.HereDoc(`
			// See https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification/#symbolKind.
			type SymbolKind int
		`)))
	})
})
