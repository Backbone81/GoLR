package spec_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/examples/golang/spec"
	"github.com/backbone81/golr/pkg/scannergen/frontend/json"
	"github.com/backbone81/golr/pkg/scannergen/frontend/yaml"
)

var _ = Describe("Golang Spec", func() {
	It("should serialize and deserialize the scanner rules to JSON", func() {
		rules := spec.GetScannerRules()
		jsonString, err := json.RulesToString(rules)
		Expect(err).ToNot(HaveOccurred())

		gotRules, err := json.RulesFromString(jsonString)
		Expect(err).ToNot(HaveOccurred())

		Expect(gotRules).To(Equal(rules))
	})

	It("should serialize and deserialize the scanner rules to YAML", func() {
		rules := spec.GetScannerRules()
		yamlString, err := yaml.RulesToString(rules)
		Expect(err).ToNot(HaveOccurred())

		gotRules, err := yaml.RulesFromString(yamlString)
		Expect(err).ToNot(HaveOccurred())

		Expect(gotRules).To(Equal(rules))
	})
})

func BenchmarkGolangScannerRules(b *testing.B) {
	rules := spec.GetScannerRules()
	b.Run("JSON", func(b *testing.B) {
		b.Run("To", func(b *testing.B) {
			for b.Loop() {
				if _, err := json.RulesToString(rules); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run("From", func(b *testing.B) {
			jsonString, err := json.RulesToString(rules)
			if err != nil {
				b.Fatal(err)
			}
			for b.Loop() {
				if _, err := json.RulesFromString(jsonString); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
	b.Run("YAML", func(b *testing.B) {
		b.Run("To", func(b *testing.B) {
			for b.Loop() {
				if _, err := yaml.RulesToString(rules); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run("From", func(b *testing.B) {
			yamlString, err := yaml.RulesToString(rules)
			if err != nil {
				b.Fatal(err)
			}
			for b.Loop() {
				if _, err := yaml.RulesFromString(yamlString); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}
