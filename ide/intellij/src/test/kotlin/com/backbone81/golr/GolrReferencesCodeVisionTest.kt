package com.backbone81.golr

import com.intellij.codeInsight.codeVision.CodeVisionHost
import com.intellij.codeInsight.codeVision.ui.model.CodeVisionListData
import com.intellij.codeInsight.codeVision.ui.renderers.CodeVisionInlayRenderer
import com.intellij.testFramework.utils.codeVision.CodeVisionTestCase

// Verifies that GolrReferencesCodeVisionProvider renders the "N usages" Code Vision inlay
// above every symbol definition, driven through the real daemon (not by calling the provider
// directly).
class GolrReferencesCodeVisionTest : CodeVisionTestCase() {

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

    // Guards that the inlay appears with NO settings changes — i.e. GoLR's group is enabled by
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

        val lenses = myFixture.editor.inlayModel.getBlockElementsInRange(0, src.length)
            .filter { it.renderer is CodeVisionInlayRenderer }
            .mapNotNull { it.getUserData(CodeVisionListData.KEY) }
            .flatMap { it.visibleLens }
            .map { it.longPresentation }

        // SourceFiles is never referenced; PackageClause is referenced once.
        assertContainsElements(lenses, "no usages", "1 usage")
    }
}
