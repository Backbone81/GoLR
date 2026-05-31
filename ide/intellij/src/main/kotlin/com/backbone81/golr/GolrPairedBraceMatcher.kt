package com.backbone81.golr

import com.intellij.lang.BracePair
import com.intellij.lang.PairedBraceMatcher
import com.intellij.psi.PsiFile
import com.intellij.psi.tree.IElementType

class GolrPairedBraceMatcher : PairedBraceMatcher {
    private val pairs = arrayOf(
        BracePair(GolrTokenTypes.LBRACE, GolrTokenTypes.RBRACE, true),
        BracePair(GolrTokenTypes.LPAREN, GolrTokenTypes.RPAREN, false),
    )

    override fun getPairs(): Array<BracePair> = pairs
    override fun isPairedBracesAllowedBeforeType(lbraceType: IElementType, contextType: IElementType?) = true
    override fun getCodeConstructStart(file: PsiFile, openingBraceOffset: Int) = openingBraceOffset
}
