package com.backbone81.golr

import com.intellij.lang.cacheBuilder.DefaultWordsScanner
import com.intellij.lang.cacheBuilder.WordsScanner
import com.intellij.lang.findUsages.FindUsagesProvider
import com.intellij.psi.PsiElement
import com.intellij.psi.tree.TokenSet

// Plugs GoLR into IntelliJ's "Find Usages" infrastructure.
//
// IntelliJ's "Find Usages" works in two phases:
//
//   Phase 1 — word index scan:
//     Before examining any PSI, IntelliJ asks getWordsScanner() how to tokenize files.
//     It uses that scanner to build a word index (think: which files contain the word
//     "expression"?).  This gives it a small candidate set without reading every file.
//
//   Phase 2 — PSI resolution:
//     For each candidate file, IntelliJ finds all GolrSymbolReference elements whose text
//     matches the searched name, calls GolrSymbolReference.GolrRef.multiResolve() on each,
//     and keeps only those that resolve to the target GolrSymbolDefinition.
//     Confirmed hits appear as rows in the "Find Usages" panel.
//
// Note: the "N usages" inlay shown above definitions is a separate Code Vision feature and is
// NOT produced by this provider. It is rendered by GolrReferencesCodeVisionProvider, which
// counts usages via ReferencesSearch. This provider only supplies the Find Usages panel and
// the word scanner.
//
// This class is registered in plugin.xml as a lang.findUsagesProvider extension.
class GolrFindUsagesProvider : FindUsagesProvider {

    // Tells IntelliJ how to scan a .golr file for words during the index phase.
    // DefaultWordsScanner runs our lexer over the file and classifies each token as
    // an identifier word, a comment, or a literal — allowing IntelliJ to add the right
    // words to the right sections of its index.
    override fun getWordsScanner(): WordsScanner =
        DefaultWordsScanner(
            GolrLexer(),
            // IDENTIFIER tokens are the "names" we want to index.  When you search for
            // usages of "expression", only files that contain "expression" as an IDENTIFIER
            // token will be examined in phase 2.
            TokenSet.create(GolrTokenTypes.IDENTIFIER),
            // Comment content is indexed separately so that IntelliJ can optionally include
            // "usages in comments" (the "Search in comments and strings" checkbox).
            TokenSet.create(GolrTokenTypes.COMMENT_LINE, GolrTokenTypes.COMMENT_BLOCK),
            // String literals (inline terminals like "+") could be indexed here too, but
            // we skip them for now because resolving string-to-scanner-rule is not yet
            // implemented.
            TokenSet.EMPTY
        )

    // "Find Usages" is available when the caret is on a GolrSymbolDefinition.
    // It is intentionally not available on GolrSymbolReference — "Find Usages" on a
    // reference is handled by first navigating to its definition (Ctrl+B) and then
    // invoking "Find Usages" from there.
    override fun canFindUsagesFor(element: PsiElement) = element is GolrSymbolDefinition

    // Returns the HTML help topic ID for this provider.  Null means "use the default
    // IntelliJ help page".
    override fun getHelpId(element: PsiElement): String? = null

    // A short type label shown in the "Find Usages" panel header, e.g.:
    //   Usages of terminal 'PLUS'   or   Usages of nonterminal 'expression'
    override fun getType(element: PsiElement): String = when {
        element is GolrSymbolDefinition && element.isTerminal() -> "terminal"
        element is GolrSymbolDefinition -> "nonterminal"
        else -> ""
    }

    // The name used in the "Find Usages" panel header and in the "N usages" tooltip.
    override fun getDescriptiveName(element: PsiElement): String =
        if (element is GolrSymbolDefinition) element.name ?: "" else ""

    // The full text shown for each result row in the panel when useFullName is false,
    // or with extra context when useFullName is true.
    override fun getNodeText(element: PsiElement, useFullName: Boolean): String =
        if (element is GolrSymbolDefinition) element.name ?: "" else element.text
}
