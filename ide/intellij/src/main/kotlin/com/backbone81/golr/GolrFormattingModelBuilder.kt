package com.backbone81.golr

import com.intellij.formatting.Block
import com.intellij.formatting.ChildAttributes
import com.intellij.formatting.FormattingContext
import com.intellij.formatting.FormattingModel
import com.intellij.formatting.FormattingModelBuilder
import com.intellij.formatting.FormattingModelProvider
import com.intellij.formatting.Indent
import com.intellij.lang.ASTNode
import com.intellij.openapi.util.TextRange

// Registers GoLR as a formattable language so the "Reformat Code" pipeline runs for .golr files.
//
// The actual reformatting is performed by GolrPreFormatProcessor before this model is built; the
// model itself is intentionally a no-op: it exposes the file as a single opaque leaf block with
// no spacing or indentation rules, so the block engine leaves the already-formatted text alone.
//
// Registered in plugin.xml as a com.intellij.lang.formatter extension.
class GolrFormattingModelBuilder : FormattingModelBuilder {
    override fun createModel(formattingContext: FormattingContext): FormattingModel {
        val file = formattingContext.containingFile
        val root = GolrRootBlock(file.node)
        return FormattingModelProvider.createFormattingModelForPsiFile(file, root, formattingContext.codeStyleSettings)
    }

    // A single leaf block spanning the whole file. With no sub-blocks and no spacing, the
    // formatting engine has nothing to change.
    private class GolrRootBlock(private val node: ASTNode) : Block {
        override fun getTextRange(): TextRange = node.textRange
        override fun getSubBlocks(): List<Block> = emptyList()
        override fun getWrap() = null
        override fun getIndent(): Indent = Indent.getNoneIndent()
        override fun getAlignment() = null
        override fun getSpacing(child1: Block?, child2: Block) = null
        override fun getChildAttributes(newChildIndex: Int) = ChildAttributes(Indent.getNoneIndent(), null)
        override fun isIncomplete() = false
        override fun isLeaf() = true
    }
}
