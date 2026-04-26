package spec_test

import (
	"golr/examples/golang/spec"
	"golr/pkg/scannergen/backend/json"
	"golr/pkg/scannergen/backend/yaml"
	"golr/pkg/scannergen/core/subset"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Golang Parser", func() {
	It("should serialize and deserialize the scanner to JSON", func() {
		rules := spec.GetScannerRules()
		dfa := subset.RulesToDFA(rules)
		jsonString, err := json.DFAToString(dfa)
		Expect(err).ToNot(HaveOccurred())

		gotDFA, err := json.DFAFromString(jsonString)
		Expect(err).ToNot(HaveOccurred())

		Expect(gotDFA).To(Equal(dfa))
	})

	It("should serialize and deserialize the scanner to YAML", func() {
		rules := spec.GetScannerRules()
		dfa := subset.RulesToDFA(rules)
		yamlString, err := yaml.DFAToString(dfa)
		Expect(err).ToNot(HaveOccurred())

		gotDFA, err := yaml.DFAFromString(yamlString)
		Expect(err).ToNot(HaveOccurred())

		Expect(gotDFA).To(Equal(dfa))
	})
})

func BenchmarkGolangScannerSubset(b *testing.B) {
	rules := spec.GetScannerRules()
	for b.Loop() {
		_ = subset.RulesToDFA(rules)
	}
}

func BenchmarkGolangScanner(b *testing.B) {
	rules := spec.GetScannerRules()
	dfa := subset.RulesToDFA(rules)
	b.Run("JSON", func(b *testing.B) {
		b.Run("To", func(b *testing.B) {
			for b.Loop() {
				if _, err := json.DFAToString(dfa); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run("From", func(b *testing.B) {
			jsonString, err := json.DFAToString(dfa)
			if err != nil {
				b.Fatal(err)
			}
			for b.Loop() {
				if _, err := json.DFAFromString(jsonString); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
	b.Run("YAML", func(b *testing.B) {
		b.Run("To", func(b *testing.B) {
			for b.Loop() {
				if _, err := yaml.DFAToString(dfa); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run("From", func(b *testing.B) {
			yamlString, err := yaml.DFAToString(dfa)
			if err != nil {
				b.Fatal(err)
			}
			for b.Loop() {
				if _, err := yaml.DFAFromString(yamlString); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}
