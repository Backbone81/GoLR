package com.backbone81.golr

import com.intellij.lang.ASTNode
import com.intellij.lang.ParserDefinition
import com.intellij.lang.PsiParser
import com.intellij.lexer.Lexer
import com.intellij.openapi.project.Project
import com.intellij.psi.FileViewProvider
import com.intellij.psi.PsiElement
import com.intellij.psi.PsiFile
import com.intellij.psi.tree.IFileElementType
import com.intellij.psi.tree.TokenSet

class GolrParserDefinition : ParserDefinition {
    companion object {
        val FILE = IFileElementType(GolrLanguage)
    }

    override fun createLexer(project: Project?): Lexer = GolrLexer()

    override fun createParser(project: Project?): PsiParser = PsiParser { _, builder ->
        val mark = builder.mark()
        while (!builder.eof()) builder.advanceLexer()
        mark.done(FILE)
        builder.treeBuilt
    }

    override fun getFileNodeType(): IFileElementType = FILE

    override fun getCommentTokens(): TokenSet =
        TokenSet.create(GolrTokenTypes.COMMENT_LINE, GolrTokenTypes.COMMENT_BLOCK)

    override fun getStringLiteralElements(): TokenSet =
        TokenSet.create(GolrTokenTypes.STRING, GolrTokenTypes.REGEX)

    override fun createElement(node: ASTNode): PsiElement =
        throw UnsupportedOperationException("No PSI elements defined")

    override fun createFile(viewProvider: FileViewProvider): PsiFile = GolrPsiFile(viewProvider)
}
