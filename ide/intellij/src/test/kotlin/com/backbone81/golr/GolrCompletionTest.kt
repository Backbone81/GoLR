package com.backbone81.golr

import com.intellij.testFramework.fixtures.BasePlatformTestCase

// Verifies that GolrCompletionContributor offers the symbols defined in the file where a symbol
// may be referenced, and the keywords valid at the caret.
class GolrCompletionTest : BasePlatformTestCase() {

    // In a parser rule body, all nonterminals defined in the file should be suggested.
    fun testNonterminalsAreSuggestedInRuleBody() {
        myFixture.configureByText(
            "test.golr",
            """
            @parser {
            expression : term <caret> ;
            term : factor ;
            factor : expression ;
            }
            """.trimIndent(),
        )
        val suggestions = myFixture.completeBasic().map { it.lookupString }
        assertContainsElements(suggestions, "expression", "term", "factor")
    }

    // Terminals declared in @scanner must be offered in @parser rule bodies too, since a
    // production may reference either kind of symbol.
    fun testTerminalsAreSuggestedInRuleBody() {
        myFixture.configureByText(
            "test.golr",
            """
            @scanner {
            INTEGER : /[0-9]+/ ;
            PLUS : "+" ;
            }
            @parser {
            expression : INTEGER <caret> ;
            }
            """.trimIndent(),
        )
        val suggestions = myFixture.completeBasic().map { it.lookupString }
        assertContainsElements(suggestions, "INTEGER", "PLUS", "expression")
    }

    // A partially typed identifier should narrow the suggestions via the platform's prefix
    // matcher, and completing a unique prefix should insert the full name.
    fun testPrefixCompletesToFullName() {
        myFixture.configureByText(
            "test.golr",
            """
            @parser {
            expression : term ;
            term : expr<caret> ;
            }
            """.trimIndent(),
        )
        // Only "expression" matches the "expr" prefix, so completion inserts it directly.
        myFixture.completeBasic()
        myFixture.checkResult(
            """
            @parser {
            expression : term ;
            term : expression<caret> ;
            }
            """.trimIndent(),
        )
    }

    // Returns the lookup strings offered for `text`, which marks the caret with <caret>.
    private fun complete(text: String): List<String> {
        myFixture.configureByText("test.golr", text.trimIndent())
        return myFixture.completeBasic().map { it.lookupString }
    }

    fun testSectionKeywordsAreSuggestedAtTopLevel() {
        val suggestions = complete(
            """
            @scanner {
            A : "a" ;
            }
            <caret>
            """,
        )
        assertSameElements(suggestions, "@scanner", "@parser")
    }

    fun testScannerAnnotationsAreSuggestedInScannerRuleBody() {
        val suggestions = complete(
            """
            @scanner {
            A : "a" <caret>
            }
            """,
        )
        assertSameElements(suggestions, "@empty", "@skip", "@fragment")
    }

    // A new rule starts here, so no symbols are offered, only the statements of a @parser section.
    fun testParserStatementKeywordsAreSuggestedAtStatementStart() {
        val suggestions = complete(
            """
            @parser {
            a : a ;
            <caret>
            }
            """,
        )
        assertSameElements(suggestions, "@start", "@precedence")
    }

    fun testAssociativityKeywordsAreSuggestedInPrecedenceBlock() {
        val suggestions = complete(
            """
            @parser {
            @precedence {
            @left : a ;
            <caret>
            }
            a : a ;
            }
            """,
        )
        assertSameElements(suggestions, "@left", "@right", "@none", "@precedence")
    }

    fun testRuleBodyKeywordsAreSuggestedAfterAt() {
        val suggestions = complete(
            """
            @parser {
            a : a @<caret> ;
            }
            """,
        )
        assertSameElements(suggestions, "@empty", "@error", "@precedence", "@name")
    }

    // Inside @name(...) a new production name is written, so nothing is offered.
    fun testNothingIsSuggestedInsideName() {
        val suggestions = complete(
            """
            @parser {
            a : a @name(<caret>) ;
            }
            """,
        )
        assertEmpty(suggestions)
    }

    // A half-typed keyword completes to the full keyword including its "@".
    fun testKeywordPrefixCompletesToFullKeyword() {
        myFixture.configureByText(
            "test.golr",
            """
            @parser {
            a : a @na<caret> ;
            }
            """.trimIndent(),
        )
        myFixture.completeBasic()
        myFixture.checkResult(
            """
            @parser {
            a : a @name<caret> ;
            }
            """.trimIndent(),
        )
    }

    fun testNothingIsSuggestedInComments() {
        val suggestions = complete(
            """
            @parser {
            // <caret>
            a : a ;
            }
            """,
        )
        assertEmpty(suggestions)
    }
}
