package com.backbone81.golr

import com.intellij.psi.PsiElement
import com.intellij.psi.PsiReference
import com.intellij.psi.search.SearchScope
import com.intellij.psi.util.PsiTreeUtil
import com.intellij.refactoring.rename.RenamePsiElementProcessor

// Plugs GoLR symbol definitions into IntelliJ's rename refactoring.
//
// IntelliJ's rename refactoring has two phases:
//
//   1. Discovery: collect all PsiReference objects that point to the definition being
//      renamed. By default this is done by ReferencesSearch, which locates files that
//      contain the symbol name by consulting the file word-index (IdIndex). That index
//      is built from registered IdIndexer / WordIndexer implementations and is typically
//      only populated for languages that have an explicit indexer. GoLR does not have
//      one, so ReferencesSearch finds 0 usages and only the single element under the
//      cursor ever changes.
//
//   2. Application: call definition.setName(newName) and then call
//      reference.handleElementRename(newName) for every reference found in phase 1.
//
// This processor replaces phase 1 with a direct PSI tree scan that is index-independent:
// it walks the live PSI tree of the containing file and collects every GolrSymbolReference
// node whose text equals the definition's name.
//
// isInplaceRenameSupported() returning true signals to IntelliJ's MemberInplaceRenameHandler
// that this element type supports inline (in-editor) rename. The handler then highlights
// all usage sites and lets the user type the new name directly in the source instead of
// showing a modal dialog.
//
// Registered in plugin.xml as a renamePsiElementProcessor extension.
class GolrRenamePsiElementProcessor : RenamePsiElementProcessor() {

    // Return true only for GolrSymbolDefinition; all other element types fall back to
    // IntelliJ's default processor.
    override fun canProcessElement(element: PsiElement): Boolean =
        element is GolrSymbolDefinition

    // Scans the containing file's PSI tree for every GolrSymbolReference whose text
    // equals the definition's name. These are the sites that handleElementRename()
    // will update in phase 2.
    //
    // GoLR grammars are self-contained single-file documents, so a file-local scan is
    // sufficient and avoids the complexity of cross-file search.
    override fun findReferences(
        element: PsiElement,
        searchScope: SearchScope,
        searchInCommentsAndStrings: Boolean,
    ): Collection<PsiReference> {
        val definition = element as? GolrSymbolDefinition ?: return emptyList()
        val name = definition.name ?: return emptyList()
        val file = definition.containingFile ?: return emptyList()

        return PsiTreeUtil.findChildrenOfType(file, GolrSymbolReference::class.java)
            .filter { it.text == name }
            .mapNotNull { it.getReference() }
    }

    // Tells IntelliJ's MemberInplaceRenameHandler that inline rename is supported.
    // With this returning true, pressing Shift+F6 activates in-editor rename instead
    // of a modal dialog: all occurrences are highlighted and the user types the new
    // name directly at the caret position.
    override fun isInplaceRenameSupported(): Boolean = true
}
