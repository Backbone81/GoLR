package com.backbone81.golr

import com.intellij.openapi.project.Project
import com.intellij.patterns.ElementPattern
import com.intellij.patterns.PlatformPatterns
import com.intellij.psi.PsiElement
import com.intellij.refactoring.rename.RenameInputValidatorEx
import com.intellij.util.ProcessingContext

// Validates the new name when renaming a string alias. For GolrAliasDefinition it takes
// precedence over GolrNamesValidator, which only accepts identifiers.
//
// Registered in plugin.xml as a com.intellij.renameInputValidator extension.
class GolrAliasInputValidator : RenameInputValidatorEx {

    override fun getPattern(): ElementPattern<out PsiElement> =
        PlatformPatterns.psiElement(GolrAliasDefinition::class.java)

    override fun isInputValid(newName: String, element: PsiElement, context: ProcessingContext): Boolean =
        isStringLiteral(newName)

    override fun getErrorMessage(newName: String, project: Project): String? =
        if (isStringLiteral(newName)) null else "'$newName' is not a string literal"

    // The new name must lex as exactly one STRING token.
    private fun isStringLiteral(text: String): Boolean {
        val lexer = GolrLexer()
        lexer.start(text)
        return lexer.tokenType == GolrTokenTypes.STRING && lexer.tokenEnd == text.length
    }
}
