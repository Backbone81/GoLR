package com.backbone81.golr

import com.intellij.openapi.application.QueryExecutorBase
import com.intellij.psi.PsiReference
import com.intellij.psi.search.searches.ReferencesSearch
import com.intellij.psi.util.PsiTreeUtil
import com.intellij.util.Processor

// Makes ReferencesSearch.search() work for GoLR symbol definitions.
//
// IntelliJ's reference search infrastructure (ReferencesSearch) lets callers find all
// PsiReferences that point to a given PsiElement. The default implementation locates
// candidate files via the word index, finds IDENTIFIER leaf tokens in those files, then
// calls leaf.getReference(). For GoLR, getReference() lives on the PARENT GolrSymbolReference
// composite node, not on the raw IDENTIFIER leaf, so every candidate is discarded and the
// search returns nothing.
//
// This executor is called by ReferencesSearch whenever the target element is a
// GolrSymbolDefinition. It bypasses the word-index / leaf-token path entirely and instead
// walks the PSI tree directly — the same approach used by GolrRenamePsiElementProcessor for
// rename, but registered at the ReferencesSearch level so all callers benefit automatically.
//
// Plugging into ReferencesSearch fixes three things at once:
//   - Find Usages (Alt+F7): results are now "confirmed" (IntelliJ marks references found via
//     ReferencesSearch as confirmed without any extra work on our side)
//   - Code Vision "N usages" count: GolrReferencesCodeVisionProvider calls
//     ReferencesSearch.search() to compute the number it displays, so the count it shows
//     depends on this executor returning the references. (The provider itself is what draws
//     the inlay — registering this executor alone is not enough to make the count appear.)
//   - Any future caller of ReferencesSearch.search(GolrSymbolDefinition) works automatically
//
// Registered in plugin.xml as a com.intellij.referencesSearch extension.
//
// Note: the features this feeds (Find Usages, the "N usages" Code Vision count) only work when
// the .golr file is opened as part of a project. A lone file opened on its own runs in
// single-file / LightEdit mode, which has no project model or indexing, so ReferencesSearch is
// never driven. This is standard JetBrains behavior — open the project root, not an individual
// file. See GolrReferencesCodeVisionProvider for the full explanation.
//
// We extend QueryExecutorBase with requireReadAction = true so the platform wraps every
// processQuery() call in a read action for us — ReferencesSearch runs on pooled background
// threads, and PSI access from there must hold a read lock. This is the idiomatic form;
// implementing the raw QueryExecutor interface would force us to manage the read action by
// hand.
class GolrReferencesSearcher : QueryExecutorBase<PsiReference, ReferencesSearch.SearchParameters>(true) {

    // Called by ReferencesSearch.search() for every registered executor, already inside a
    // read action. We feed each matching reference to the consumer; if it returns false the
    // search has been cancelled, so we stop early.
    //
    // GoLR grammars are self-contained single-file documents, so scanning the target's
    // containing file is sufficient and we intentionally do not consult
    // params.effectiveSearchScope (there are no cross-file references to find).
    override fun processQuery(
        params: ReferencesSearch.SearchParameters,
        consumer: Processor<in PsiReference>,
    ) {
        val target = params.elementToSearch as? GolrSymbolDefinition ?: return
        val name = target.name ?: return
        val file = target.containingFile ?: return

        for (ref in PsiTreeUtil.findChildrenOfType(file, GolrSymbolReference::class.java)) {
            if (ref.text == name) {
                val reference = ref.reference ?: continue
                if (!consumer.process(reference)) return
            }
        }
    }
}
