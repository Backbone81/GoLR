package fmt_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	golrfmt "github.com/backbone81/golr/internal/fmt"
)

var _ = Describe("GoLR formatting", func() {
	Context("basic layout", func() {
		It("returns empty output for empty input", func() {
			Expect(golrfmt.GoLRString("")).To(Equal(""))
		})

		It("returns empty output for whitespace-only input", func() {
			Expect(golrfmt.GoLRString("   \n  \n")).To(Equal(""))
		})

		It("column-aligns a messy scanner section", func() {
			input := "@scanner{\nPLUS:\"+\";\nINTEGER:/[0-9]+/;\n}"
			expected := "@scanner {\n" +
				"    PLUS:    \"+\";\n" +
				"    INTEGER: /[0-9]+/;\n" +
				"}\n"
			Expect(golrfmt.GoLRString(input)).To(Equal(expected))
		})

		It("puts one parser alternative per line, led by ':' and '|', with ';' on its own line", func() {
			input := "@parser{\nexpression:term \"+\" term|term;\n}"
			expected := "@parser {\n" +
				"    expression\n" +
				"        : term \"+\" term\n" +
				"        | term\n" +
				"        ;\n" +
				"}\n"
			Expect(golrfmt.GoLRString(input)).To(Equal(expected))
		})

		It("keeps control directives single-line and tightens inline @precedence", func() {
			input := "@parser{\n@start:Program;\ne:e \"+\" e @precedence ( PLUS );\n}"
			expected := "@parser {\n" +
				"    @start: Program;\n" +
				"    e\n" +
				"        : e \"+\" e @precedence(PLUS)\n" +
				"        ;\n" +
				"}\n"
			Expect(golrfmt.GoLRString(input)).To(Equal(expected))
		})

		It("collapses multiple blank lines between items to at most one", func() {
			input := "@scanner {\nA: \"a\";\n\n\n\nB: \"b\";\n}"
			expected := "@scanner {\n" +
				"    A: \"a\";\n" +
				"\n" +
				"    B: \"b\";\n" +
				"}\n"
			Expect(golrfmt.GoLRString(input)).To(Equal(expected))
		})

		It("is idempotent", func() {
			input := "@parser{\nexpression:term \"+\" term|term;\nterm:INTEGER;\n}"
			once, err := golrfmt.GoLRString(input)
			Expect(err).ToNot(HaveOccurred())
			twice, err := golrfmt.GoLRString(once)
			Expect(err).ToNot(HaveOccurred())
			Expect(twice).To(Equal(once))
		})

		It("tightens an inline @name(...) annotation the same way it tightens @precedence(...)", func() {
			input := "@parser{\nfile:scanner_section parser_section @name ( file );\n}"
			expected := "@parser {\n" +
				"    file\n" +
				"        : scanner_section parser_section @name(file)\n" +
				"        ;\n" +
				"}\n"
			Expect(golrfmt.GoLRString(input)).To(Equal(expected))
		})
	})

	Context("comments", func() {
		It("preserves a leading top-level comment and the blank line after it", func() {
			input := "// file header\n\n@scanner {\n    PLUS: \"+\";\n}\n"
			Expect(golrfmt.GoLRString(input)).To(Equal(input))
		})

		It("preserves a line comment immediately before a scanner rule", func() {
			input := "@scanner {\n    // marks the arithmetic operators\n    PLUS: \"+\";\n}\n"
			Expect(golrfmt.GoLRString(input)).To(Equal(input))
		})

		It("preserves a block comment on its own line inside a parser section", func() {
			input := "@parser {\n    /* entry point */\n    file\n        : @empty\n        ;\n}\n"
			Expect(golrfmt.GoLRString(input)).To(Equal(input))
		})

		It("keeps a comment documenting one alternative on its own line, not run onto the previous one", func() {
			// This is the idiom the project's own bootstrap grammar (golr.golr) uses to document individual
			// error-recovery alternatives. A comment between two "|" alternatives must stay on its own line
			// rather than being swept into the token stream of the alternative before it.
			input := "@parser {\n" +
				"    expression\n" +
				"        : a\n" +
				"\n" +
				"        // explains the next alternative\n" +
				"        | b\n" +
				"\n" +
				"        | c\n" +
				"        ;\n" +
				"}\n"
			Expect(golrfmt.GoLRString(input)).To(Equal(input))
		})

		It("preserves a trailing comment before a closing brace, along with its blank line", func() {
			input := "@scanner {\n    PLUS: \"+\";\n\n    // trailing note\n}\n"
			Expect(golrfmt.GoLRString(input)).To(Equal(input))
		})
	})

	Context("malformed input", func() {
		It("does not fail on an unterminated block comment and keeps its bytes", func() {
			input := "@scanner {\n    PLUS: \"+\"; /* unterminated"
			output, err := golrfmt.GoLRString(input)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(ContainSubstring("PLUS"))
			Expect(output).To(ContainSubstring("unterminated"))
		})

		It("keeps a rule name whose body is still being typed, without inventing the ';' or '}' it doesn't have yet", func() {
			// A formatter repositions whitespace around the tokens that exist; it must not invent a closing
			// '}' (or ':'/';') the source never had, even to keep the output superficially well-formed.
			input := "@parser {\n    file\n        : @empty\n        ;\n\n    partial"
			expected := "@parser {\n" +
				"    file\n" +
				"        : @empty\n" +
				"        ;\n" +
				"\n" +
				"    partial\n"
			Expect(golrfmt.GoLRString(input)).To(Equal(expected))
		})
	})

	Context("real grammar files", func() {
		It("keeps every canonical .golr file in the repository unchanged", func() {
			// "ide" carries its own, deliberately non-canonical fixtures for the extension's own formatter
			// tests, so it is excluded rather than treated as a regression.
			const repoRoot = "../.."
			var paths []string
			Expect(filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() && d.Name() == "ide" {
					return filepath.SkipDir
				}
				if !d.IsDir() && strings.HasSuffix(path, ".golr") {
					paths = append(paths, path)
				}
				return nil
			})).To(Succeed())
			Expect(paths).ToNot(BeEmpty())

			for _, path := range paths {
				data, err := os.ReadFile(path)
				Expect(err).ToNot(HaveOccurred(), path)
				formatted, err := golrfmt.GoLRString(string(data))
				Expect(err).ToNot(HaveOccurred(), path)
				Expect(formatted).To(Equal(string(data)), path)
			}
		})
	})
})
