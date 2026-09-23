package com.backbone81.golr

import com.intellij.openapi.actionSystem.CommonDataKeys
import com.intellij.openapi.actionSystem.impl.SimpleDataContext
import com.intellij.psi.util.PsiTreeUtil
import com.intellij.refactoring.rename.RenameUtil
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

    // Renaming a terminal updates references by name but keeps references by string alias.
    fun testInlineRenameKeepsAliasReferences() {
        myFixture.configureByText(
            "test.golr",
            """
            @scanner {
            PLUS<caret> : "+" ;
            }
            @parser {
            a : a "+" a | a PLUS a ;
            }
            """.trimIndent(),
        )
        CodeInsightTestUtil.doInlineRename(MemberInplaceRenameHandler(), "ADD", myFixture)
        myFixture.checkResult(
            """
            @scanner {
            ADD : "+" ;
            }
            @parser {
            a : a "+" a | a ADD a ;
            }
            """.trimIndent(),
        )
    }

    // Renaming a string from a reference renames the terminal's string and every reference by
    // that string, and keeps the terminal name.
    fun testInlineRenameFromAliasReferenceUpdatesAllAliases() {
        myFixture.configureByText(
            "test.golr",
            """
            @scanner {
            PLUS : "+" ;
            }
            @parser {
            @precedence {
            @left : "+" ;
            }
            a : a "<caret>+" a | a PLUS a ;
            }
            """.trimIndent(),
        )
        CodeInsightTestUtil.doInlineRename(MemberInplaceRenameHandler(), "\"plus\"", myFixture)
        myFixture.checkResult(
            """
            @scanner {
            PLUS : "plus" ;
            }
            @parser {
            @precedence {
            @left : "plus" ;
            }
            a : a "plus" a | a PLUS a ;
            }
            """.trimIndent(),
        )
    }

    // Renaming the string in the terminal definition renames every reference by that string.
    fun testInlineRenameFromAliasDefinitionUpdatesAllAliases() {
        myFixture.configureByText(
            "test.golr",
            """
            @scanner {
            PLUS : "<caret>+" ;
            }
            @parser {
            a : a "+" a | a PLUS a ;
            }
            """.trimIndent(),
        )
        CodeInsightTestUtil.doInlineRename(MemberInplaceRenameHandler(), "\"plus\"", myFixture)
        myFixture.checkResult(
            """
            @scanner {
            PLUS : "plus" ;
            }
            @parser {
            a : a "plus" a | a PLUS a ;
            }
            """.trimIndent(),
        )
    }

    // A string can only be renamed to another string, and a name only to another identifier.
    fun testRenameValidatesStringsAndNames() {
        myFixture.configureByText(
            "test.golr",
            """
            @scanner {
            PLUS : "+" ;
            }
            """.trimIndent(),
        )
        val definition = PsiTreeUtil.findChildOfType(myFixture.file, GolrSymbolDefinition::class.java)!!
        val alias = definition.alias()!!
        assertTrue(RenameUtil.isValidName(project, alias, "\"plus\""))
        assertFalse(RenameUtil.isValidName(project, alias, "plus"))
        assertFalse(RenameUtil.isValidName(project, alias, "\"a\" \"b\""))
        assertTrue(RenameUtil.isValidName(project, definition, "ADD"))
        assertFalse(RenameUtil.isValidName(project, definition, "\"plus\""))
    }
}
