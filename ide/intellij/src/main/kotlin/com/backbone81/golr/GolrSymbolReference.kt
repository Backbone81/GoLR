package com.backbone81.golr

import com.intellij.extapi.psi.ASTWrapperPsiElement
import com.intellij.lang.ASTNode
import com.intellij.openapi.util.TextRange
import com.intellij.psi.PsiElement
import com.intellij.psi.PsiElementResolveResult
import com.intellij.psi.PsiReferenceBase
import com.intellij.psi.ResolveResult
import com.intellij.psi.util.PsiTreeUtil

// Represents an identifier that refers to a symbol defined elsewhere.  Examples:
//   expression : term "+" term ;   ← "term" appears twice as a GolrSymbolReference
//   @left : PLUS MINUS ;           ← "PLUS" and "MINUS" are GolrSymbolReferences
//
// Go to Definition (Ctrl+B / Cmd+B):
//   IntelliJ calls getReference() on this element.  The returned GolrRef object's
//   multiResolve() finds the GolrSymbolDefinition(s) with the matching name and returns
//   them. If exactly one definition is found IntelliJ jumps there directly; if several
//   are found it shows a chooser popup.
//
// Rename (Shift+F6):
//   After the user renames a GolrSymbolDefinition, IntelliJ looks up all references that
//   resolve to that definition (by calling multiResolve() on each GolrRef in the project)
//   and calls handleElementRename() on each one to update the reference text.
//
// Find Usages (Alt+F7):
//   IntelliJ searches the file index for tokens matching the symbol name, then calls
//   multiResolve() on each candidate reference to confirm it actually points to the target
//   definition. Every confirmed hit becomes a row in the "Find Usages" panel and is
//   counted toward the "N usages" inlay.
class GolrSymbolReference(node: ASTNode) : ASTWrapperPsiElement(node) {

    // getReference() is how IntelliJ discovers that a PSI element *is* a reference.
    // Returning a non-null PsiReference here opts this element into all reference-based
    // IDE features: navigation, rename, and usage search.
    override fun getReference() = GolrRef(this)

    // PsiReferenceBase.Poly is IntelliJ's base class for references that may resolve to
    // more than one target (poly = multiple). Compared to the simpler PsiReferenceBase,
    // it adds multiResolve() which IntelliJ calls when it needs all possible targets, not
    // just the primary one.
    //
    // The type parameter <GolrSymbolReference> is the PSI element that *has* the reference
    // (the source side of the reference arrow).
    // TextRange(0, element.textLength) tells IntelliJ that the entire text of this node
    // is the "active" part of the reference (the portion that gets underlined on Ctrl+hover
    // and that the rename dialog pre-selects).
    //
    // We must pass this range to the constructor rather than letting PsiReferenceBase
    // calculate it lazily.  The lazy path calls calculateDefaultRangeInElement(), which
    // looks for a registered ElementManipulator for GolrSymbolReference.  We have not
    // registered one (it is not needed for our use case), so the lazy path throws a
    // PluginException.  Providing the range upfront bypasses that code path entirely.
    class GolrRef(element: GolrSymbolReference) : PsiReferenceBase.Poly<GolrSymbolReference>(
        element, TextRange(0, element.textLength), false
    ) {

        // Finds the definition(s) this reference points to.
        //
        // incompleteCode is true when IntelliJ is resolving speculatively (e.g. during
        // completion) and the code may be syntactically incomplete. We ignore it here
        // because GoLR does not need different resolution logic in that case.
        //
        // The return value is an array of ResolveResult. Each entry wraps a PsiElement
        // (the definition) and a validity flag. IntelliJ:
        //   - Jumps directly to the single definition when the array has exactly one entry.
        //   - Shows a chooser popup when there are multiple entries (e.g. the same symbol
        //     is accidentally defined twice).
        //   - Shows an "Unresolved reference" warning when the array is empty.
        override fun multiResolve(incompleteCode: Boolean): Array<ResolveResult> {
            val name = element.text
            val file = element.containingFile
            return PsiTreeUtil.findChildrenOfType(file, GolrSymbolDefinition::class.java)
                .filter { it.name == name }
                .map { PsiElementResolveResult(it) }
                .toTypedArray()
        }

        // Called by IntelliJ's rename refactoring for each reference site after the
        // definition has been renamed. We replace this GolrSymbolReference node with a
        // freshly parsed one carrying the new name, keeping the PSI tree consistent.
        override fun handleElementRename(newElementName: String): PsiElement {
            val newRef = GolrPsiFactory.createSymbolReference(element.project, newElementName)
            return element.replace(newRef)
        }

        // Completion candidates for this reference position. An empty array means
        // no custom completions; IntelliJ could be extended here later to suggest all
        // defined symbol names.
        override fun getVariants(): Array<Any> = emptyArray()
    }
}
