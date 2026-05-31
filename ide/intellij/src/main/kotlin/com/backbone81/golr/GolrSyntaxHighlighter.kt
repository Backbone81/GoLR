package com.backbone81.golr

import com.intellij.lexer.Lexer
import com.intellij.openapi.editor.DefaultLanguageHighlighterColors
import com.intellij.openapi.editor.colors.TextAttributesKey
import com.intellij.openapi.fileTypes.SyntaxHighlighterBase
import com.intellij.psi.tree.IElementType

class GolrSyntaxHighlighter : SyntaxHighlighterBase() {
    companion object {
        val COMMENT = TextAttributesKey.createTextAttributesKey("GOLR_COMMENT", DefaultLanguageHighlighterColors.LINE_COMMENT)
        val KEYWORD_SECTION = TextAttributesKey.createTextAttributesKey("GOLR_KEYWORD_SECTION", DefaultLanguageHighlighterColors.KEYWORD)
        val KEYWORD_CONTROL = TextAttributesKey.createTextAttributesKey("GOLR_KEYWORD_CONTROL", DefaultLanguageHighlighterColors.KEYWORD)
        val REGEX = TextAttributesKey.createTextAttributesKey("GOLR_REGEX", DefaultLanguageHighlighterColors.STRING)
        val STRING = TextAttributesKey.createTextAttributesKey("GOLR_STRING", DefaultLanguageHighlighterColors.STRING)
        val IDENTIFIER = TextAttributesKey.createTextAttributesKey("GOLR_IDENTIFIER", DefaultLanguageHighlighterColors.IDENTIFIER)
        val OPERATION_SIGN = TextAttributesKey.createTextAttributesKey("GOLR_OPERATION_SIGN", DefaultLanguageHighlighterColors.OPERATION_SIGN)
        val SEMICOLON = TextAttributesKey.createTextAttributesKey("GOLR_SEMICOLON", DefaultLanguageHighlighterColors.SEMICOLON)
        val BRACES = TextAttributesKey.createTextAttributesKey("GOLR_BRACES", DefaultLanguageHighlighterColors.BRACES)
        val PARENTHESES = TextAttributesKey.createTextAttributesKey("GOLR_PARENTHESES", DefaultLanguageHighlighterColors.PARENTHESES)
    }

    override fun getHighlightingLexer(): Lexer = GolrLexer()

    override fun getTokenHighlights(tokenType: IElementType): Array<TextAttributesKey> =
        when (tokenType) {
            GolrTokenTypes.COMMENT_LINE,
            GolrTokenTypes.COMMENT_BLOCK  -> pack(COMMENT)
            GolrTokenTypes.KEYWORD_SECTION -> pack(KEYWORD_SECTION)
            GolrTokenTypes.KEYWORD_CONTROL -> pack(KEYWORD_CONTROL)
            GolrTokenTypes.REGEX           -> pack(REGEX)
            GolrTokenTypes.STRING          -> pack(STRING)
            GolrTokenTypes.IDENTIFIER      -> pack(IDENTIFIER)
            GolrTokenTypes.PIPE,
            GolrTokenTypes.COLON           -> pack(OPERATION_SIGN)
            GolrTokenTypes.SEMICOLON       -> pack(SEMICOLON)
            GolrTokenTypes.LBRACE,
            GolrTokenTypes.RBRACE          -> pack(BRACES)
            GolrTokenTypes.LPAREN,
            GolrTokenTypes.RPAREN          -> pack(PARENTHESES)
            else                           -> emptyArray()
        }
}
