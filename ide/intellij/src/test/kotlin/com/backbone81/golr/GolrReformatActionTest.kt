package com.backbone81.golr

import com.intellij.openapi.command.WriteCommandAction
import com.intellij.openapi.util.TextRange
import com.intellij.psi.codeStyle.CodeStyleManager
import com.intellij.testFramework.fixtures.BasePlatformTestCase

// Verifies that the IDE's "Reformat Code" action is wired to GolrFormatter through
// GolrPreFormatProcessor: running the real reformat pipeline produces the canonical layout.
class GolrReformatActionTest : BasePlatformTestCase() {

    private fun reformat(input: String, expected: String) {
        myFixture.configureByText("test.golr", input.trimIndent())
        WriteCommandAction.runWriteCommandAction(project) {
            CodeStyleManager.getInstance(project)
                .reformatText(myFixture.file, listOf(TextRange(0, myFixture.file.textLength)))
        }
        myFixture.checkResult(expected)
    }

    fun testReformatExpandsParserRule() {
        reformat(
            """
            @parser {
            Operand : Literal | OperandName ;
            }
            """,
            "@parser {\n" +
                "    Operand\n" +
                "        : Literal\n" +
                "        | OperandName\n" +
                "        ;\n" +
                "}\n",
        )
    }

    fun testReformatAlignsScannerBodies() {
        reformat(
            """
            @scanner {
            add: "+";
            shift_left: "<<";
            }
            """,
            "@scanner {\n" +
                "    add:        \"+\";\n" +
                "    shift_left: \"<<\";\n" +
                "}\n",
        )
    }
}
