package com.backbone81.golr

import com.intellij.openapi.actionSystem.CommonDataKeys
import com.intellij.openapi.actionSystem.impl.SimpleDataContext
import com.intellij.refactoring.rename.inplace.MemberInplaceRenameHandler
import com.intellij.refactoring.rename.inplace.VariableInplaceRenameHandler
import com.intellij.testFramework.fixtures.BasePlatformTestCase
import com.intellij.testFramework.fixtures.CodeInsightTestUtil

// Verifies that Shift+F6 on a GoLR symbol performs an INLINE rename (not a modal dialog) and
// updates the definition together with every reference.
class GolrRenameTest : BasePlatformTestCase() {

    // Guards the regression where the rename fell back to the modal dialog: the local-variable
    // inline handler must DECLINE GoLR symbols (otherwise its block-scoped renamer fails and
    // doRename() shows the dialog), and the member inline handler must ACCEPT them.
    fun testMemberInlineHandlerIsSelectedNotTheVariableHandler() {
        myFixture.configureByText(
            "test.golr",
            """
            @parser {
            a<caret> : b b ;
            b : a ;
            }
            """.trimIndent(),
        )
        val element = myFixture.elementAtCaret
        assertInstanceOf(element, GolrSymbolDefinition::class.java)

        val context = SimpleDataContext.builder()
            .add(CommonDataKeys.PSI_ELEMENT, element)
            .add(CommonDataKeys.EDITOR, myFixture.editor)
            .add(CommonDataKeys.PSI_FILE, myFixture.file)
            .add(CommonDataKeys.PROJECT, project)
            .build()

        assertFalse(
            "VariableInplaceRenameHandler must decline GoLR symbols",
            VariableInplaceRenameHandler().isAvailableOnDataContext(context),
        )
        assertTrue(
            "MemberInplaceRenameHandler must handle GoLR symbols",
            MemberInplaceRenameHandler().isAvailableOnDataContext(context),
        )
    }

    fun testInlineRenameFromDefinitionUpdatesAllReferences() {
        myFixture.configureByText(
            "test.golr",
            """
            @parser {
            a<caret> : b b ;
            b : a ;
            }
            """.trimIndent(),
        )
        CodeInsightTestUtil.doInlineRename(MemberInplaceRenameHandler(), "x", myFixture)
        myFixture.checkResult(
            """
            @parser {
            x : b b ;
            b : x ;
            }
            """.trimIndent(),
        )
    }

    fun testInlineRenameFromReferenceUpdatesDefinition() {
        myFixture.configureByText(
            "test.golr",
            """
            @parser {
            a : b b ;
            b : a<caret> ;
            }
            """.trimIndent(),
        )
        CodeInsightTestUtil.doInlineRename(MemberInplaceRenameHandler(), "x", myFixture)
        myFixture.checkResult(
            """
            @parser {
            x : b b ;
            b : x ;
            }
            """.trimIndent(),
        )
    }
}
