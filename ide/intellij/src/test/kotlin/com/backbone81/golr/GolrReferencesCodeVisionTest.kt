package com.backbone81.golr

import com.intellij.codeInsight.codeVision.CodeVisionAnchorKind
import com.intellij.codeInsight.codeVision.CodeVisionHost
import com.intellij.codeInsight.codeVision.settings.CodeVisionSettings
import com.intellij.testFramework.utils.codeVision.CodeVisionTestCase

// Verifies that GolrReferencesCodeVisionProvider renders the "N usages" Code Vision inlay
// above every symbol definition, driven through the real daemon (not by calling the provider
// directly).
class GolrReferencesCodeVisionTest : CodeVisionTestCase() {

    // testUsageCountsAreShownAboveDefinitions asserts "block" markers, which is how the Top anchor
    // renders. The anchor is a global setting the provider inherits, and the bundled Java plugin
    // defaults it to Right, so pin it here.
    override fun setUp() {
        super.setUp()
        CodeVisionSettings.getInstance().defaultPosition = CodeVisionAnchorKind.Top
    }

    override fun tearDown() {
        try {
            CodeVisionSettings.getInstance().resetDefaultPosition()
        } finally {
            super.tearDown()
        }
    }

    // CodeVisionTestCase strips the /*<# block ... #>*/ markers from the input, runs the code
    // vision pass for the enabled groups, re-inserts the markers from the actual inlays, and
    // asserts the result equals the input. So the markers below ARE the assertion. Each lens is
    // rendered as [<presentation>], hence the brackets around the counts.
    //
    // The vararg of testProviders is the set of enabled GROUP ids. GoLR has its own group
    // (see GolrReferencesCodeVisionProvider.GROUP_ID), so we enable "golr.usages".
    fun testUsageCountsAreShownAboveDefinitions() {
        // a is referenced once (in "b : a")     -> 1 usage
        // b is referenced twice (in "a : b b")  -> 2 usages
        testProviders(
            """
            @parser {
            /*<# block [1 usage] #>*/
            a : b b ;
            /*<# block [2 usages] #>*/
            b : a ;
            }
            """.trimIndent(),
            "test.golr",
            GolrReferencesCodeVisionProvider.GROUP_ID,
        )
    }

    // Guards that the inlay appears without enabling any group — i.e. GoLR's group is enabled by
    // default and does not depend on the shared platform "Usages" toggle. Uses golang.golr's
    // layout (name on its own line, ":" on the next, multi-line body).
    fun testUsageCountsAreShownWithDefaultSettings() {
        val src = """
            @parser {
                SourceFiles
                    : PackageClause ";"
                    | @empty
                    ;
                PackageClause
                    : "package" identifier
                    ;
            }
        """.trimIndent()
        myFixture.configureByText("test.golr", src)
        myFixture.doHighlighting()

        val host = project.getService(CodeVisionHost::class.java)
        project.putUserData(CodeVisionHost.isCodeVisionTestKey, true)
        host.calculateCodeVisionSync(myFixture.editor, testRootDisposable)

        // SourceFiles and PackageClause are the two definitions in the grammar above.
        val lenses = myFixture.editor.awaitCodeVisionLenses(src.length, expected = 2)
            .map { it.longPresentation }

        // SourceFiles is never referenced; PackageClause is referenced once.
        assertContainsElements(lenses, "no usages", "1 usage")
    }
}
