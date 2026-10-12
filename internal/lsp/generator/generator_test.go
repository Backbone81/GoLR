package generator_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/generator"
)

var _ = Describe("MetaModelToDirectory", func() {
	It("should write the generated files", func() {
		directory := GinkgoT().TempDir()
		Expect(generator.MetaModelToDirectory(directory, loadMetaModel())).To(Succeed())

		entries, err := os.ReadDir(directory)
		Expect(err).ToNot(HaveOccurred())
		var names []string
		for _, entry := range entries {
			names = append(names, entry.Name())
			content, err := os.ReadFile(filepath.Join(directory, entry.Name()))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(MatchRegexp(`^// Code generated .* DO NOT EDIT\.\n`))
		}
		Expect(names).To(ConsistOf("enums.gen.go", "methods.gen.go", "structures.gen.go", "unions.gen.go"))
	})
})
