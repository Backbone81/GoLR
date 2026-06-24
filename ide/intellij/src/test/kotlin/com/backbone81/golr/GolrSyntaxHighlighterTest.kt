package com.backbone81.golr

import com.intellij.openapi.editor.colors.TextAttributesKey
import com.intellij.testFramework.fixtures.BasePlatformTestCase

// Verifies syntax highlighting: that the highlighter's lexer classifies the tokens of a real
// grammar snippet and that each token type is mapped to the expected color attribute key.
//
// Running through GolrSyntaxHighlighter.getHighlightingLexer() exercises the same lexer the
// editor uses, so this covers both token classification and the type -> color mapping in one go.
class GolrSyntaxHighlighterTest : BasePlatformTestCase() {

    private val highlighter = GolrSyntaxHighlighter()

    // Lexes `text` and returns, for every token, its text paired with the first highlight key
    // the highlighter assigns to it (or null when the token is not highlighted, e.g. whitespace).
    private fun highlightsOf(text: String): List<Pair<String, TextAttributesKey?>> {
        val lexer = highlighter.highlightingLexer
        val result = mutableListOf<Pair<String, TextAttributesKey?>>()
        lexer.start(text)
        while (lexer.tokenType != null) {
            val tokenText = text.substring(lexer.tokenStart, lexer.tokenEnd)
            val key = highlighter.getTokenHighlights(lexer.tokenType!!).firstOrNull()
            result.add(tokenText to key)
            lexer.advance()
        }
        return result
    }

    // Asserts that the (non-whitespace) token with the given text is highlighted with `key`.
    private fun assertHighlight(highlights: List<Pair<String, TextAttributesKey?>>, token: String, key: TextAttributesKey) {
        val actual = highlights.firstOrNull { it.first == token }?.second
        assertEquals("highlight for token '$token'", key, actual)
    }

    fun testScannerSectionTokensAreHighlighted() {
        val highlights = highlightsOf(
            """
            // line comment
            @scanner {
            INTEGER : /[0-9]+/ ;
            PLUS : "+" ;
            }
            """.trimIndent(),
        )

        assertHighlight(highlights, "// line comment", GolrSyntaxHighlighter.COMMENT)
        assertHighlight(highlights, "@scanner", GolrSyntaxHighlighter.KEYWORD_SECTION)
        assertHighlight(highlights, "{", GolrSyntaxHighlighter.BRACES)
        assertHighlight(highlights, "}", GolrSyntaxHighlighter.BRACES)
        assertHighlight(highlights, "INTEGER", GolrSyntaxHighlighter.IDENTIFIER)
        assertHighlight(highlights, ":", GolrSyntaxHighlighter.OPERATION_SIGN)
        assertHighlight(highlights, "/[0-9]+/", GolrSyntaxHighlighter.REGEX)
        assertHighlight(highlights, "\"+\"", GolrSyntaxHighlighter.STRING)
        assertHighlight(highlights, ";", GolrSyntaxHighlighter.SEMICOLON)
    }

    fun testParserSectionTokensAreHighlighted() {
        val highlights = highlightsOf(
            """
            @parser {
            @precedence {
            @left : PLUS ;
            }
            expression : expression PLUS expression @precedence(PLUS) | term ;
            }
            """.trimIndent(),
        )

        assertHighlight(highlights, "@parser", GolrSyntaxHighlighter.KEYWORD_SECTION)
        assertHighlight(highlights, "@precedence", GolrSyntaxHighlighter.KEYWORD_CONTROL)
        assertHighlight(highlights, "@left", GolrSyntaxHighlighter.KEYWORD_CONTROL)
        assertHighlight(highlights, "expression", GolrSyntaxHighlighter.IDENTIFIER)
        assertHighlight(highlights, "|", GolrSyntaxHighlighter.OPERATION_SIGN)
        assertHighlight(highlights, "(", GolrSyntaxHighlighter.PARENTHESES)
        assertHighlight(highlights, ")", GolrSyntaxHighlighter.PARENTHESES)
    }
}
