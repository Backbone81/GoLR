package com.backbone81.golr

import com.intellij.codeInsight.codeVision.CodeVisionHost
import com.intellij.codeInsight.codeVision.ui.model.CodeVisionListData
import com.intellij.codeInsight.codeVision.ui.renderers.CodeVisionInlayRenderer
import com.intellij.psi.util.PsiTreeUtil
import com.intellij.testFramework.utils.codeVision.CodeVisionTestCase
import java.io.File

// Exercises the code vision pass against the real, large golang.golr grammar to prove the
// "N usages" inlay is produced for every rule definition in a realistic file (not just the
// small synthetic grammars in GolrReferencesCodeVisionTest).
//
// The file lives in the examples/ submodule; if it has not been checked out the test skips
// rather than failing, so it stays CI-safe.
class GolrGolangGrammarCodeVisionTest : CodeVisionTestCase() {

    fun testEveryDefinitionInGolangGrammarGetsAUsageLens() {
        val file = listOf(
            File("../../examples/golang/spec/golang.golr"),
            File("examples/golang/spec/golang.golr"),
        ).firstOrNull { it.exists() } ?: return // submodule not checked out -> skip

        val src = file.readText()
        myFixture.configureByText("golang.golr", src)
        myFixture.doHighlighting()

        val host = project.getService(CodeVisionHost::class.java)
        project.putUserData(CodeVisionHost.isCodeVisionTestKey, true)
        host.calculateCodeVisionSync(myFixture.editor, testRootDisposable)

        val definitionCount = PsiTreeUtil.findChildrenOfType(myFixture.file, GolrSymbolDefinition::class.java).size
        val lensCount = myFixture.editor.inlayModel.getBlockElementsInRange(0, src.length)
            .filter { it.renderer is CodeVisionInlayRenderer }
            .mapNotNull { it.getUserData(CodeVisionListData.KEY) }
            .flatMap { it.visibleLens }
            .size

        assertTrue("expected many rule definitions in golang.golr", definitionCount > 200)
        assertEquals("every definition should get a usage-count lens", definitionCount, lensCount)
    }
}
