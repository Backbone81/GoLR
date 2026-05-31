package com.backbone81.golr

import com.intellij.psi.TokenType
import com.intellij.psi.tree.IElementType

object GolrTokenTypes {
    val WHITE_SPACE: IElementType = TokenType.WHITE_SPACE
    val BAD_CHARACTER: IElementType = TokenType.BAD_CHARACTER

    val COMMENT_LINE = IElementType("COMMENT_LINE", GolrLanguage)
    val COMMENT_BLOCK = IElementType("COMMENT_BLOCK", GolrLanguage)

    val KEYWORD_SECTION = IElementType("KEYWORD_SECTION", GolrLanguage)
    val KEYWORD_CONTROL = IElementType("KEYWORD_CONTROL", GolrLanguage)

    val REGEX = IElementType("REGEX", GolrLanguage)
    val STRING = IElementType("STRING", GolrLanguage)
    val IDENTIFIER = IElementType("IDENTIFIER", GolrLanguage)

    val COLON = IElementType("COLON", GolrLanguage)
    val SEMICOLON = IElementType("SEMICOLON", GolrLanguage)
    val PIPE = IElementType("PIPE", GolrLanguage)
    val LBRACE = IElementType("LBRACE", GolrLanguage)
    val RBRACE = IElementType("RBRACE", GolrLanguage)
    val LPAREN = IElementType("LPAREN", GolrLanguage)
    val RPAREN = IElementType("RPAREN", GolrLanguage)
}
