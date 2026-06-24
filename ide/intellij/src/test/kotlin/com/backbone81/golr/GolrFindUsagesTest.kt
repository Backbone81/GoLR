package com.backbone81.golr

import com.intellij.psi.util.PsiTreeUtil
import com.intellij.testFramework.fixtures.BasePlatformTestCase

// Verifies "Find Usages" (Alt+F7): every reference to a symbol definition is reported. This
// drives the full stack — myFixture.findUsages() goes through ReferencesSearch, which is routed
// to GolrReferencesSearcher and confirmed via GolrSymbolReference resolution.
class GolrFindUsagesTest : BasePlatformTestCase() {

    // Returns the GolrSymbolDefinition named `name` in the currently configured file.
    private fun definition(name: String): GolrSymbolDefinition =
        PsiTreeUtil.findChildrenOfType(myFixture.file, GolrSymbolDefinition::class.java)
            .first { it.name == name }

    fun testFindUsagesReportsAllReferences() {
        myFixture.configureByText(
            "test.golr",
            """
            @parser {
            a : b b ;
            b : a ;
            }
            """.trimIndent(),
        )
        // b is referenced twice (in "a : b b"); a is referenced once (in "b : a").
        assertEquals(2, myFixture.findUsages(definition("b")).size)
        assertEquals(1, myFixture.findUsages(definition("a")).size)
    }

    // A terminal defined in @scanner has its usages found in @parser rule bodies.
    fun testFindUsagesAcrossSections() {
        myFixture.configureByText(
            "test.golr",
            """
            @scanner {
            PLUS : "+" ;
            }
            @parser {
            expression : expression PLUS expression ;
            }
            """.trimIndent(),
        )
        assertEquals(1, myFixture.findUsages(definition("PLUS")).size)
        // expression references itself twice in its own body.
        assertEquals(2, myFixture.findUsages(definition("expression")).size)
    }

    // A definition that is never referenced reports zero usages.
    fun testUnusedSymbolHasNoUsages() {
        myFixture.configureByText(
            "test.golr",
            """
            @parser {
            a : a ;
            unused : a ;
            }
            """.trimIndent(),
        )
        assertEquals(0, myFixture.findUsages(definition("unused")).size)
    }

    // Find Usages is offered on definitions but not on references (a reference defers to its
    // definition first). Guards GolrFindUsagesProvider.canFindUsagesFor().
    fun testCanFindUsagesOnlyForDefinitions() {
        myFixture.configureByText(
            "test.golr",
            """
            @parser {
            a : b ;
            b : a ;
            }
            """.trimIndent(),
        )
        val provider = GolrFindUsagesProvider()
        val def = definition("a")
        val ref = PsiTreeUtil.findChildrenOfType(myFixture.file, GolrSymbolReference::class.java).first()

        assertTrue("Find Usages available on a definition", provider.canFindUsagesFor(def))
        assertFalse("Find Usages not available on a reference", provider.canFindUsagesFor(ref))
    }
}
