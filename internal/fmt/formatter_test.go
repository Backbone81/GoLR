package fmt_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	golrfmt "github.com/backbone81/golr/internal/fmt"
	"github.com/backbone81/golr/internal/utils"
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
			input := utils.HereDoc(`
				@scanner{
				PLUS:"+";
				INTEGER:/[0-9]+/;
				}
			`)
			expected := utils.HereDoc(`
				@scanner {
				    PLUS:    "+";
				    INTEGER: /[0-9]+/;
				}
			`)
			Expect(golrfmt.GoLRString(input)).To(Equal(expected))
		})

		It("puts one parser alternative per line, led by ':' and '|', with ';' on its own line", func() {
			input := utils.HereDoc(`
				@parser{
				expression:term "+" term|term;
				}
			`)
			expected := utils.HereDoc(`
				@parser {
				    expression
				        : term "+" term
				        | term
				        ;
				}
			`)
			Expect(golrfmt.GoLRString(input)).To(Equal(expected))
		})

		It("keeps control directives single-line and tightens inline @precedence", func() {
			input := utils.HereDoc(`
				@parser{
				@start:Program;
				e:e "+" e @precedence ( PLUS );
				}
			`)
			expected := utils.HereDoc(`
				@parser {
				    @start: Program;

				    e
				        : e "+" e @precedence(PLUS)
				        ;
				}
			`)
			Expect(golrfmt.GoLRString(input)).To(Equal(expected))
		})

		It("collapses multiple blank lines between items to at most one", func() {
			input := utils.HereDoc(`
				@scanner {
				A: "a";



				B: "b";
				}
			`)
			expected := utils.HereDoc(`
				@scanner {
				    A: "a";

				    B: "b";
				}
			`)
			Expect(golrfmt.GoLRString(input)).To(Equal(expected))
		})

		It("is idempotent", func() {
			input := utils.HereDoc(`
				@parser{
				expression:term "+" term|term;
				term:INTEGER;
				}
			`)
			once, err := golrfmt.GoLRString(input)
			Expect(err).ToNot(HaveOccurred())
			twice, err := golrfmt.GoLRString(once)
			Expect(err).ToNot(HaveOccurred())
			Expect(twice).To(Equal(once))
		})

		It("tightens an inline @name(...) annotation the same way it tightens @precedence(...)", func() {
			input := utils.HereDoc(`
				@parser{
				file:scanner_section parser_section @name ( file );
				}
			`)
			expected := utils.HereDoc(`
				@parser {
				    file
				        : scanner_section parser_section @name(file)
				        ;
				}
			`)
			Expect(golrfmt.GoLRString(input)).To(Equal(expected))
		})

		It("does not invent a blank line inside an empty scanner block", func() {
			// The "}" arrives right after "{" with nothing in between, so it is already at the start of a
			// fresh, indented line; it must not schedule a linebreak of its own on top of that.
			expected := utils.HereDoc(`
				@scanner {
				}
			`)
			Expect(golrfmt.GoLRString("@scanner{}")).To(Equal(expected))
		})

		It("does not invent a blank line inside an empty parser block", func() {
			expected := utils.HereDoc(`
				@parser {
				}
			`)
			Expect(golrfmt.GoLRString("@parser{}")).To(Equal(expected))
		})
	})

	Context("comments", func() {
		It("preserves a leading top-level comment and the blank line after it", func() {
			input := utils.HereDoc(`
				// file header

				@scanner {
				    PLUS: "+";
				}
			`)
			Expect(golrfmt.GoLRString(input)).To(Equal(input))
		})

		It("preserves a line comment immediately before a scanner rule", func() {
			input := utils.HereDoc(`
				@scanner {
				    // marks the arithmetic operators
				    PLUS: "+";
				}
			`)
			Expect(golrfmt.GoLRString(input)).To(Equal(input))
		})

		It("preserves a block comment on its own line inside a parser section", func() {
			input := utils.HereDoc(`
				@parser {
				    /* entry point */
				    file
				        : @empty
				        ;
				}
			`)
			Expect(golrfmt.GoLRString(input)).To(Equal(input))
		})

		It("keeps a comment documenting one alternative on its own line, not run onto the previous one", func() {
			// This is the idiom the project's own bootstrap grammar (golr.golr) uses to document individual
			// error-recovery alternatives. A comment between two "|" alternatives must stay on its own line
			// rather than being swept into the token stream of the alternative before it.
			input := utils.HereDoc(`
				@parser {
				    expression
				        : a

				        // explains the next alternative
				        | b
				        | c
				        ;
				}
			`)
			Expect(golrfmt.GoLRString(input)).To(Equal(input))
		})

		It("preserves a trailing comment before a closing brace, along with its blank line", func() {
			input := utils.HereDoc(`
				@scanner {
				    PLUS: "+";

				    // trailing note
				}
			`)
			Expect(golrfmt.GoLRString(input)).To(Equal(input))
		})

		It("keeps a comment trailing a scanner rule on the same line, without an extra blank line after it", func() {
			// Regression test: a trailing comment used to leave a linebreak pending for the token after it
			// (here ";"), which the "}" then scheduled a second one on top of, producing a spurious blank
			// line. It also used to misuse the pending indent as the separator before "//", widening a single
			// space into a full indentation string.
			input := utils.HereDoc(`
				@scanner {
				    PLUS: "+"; // marks the plus operator
				}
			`)
			Expect(golrfmt.GoLRString(input)).To(Equal(input))
		})

		It("keeps a comment trailing the opening brace of a section on the same line", func() {
			input := utils.HereDoc(`
				@scanner { // arithmetic operators
				    PLUS: "+";
				}
			`)
			Expect(golrfmt.GoLRString(input)).To(Equal(input))
		})

		It("keeps a comment trailing one alternative on the same line, without an extra blank line before the ';'", func() {
			// Regression test: the same spurious-blank-line bug as above, but reached through onTokenSemi's
			// parser-rule branch instead of onTokenRbrace, and with a fresh indent level in between.
			input := utils.HereDoc(`
				@parser {
				    e
				        : a
				        | b // choose a or b
				        ;
				}
			`)
			Expect(golrfmt.GoLRString(input)).To(Equal(input))
		})
	})

	Context("malformed input", func() {
		It("does not fail on an unterminated block comment and keeps its bytes", func() {
			// The input deliberately ends without a trailing newline: it represents a source cut off mid-comment,
			// and closing the heredoc on its own line would put a newline into the string that was never there.
			input := utils.HereDoc(`
				@scanner {
				    PLUS: "+"; /* unterminated`)
			output, err := golrfmt.GoLRString(input)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(ContainSubstring("PLUS"))
			Expect(output).To(ContainSubstring("unterminated"))
		})

		It("keeps a rule name whose body is still being typed, without inventing the ';' or '}' it doesn't have yet", func() {
			// A formatter repositions whitespace around the tokens that exist; it must not invent a closing
			// '}' (or ':'/';') the source never had, even to keep the output superficially well-formed.
			// The input deliberately ends without a trailing newline: it represents a file still being typed, and
			// closing the heredoc on its own line would put a newline into the string that was never there.
			input := utils.HereDoc(`
				@parser {
				    file
				        : @empty
				        ;

				    partial`)
			expected := utils.HereDoc(`
				@parser {
				    file
				        : @empty
				        ;

				    partial
			`)
			Expect(golrfmt.GoLRString(input)).To(Equal(expected))
		})

		It("closes the block right after a rule missing its terminating ';'", func() {
			// The identifier's own emit leaves indentNext false (no linebreak was scheduled after it, since
			// the ';' that would normally do so is missing); the closing "}" must still start its own line
			// instead of trailing "PLUS" on the same one.
			expected := utils.HereDoc(`
				@scanner {
				    PLUS
				}
			`)
			Expect(golrfmt.GoLRString("@scanner{PLUS}")).To(Equal(expected))
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
