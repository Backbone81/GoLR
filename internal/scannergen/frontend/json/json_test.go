package json_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/scannergen/frontend/json"
)

var _ = Describe("JSON", func() {
	It("returns an error for a node without the value of its kind", func() {
		_, err := json.RulesFromString(`[{"name":"A","regex":{"kind":"Literal"}}]`)
		Expect(err).To(HaveOccurred())
	})

	It("returns an error for a rule with an invalid regular expression", func() {
		_, err := json.RulesFromString(`[{"name":"A","regex":{"kind":"Literal","literal":{"text":""}}}]`)
		Expect(err).To(HaveOccurred())
	})
})
