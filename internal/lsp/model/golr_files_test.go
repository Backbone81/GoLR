package model_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// golrFilePatterns match the grammar files of the repository the model is checked against the golr frontend on.
var golrFilePatterns = []string{
	"../../../examples/*/*.golr",
	"../../../examples/*/spec/*.golr",
	"../../backendtest/golden/*/spec.golr",
	"../../parsergen/frontend/golr/spec/golr.golr",
}

// golrFiles returns the files the patterns match.
func golrFiles(patterns []string) []string {
	var files []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			panic(err)
		}
		files = append(files, matches...)
	}
	return files
}

var _ = Describe("golr files", func() {
	It("finds the grammar files", func() {
		Expect(len(golrFiles(golrFilePatterns))).To(BeNumerically(">", 40))
	})

	for _, filePath := range golrFiles(golrFilePatterns) {
		It("matches the golr frontend on "+filePath, func() {
			text, err := os.ReadFile(filePath)
			Expect(err).ToNot(HaveOccurred())
			expectModelMatchesFrontend(text, filePath)
		})
	}
})
