package generator_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/generator"
)

// metaModelPath is the meta model the protocol package is generated from.
const metaModelPath = "../spec/metaModel.json"

// loadMetaModel returns the meta model the protocol package is generated from.
func loadMetaModel() *generator.MetaModel {
	data, err := os.ReadFile(metaModelPath)
	Expect(err).ToNot(HaveOccurred())

	model, err := generator.ParseMetaModel(data)
	Expect(err).ToNot(HaveOccurred())
	return model
}

var _ = Describe("ParseMetaModel", func() {
	It("should parse the meta model of the specification", func() {
		model := loadMetaModel()
		Expect(model.MetaData.Version).ToNot(BeEmpty())
		Expect(model.Requests).ToNot(BeEmpty())
		Expect(model.Notifications).ToNot(BeEmpty())
		Expect(model.Structures).ToNot(BeEmpty())
		Expect(model.Enumerations).ToNot(BeEmpty())
		Expect(model.TypeAliases).ToNot(BeEmpty())
	})

	It("should parse the properties of a structure", func() {
		model, err := generator.ParseMetaModel([]byte(`{
			"metaData": {"version": "1.0.0"},
			"structures": [{
				"name": "Position",
				"documentation": "A position.",
				"since": "3.0.0",
				"properties": [{"name": "line", "type": {"kind": "base", "name": "uinteger"}, "optional": true}]
			}]
		}`))
		Expect(err).ToNot(HaveOccurred())
		Expect(model.Structures).To(Equal([]generator.Structure{{
			Annotations: generator.Annotations{Documentation: "A position.", Since: "3.0.0"},
			Name:        "Position",
			Properties: []generator.Property{{
				Name:     "line",
				Type:     generator.Type{Kind: generator.TypeKindBase, Name: generator.BaseTypeUinteger},
				Optional: true,
			}},
		}}))
	})

	It("should fail on an unknown property", func() {
		_, err := generator.ParseMetaModel([]byte(`{"metaData": {"version": "1.0.0", "unknown": true}}`))
		Expect(err).To(MatchError(ContainSubstring(`unknown field "unknown"`)))
	})

	It("should fail on params given as a list of types", func() {
		_, err := generator.ParseMetaModel([]byte(`{"notifications": [{
			"method": "a", "messageDirection": "both", "params": [{"kind": "base", "name": "string"}]
		}]}`))
		Expect(err).To(HaveOccurred())
	})
})
