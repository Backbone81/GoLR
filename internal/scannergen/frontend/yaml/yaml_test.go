package yaml_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/scannergen/frontend/yaml"
)

var _ = Describe("YAML", func() {
	It("returns an error for a node without the value of its kind", func() {
		_, err := yaml.RulesFromString("- name: A\n  regex:\n    kind: Literal\n")
		Expect(err).To(HaveOccurred())
	})

	It("returns an error for a rule with an invalid regular expression", func() {
		_, err := yaml.RulesFromString("- name: A\n  regex:\n    kind: Literal\n    literal:\n      text: \"\"\n")
		Expect(err).To(HaveOccurred())
	})
})
