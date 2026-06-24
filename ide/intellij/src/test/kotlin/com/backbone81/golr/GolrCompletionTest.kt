package com.backbone81.golr

import com.intellij.testFramework.fixtures.BasePlatformTestCase

// Verifies that GolrCompletionContributor offers every symbol defined in the file as an
// auto-completion candidate when the caret sits on an identifier.
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
}
