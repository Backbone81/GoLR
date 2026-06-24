package com.backbone81.golr

import com.intellij.testFramework.fixtures.BasePlatformTestCase
import java.io.File

// Unit tests for the pure formatting logic in GolrFormatter. Runs inside a platform fixture
// because GolrFormatter reuses GolrLexer, whose IElementType tokens require the platform.
class GolrFormatterTest : BasePlatformTestCase() {

    private fun assertFormatted(input: String, expected: String) {
        assertEquals(expected, GolrFormatter.format(input.trimIndent()))
    }

    // Scanner bodies are column-aligned to one space past the longest "name:" in the group.
    fun testScannerBodiesAreColumnAligned() {
        assertFormatted(
            """
            @scanner {
            horizontal_whitespace: /[ \t]/ @fragment;
            vertical_whitespace: /[\r\n]/ @fragment;
            whitespace: /x/ @skip;
            }
            """,
            "@scanner {\n" +
                "    horizontal_whitespace: /[ \\t]/ @fragment;\n" +
                "    vertical_whitespace:   /[\\r\\n]/ @fragment;\n" +
                "    whitespace:            /x/ @skip;\n" +
                "}\n",
        )
    }

    // A blank line starts a new alignment group, so the two groups align independently.
    fun testBlankLineSeparatesAlignmentGroups() {
        assertFormatted(
            """
            @scanner {
            add: "+";
            sub: "-";

            shift_left: "<<";
            and: "&";
            }
            """,
            "@scanner {\n" +
                "    add: \"+\";\n" +
                "    sub: \"-\";\n" +
                "\n" +
                "    shift_left: \"<<\";\n" +
                "    and:        \"&\";\n" +
                "}\n",
        )
    }

    // Messy whitespace and a single-line rule are reflowed: name on its own line, one
    // alternative per line, ";" on its own line, all indented with 4 spaces.
    fun testParserRuleIsExpandedToMultiline() {
        assertFormatted(
            """
            @parser {
            Operand : Literal | OperandName | "(" Expression ")" ;
            }
            """,
            "@parser {\n" +
                "    Operand\n" +
                "        : Literal\n" +
                "        | OperandName\n" +
                "        | \"(\" Expression \")\"\n" +
                "        ;\n" +
                "}\n",
        )
    }

    // @start and @precedence directives stay single-line; the @precedence block nests one level.
    fun testControlDirectivesAndPrecedenceBlock() {
        assertFormatted(
            """
            @parser {
            @start : SourceFiles ;
            @precedence {
            @left : "*" "/" ;
            @left : "+" "-" ;
            }
            }
            """,
            "@parser {\n" +
                "    @start: SourceFiles;\n" +
                "    @precedence {\n" +
                "        @left: \"*\" \"/\";\n" +
                "        @left: \"+\" \"-\";\n" +
                "    }\n" +
                "}\n",
        )
    }

    // Comments are preserved, re-indented to their context, and keep an attached rule together
    // while a preceding blank line is kept (collapsed to a single blank).
    fun testCommentsArePreservedAndReindented() {
        assertFormatted(
            """
            @parser {



            // leading comment
            Foo : bar ;
            }
            """,
            "@parser {\n" +
                "    // leading comment\n" +
                "    Foo\n" +
                "        : bar\n" +
                "        ;\n" +
                "}\n",
        )
    }

    // Formatting is idempotent: formatting already-formatted text changes nothing.
    fun testFormattingIsIdempotent() {
        val golang = golangGrammarOrNull() ?: return
        val once = GolrFormatter.format(golang)
        val twice = GolrFormatter.format(once)
        assertEquals(once, twice)
    }

    // The reference grammar is already in canonical form, so formatting it is a no-op. This is
    // the strongest guarantee that the formatter reproduces the intended layout exactly.
    fun testReferenceGrammarIsUnchanged() {
        val golang = golangGrammarOrNull() ?: return
        assertEquals(golang, GolrFormatter.format(golang))
    }

    // Loads examples/golang/spec/golang.golr, or returns null (skipping) if the submodule with
    // the example grammars has not been checked out.
    private fun golangGrammarOrNull(): String? =
        listOf(
            File("../../examples/golang/spec/golang.golr"),
            File("examples/golang/spec/golang.golr"),
        ).firstOrNull { it.exists() }?.readText()
}
