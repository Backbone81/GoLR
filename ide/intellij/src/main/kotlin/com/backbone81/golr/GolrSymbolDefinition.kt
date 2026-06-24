package com.backbone81.golr

import com.intellij.extapi.psi.ASTWrapperPsiElement
import com.intellij.lang.ASTNode
import com.intellij.psi.PsiElement
import com.intellij.psi.PsiNameIdentifierOwner

// Represents a complete rule definition in a .golr file.  Examples:
//   INTEGER:    /[0-9]+/;              (terminal in @scanner section)
//   expression : term "+" term | term; (nonterminal in @parser section)
//
// Go to Definition:
//   GolrSymbolReference.GolrRef.multiResolve() searches the file for GolrSymbolDefinition
//   nodes whose name matches the reference text, and returns them as the jump targets.
//   IntelliJ lands the caret on getNameIdentifier() when it jumps here.
//
// Rename (Shift+F6):
//   IntelliJ's built-in rename refactoring calls getName() to read the current name,
//   setName() to write the new one, and getNameIdentifier() to position the dialog
//   pre-selection. After setName() returns, IntelliJ iterates every PsiReference that
//   resolves to this definition and calls handleElementRename() on each of them.
//
// Find Usages / usage count:
//   GolrFindUsagesProvider.canFindUsagesFor() returns true for this class, so "Find Usages"
//   is available when the caret is on a definition.  The provider's getDescriptiveName()
//   returns getName() for display in the results panel.
//
// PsiNameIdentifierOwner extends PsiNamedElement (getName/setName) and adds
// getNameIdentifier() — the sub-element that holds the name text.
class GolrSymbolDefinition(node: ASTNode) : ASTWrapperPsiElement(node), PsiNameIdentifierOwner {

    // Returns the GolrNameElement child that contains the defining identifier.
    // IntelliJ uses this node to:
    //   (a) highlight the name when the caret is inside a definition, and
    //   (b) pre-fill the rename dialog with the correct text range.
    override fun getNameIdentifier(): PsiElement? =
        node.findChildByType(GolrElementTypes.NAME_ELEMENT)?.psi

    // Returns the symbol's name (e.g. "expression", "INTEGER").
    override fun getName(): String? = nameIdentifier?.text

    // Called by IntelliJ's rename refactoring with the new name the user confirmed.
    // We create a fresh GolrNameElement from a dummy file, then swap it in for the old one.
    // The PSI tree is modified in-place; IntelliJ handles undo automatically.
    override fun setName(name: String): PsiElement {
        val nameEl = nameIdentifier ?: return this
        val newNameEl = GolrPsiFactory.createNameElement(project, name)
        nameEl.replace(newNameEl)
        return this
    }

    // Returns true when this definition is inside a @scanner { } block (i.e. it defines a
    // terminal), false when it is inside @parser { } (nonterminal).
    // Used by GolrFindUsagesProvider to label results as "terminal" or "nonterminal".
    //
    // Strategy: walk backwards through this node's siblings at the FILE level until we
    // find a @scanner or @parser keyword token. Sections never nest, so the first
    // keyword we encounter is the one this definition belongs to.
    fun isTerminal(): Boolean {
        var prev = node.treePrev
        while (prev != null) {
            if (prev.elementType == GolrTokenTypes.KEYWORD_SECTION) {
                return prev.text == "@scanner"
            }
            prev = prev.treePrev
        }
        return false
    }
}
