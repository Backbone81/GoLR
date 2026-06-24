package com.backbone81.golr

import com.intellij.testFramework.fixtures.BasePlatformTestCase

// Verifies "Go to Definition": resolving the reference under the caret lands on the matching
// GolrSymbolDefinition. This exercises GolrSymbolReference.GolrRef.multiResolve(), the same
// resolution path Ctrl+B / Cmd+B uses in the editor.
class GolrGoToDefinitionTest : BasePlatformTestCase() {

    // A reference in a parser rule body resolves to the nonterminal definition.
    fun testReferenceInRuleBodyResolvesToDefinition() {
        myFixture.configureByText(
            "test.golr",
            """
            @parser {
            a : <caret>b ;
            b : a ;
            }
            """.trimIndent(),
        )
        val reference = myFixture.file.findReferenceAt(myFixture.caretOffset)
        assertNotNull("expected a reference under the caret", reference)

        val target = reference!!.resolve()
        assertInstanceOf(target, GolrSymbolDefinition::class.java)
        assertEquals("b", (target as GolrSymbolDefinition).name)
        assertFalse("b is a nonterminal", target.isTerminal())
    }

    // A terminal referenced from the @parser section resolves to its @scanner definition.
    fun testReferenceResolvesToTerminalDefinition() {
        myFixture.configureByText(
            "test.golr",
            """
            @scanner {
            PLUS : "+" ;
            }
            @parser {
            expression : <caret>PLUS ;
            }
            """.trimIndent(),
        )
        val target = myFixture.file.findReferenceAt(myFixture.caretOffset)?.resolve()
        assertInstanceOf(target, GolrSymbolDefinition::class.java)
        assertEquals("PLUS", (target as GolrSymbolDefinition).name)
        assertTrue("PLUS is a terminal", target.isTerminal())
    }

    // The @start declaration's symbol is a reference that resolves to its definition.
    fun testStartDeclarationResolvesToDefinition() {
        myFixture.configureByText(
            "test.golr",
            """
            @parser {
            @start : <caret>expression ;
            expression : expression ;
            }
            """.trimIndent(),
        )
        val target = myFixture.file.findReferenceAt(myFixture.caretOffset)?.resolve()
        assertInstanceOf(target, GolrSymbolDefinition::class.java)
        assertEquals("expression", (target as GolrSymbolDefinition).name)
    }

    // An identifier that is not defined anywhere resolves to nothing (no jump target).
    fun testUnresolvedReferenceResolvesToNull() {
        myFixture.configureByText(
            "test.golr",
            """
            @parser {
            a : <caret>undefined ;
            }
            """.trimIndent(),
        )
        val reference = myFixture.file.findReferenceAt(myFixture.caretOffset)
        assertNotNull("expected a reference under the caret", reference)
        assertNull("undefined symbol should not resolve", reference!!.resolve())
    }
}
