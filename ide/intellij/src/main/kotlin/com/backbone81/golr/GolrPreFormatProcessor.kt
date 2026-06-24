package com.backbone81.golr

import com.intellij.lang.ASTNode
import com.intellij.openapi.util.TextRange
import com.intellij.psi.PsiFileFactory
import com.intellij.psi.impl.source.codeStyle.PreFormatProcessor

// Bridges GolrFormatter into IntelliJ's "Reformat Code" action (Ctrl+Alt+L).
//
// The IDE's formatting engine is block based: it only adjusts whitespace between existing PSI
// nodes. GoLR's canonical layout, however, restructures content (rule name onto its own line,
// each alternative onto its own line, column-aligned scanner bodies), which the block engine
// cannot express against our coarse PSI. A PreFormatProcessor runs before the block engine and
// is allowed to rewrite the PSI, so we do the whole reformat here and leave the block engine a
// no-op (see GolrFormattingModelBuilder).
//
// Registered in plugin.xml as a com.intellij.preFormatProcessor extension.
class GolrPreFormatProcessor : PreFormatProcessor {
    override fun process(element: ASTNode, range: TextRange): TextRange {
        val file = element.psi?.containingFile as? GolrPsiFile ?: return range

        // Only handle whole-file reformatting. A partial-range reformat would otherwise reflow
        // the entire file, surprising the user, so we leave those untouched.
        if (range.startOffset > 0 || range.endOffset < file.textLength) return range

        val original = file.text
        val formatted = GolrFormatter.format(original)
        if (formatted == original) return range

        // Replace the file's PSI with a freshly parsed, formatted copy. Editing the AST (rather
        // than the document directly) keeps the change inside the formatting write action and
        // lets IntelliJ track it for undo.
        val newFile = PsiFileFactory.getInstance(file.project)
            .createFileFromText("dummy.golr", GolrLanguageFileType.INSTANCE, formatted)
        file.node.replaceAllChildrenToChildrenOf(newFile.node)

        return TextRange(0, formatted.length)
    }
}
