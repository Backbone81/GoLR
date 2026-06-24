package com.backbone81.golr

import com.intellij.extapi.psi.ASTWrapperPsiElement
import com.intellij.lang.ASTNode
import com.intellij.lang.ParserDefinition
import com.intellij.lexer.Lexer
import com.intellij.openapi.project.Project
import com.intellij.psi.FileViewProvider
import com.intellij.psi.PsiElement
import com.intellij.psi.PsiFile
import com.intellij.psi.tree.IFileElementType
import com.intellij.psi.tree.TokenSet

// ParserDefinition is the central wiring point that IntelliJ calls to obtain the
// lexer, the parser, and the PSI node factory for a language.
//
//   createLexer  → used by syntax highlighting and by GolrFindUsagesProvider's
//                  word scanner to index symbol names.
//   createParser → creates a GolrPsiParser that produces a structured tree with
//                  SYMBOL_DEFINITION, NAME_ELEMENT, SYMBOL_REFERENCE, and
//                  PRECEDENCE_DECLARATION nodes
//   createElement → maps each element type to the Kotlin class that implements
//                   the IDE functionality for that node kind.
class GolrParserDefinition : ParserDefinition {
    companion object {
        val FILE = IFileElementType(GolrLanguage)
    }

    override fun createLexer(project: Project?): Lexer = GolrLexer()

    // Returns the real parser.  The PsiBuilder passed to GolrPsiParser.parse() will
    // automatically skip tokens whose type is in getWhitespaceTokens() and
    // getCommentTokens() (see below), so the parser only sees meaningful tokens.
    override fun createParser(project: Project?) = GolrPsiParser()

    override fun getFileNodeType(): IFileElementType = FILE

    // These two methods tell the PsiBuilder which token types to skip automatically.
    // Without them the parser would need to handle whitespace and comments explicitly
    // everywhere, which would make the parsing code much noisier.
    override fun getWhitespaceTokens(): TokenSet =
        TokenSet.create(GolrTokenTypes.WHITE_SPACE)

    override fun getCommentTokens(): TokenSet =
        TokenSet.create(GolrTokenTypes.COMMENT_LINE, GolrTokenTypes.COMMENT_BLOCK)

    override fun getStringLiteralElements(): TokenSet =
        TokenSet.create(GolrTokenTypes.STRING, GolrTokenTypes.REGEX)

    // Maps each composite element type to its PSI class.  IntelliJ calls this for
    // every non-token node in the tree after the parser finishes. The resulting PSI
    // objects are what the IDE features operate on at runtime:
    //   SYMBOL_DEFINITION   → GolrSymbolDefinition  (rename, find usages, jump target)
    //   NAME_ELEMENT        → GolrNameElement        (name range inside a definition)
    //   SYMBOL_REFERENCE    → GolrSymbolReference    (reference resolution, rename sites)
    //   PRECEDENCE_DECLARATION → plain wrapper, no special behaviour needed
    override fun createElement(node: ASTNode): PsiElement = when (node.elementType) {
        GolrElementTypes.SYMBOL_DEFINITION     -> GolrSymbolDefinition(node)
        GolrElementTypes.NAME_ELEMENT          -> GolrNameElement(node)
        GolrElementTypes.SYMBOL_REFERENCE      -> GolrSymbolReference(node)
        GolrElementTypes.PRECEDENCE_DECLARATION -> ASTWrapperPsiElement(node)
        else -> throw UnsupportedOperationException("Unknown element type: ${node.elementType}")
    }

    override fun createFile(viewProvider: FileViewProvider): PsiFile = GolrPsiFile(viewProvider)
}
