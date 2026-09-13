package com.backbone81.golr

import java.io.File
import junit.framework.TestCase

// Mechanical port of internal/fmt/formatter_test.go, kept as close as possible to that file's
// Describe/Context/It structure (each test method here corresponds 1:1 to one Go "It", named and
// grouped the same way) and to its fixtures (each utils.HereDoc(`...`) call becomes a
// hereDoc("""...""") call on the same content, closing quote aligned with the "val ... =" line
// the same way Go's closing backtick aligns with "... := ") so that a new Go test case has an
// obvious, same-shaped place to add here too. GolrFormatter no longer touches any IntelliJ
// platform types (Formatter drives the generated Scanner directly), so this no longer needs the
// heavy BasePlatformTestCase fixture the old item-model formatter's GolrLexer dependency required
// - a plain JUnit3-style TestCase is enough, and considerably faster.
class GolrFormatterTest : TestCase() {

    // --- Context("basic layout") ---

    // It("returns empty output for empty input")
    fun testBasicLayoutReturnsEmptyOutputForEmptyInput() {
        assertEquals("", GolrFormatter.format(""))
    }

    // It("returns empty output for whitespace-only input")
    fun testBasicLayoutReturnsEmptyOutputForWhitespaceOnlyInput() {
        assertEquals("", GolrFormatter.format("   \n  \n"))
    }

    // It("column-aligns a messy scanner section")
    fun testBasicLayoutColumnAlignsMessyScannerSection() {
        val input = hereDoc("""
            @scanner{
            PLUS:"+";
            INTEGER:/[0-9]+/;
            }
        """)
        val expected = hereDoc("""
            @scanner {
                PLUS:    "+";
                INTEGER: /[0-9]+/;
            }
        """)
        assertEquals(expected, GolrFormatter.format(input))
    }

    // It("puts one parser alternative per line, led by ':' and '|', with ';' on its own line")
    fun testBasicLayoutOneAlternativePerLine() {
        val input = hereDoc("""
            @parser{
            expression:term "+" term|term;
            }
        """)
        val expected = hereDoc("""
            @parser {
                expression
                    : term "+" term
                    | term
                    ;
            }
        """)
        assertEquals(expected, GolrFormatter.format(input))
    }

    // It("keeps control directives single-line and tightens inline @precedence")
    fun testBasicLayoutControlDirectivesSingleLineAndTightensPrecedence() {
        val input = hereDoc("""
            @parser{
            @start:Program;
            e:e "+" e @precedence ( PLUS );
            }
        """)
        val expected = hereDoc("""
            @parser {
                @start: Program;

                e
                    : e "+" e @precedence(PLUS)
                    ;
            }
        """)
        assertEquals(expected, GolrFormatter.format(input))
    }

    // It("collapses multiple blank lines between items to at most one")
    fun testBasicLayoutCollapsesMultipleBlankLinesToAtMostOne() {
        val input = hereDoc("""
            @scanner {
            A: "a";



            B: "b";
            }
        """)
        val expected = hereDoc("""
            @scanner {
                A: "a";

                B: "b";
            }
        """)
        assertEquals(expected, GolrFormatter.format(input))
    }

    // It("is idempotent")
    fun testBasicLayoutIsIdempotent() {
        val input = hereDoc("""
            @parser{
            expression:term "+" term|term;
            term:INTEGER;
            }
        """)
        val once = GolrFormatter.format(input)
        val twice = GolrFormatter.format(once)
        assertEquals(once, twice)
    }

    // It("tightens an inline @name(...) annotation the same way it tightens @precedence(...)")
    fun testBasicLayoutTightensNameAnnotationLikePrecedence() {
        val input = hereDoc("""
            @parser{
            file:scanner_section parser_section @name ( file );
            }
        """)
        val expected = hereDoc("""
            @parser {
                file
                    : scanner_section parser_section @name(file)
                    ;
            }
        """)
        assertEquals(expected, GolrFormatter.format(input))
    }

    // It("does not invent a blank line inside an empty scanner block")
    fun testBasicLayoutDoesNotInventBlankLineInsideEmptyScannerBlock() {
        // The "}" arrives right after "{" with nothing in between, so it is already at the start
        // of a fresh, indented line; it must not schedule a linebreak of its own on top of that.
        val expected = hereDoc("""
            @scanner {
            }
        """)
        assertEquals(expected, GolrFormatter.format("@scanner{}"))
    }

    // It("does not invent a blank line inside an empty parser block")
    fun testBasicLayoutDoesNotInventBlankLineInsideEmptyParserBlock() {
        val expected = hereDoc("""
            @parser {
            }
        """)
        assertEquals(expected, GolrFormatter.format("@parser{}"))
    }

    // --- Context("comments") ---

    // It("preserves a leading top-level comment and the blank line after it")
    fun testCommentsPreservesLeadingTopLevelCommentAndBlankLineAfterIt() {
        val input = hereDoc("""
            // file header

            @scanner {
                PLUS: "+";
            }
        """)
        assertEquals(input, GolrFormatter.format(input))
    }

    // It("preserves a line comment immediately before a scanner rule")
    fun testCommentsPreservesLineCommentImmediatelyBeforeScannerRule() {
        val input = hereDoc("""
            @scanner {
                // marks the arithmetic operators
                PLUS: "+";
            }
        """)
        assertEquals(input, GolrFormatter.format(input))
    }

    // It("preserves a block comment on its own line inside a parser section")
    fun testCommentsPreservesBlockCommentOnOwnLineInsideParserSection() {
        val input = hereDoc("""
            @parser {
                /* entry point */
                file
                    : @empty
                    ;
            }
        """)
        assertEquals(input, GolrFormatter.format(input))
    }

    // It("keeps a comment documenting one alternative on its own line, not run onto the previous one")
    fun testCommentsKeepsCommentDocumentingAlternativeOnOwnLine() {
        val input = hereDoc("""
            @parser {
                expression
                    : a

                    // explains the next alternative
                    | b
                    | c
                    ;
            }
        """)
        assertEquals(input, GolrFormatter.format(input))
    }

    // It("preserves a trailing comment before a closing brace, along with its blank line")
    fun testCommentsPreservesTrailingCommentBeforeClosingBrace() {
        val input = hereDoc("""
            @scanner {
                PLUS: "+";

                // trailing note
            }
        """)
        assertEquals(input, GolrFormatter.format(input))
    }

    // It("keeps a comment trailing a scanner rule on the same line, without an extra blank line after it")
    fun testCommentsKeepsCommentTrailingScannerRuleOnSameLine() {
        val input = hereDoc("""
            @scanner {
                PLUS: "+"; // marks the plus operator
            }
        """)
        assertEquals(input, GolrFormatter.format(input))
    }

    // It("keeps a comment trailing the opening brace of a section on the same line")
    fun testCommentsKeepsCommentTrailingOpeningBraceOnSameLine() {
        val input = hereDoc("""
            @scanner { // arithmetic operators
                PLUS: "+";
            }
        """)
        assertEquals(input, GolrFormatter.format(input))
    }

    // It("keeps a comment trailing one alternative on the same line, without an extra blank line before the ';'")
    fun testCommentsKeepsCommentTrailingAlternativeOnSameLine() {
        val input = hereDoc("""
            @parser {
                e
                    : a
                    | b // choose a or b
                    ;
            }
        """)
        assertEquals(input, GolrFormatter.format(input))
    }

    // --- Context("malformed input") ---

    // It("does not fail on an unterminated block comment and keeps its bytes")
    fun testMalformedInputUnterminatedBlockCommentKeepsBytes() {
        // The input deliberately ends without a trailing newline: it represents a source cut off
        // mid-comment, and closing the raw string on its own line would put a newline into the
        // string that was never there.
        val input = hereDoc("""
            @scanner {
                PLUS: "+"; /* unterminated""")
        val output = GolrFormatter.format(input)
        assertTrue(output.contains("PLUS"))
        assertTrue(output.contains("unterminated"))
    }

    // It("keeps a rule name whose body is still being typed, without inventing the ';' or '}' it doesn't have yet")
    fun testMalformedInputKeepsPartialRuleNameWithoutInventingStructure() {
        // The input deliberately ends without a trailing newline: it represents a file still
        // being typed, and closing the raw string on its own line would put a newline into the
        // string that was never there.
        val input = hereDoc("""
            @parser {
                file
                    : @empty
                    ;

                partial""")
        val expected = hereDoc("""
            @parser {
                file
                    : @empty
                    ;

                partial
        """)
        assertEquals(expected, GolrFormatter.format(input))
    }

    // It("closes the block right after a rule missing its terminating ';'")
    fun testMalformedInputClosesBlockRightAfterRuleMissingSemicolon() {
        // The identifier's own emit leaves indentNext false (no linebreak was scheduled after
        // it, since the ';' that would normally do so is missing); the closing "}" must still
        // start its own line instead of trailing "PLUS" on the same one.
        val expected = hereDoc("""
            @scanner {
                PLUS
            }
        """)
        assertEquals(expected, GolrFormatter.format("@scanner{PLUS}"))
    }

    // --- Context("real grammar files") ---

    // It("keeps every canonical .golr file in the repository unchanged")
    //
    // Go's version relies on "go test"'s cwd always being the package directory ("../.." from
    // internal/fmt is the repo root, deterministically). Gradle/IntelliJ test runners don't
    // guarantee a cwd relationship to the repo root the same way, so this walks upward from the
    // cwd looking for go.mod instead - a mechanical adaptation of "repoRoot", not a behavioral one.
    fun testRealGrammarFilesKeepsEveryCanonicalGolrFileUnchanged() {
        val repoRoot = findRepoRoot() ?: return

        val paths = mutableListOf<File>()
        repoRoot.walkTopDown()
            .onEnter { it.name != "ide" }
            .filterTo(paths) { it.isFile && it.extension == "golr" }
        assertTrue("expected to find at least one .golr file under $repoRoot", paths.isNotEmpty())

        for (path in paths) {
            val data = path.readText()
            assertEquals(path.toString(), data, GolrFormatter.format(data))
        }
    }

    // Walks upward from the current working directory looking for go.mod, the repo root's marker
    // file. Returns null (skipping the test) rather than failing outright if it can't be found,
    // since that would be an environment problem rather than a formatting one.
    private fun findRepoRoot(): File? {
        var dir: File? = File(".").absoluteFile
        while (dir != null) {
            if (File(dir, "go.mod").isFile) {
                return dir
            }
            dir = dir.parentFile
        }
        return null
    }
}
