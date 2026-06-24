package com.backbone81.golr

import com.intellij.find.findUsages.FindUsagesHandler
import com.intellij.find.findUsages.FindUsagesHandlerFactory
import com.intellij.psi.PsiElement

// Provides a FindUsagesHandler for GoLR symbol definitions.
//
// IntelliJ's Find Usages infrastructure calls FindUsagesManager.getNewFindUsagesHandler()
// when Alt+F7 is pressed. It iterates FindUsagesHandlerFactory extensions and uses the first
// one whose canFindUsages() returns true. Without this factory, IntelliJ might not recognise
// GolrSymbolDefinition as a searchable element.
//
// The handler itself adds no behaviour — it inherits FindUsagesHandlerBase.processElementUsages()
// which calls ReferencesSearch.search(). GolrReferencesSearcher (registered separately) hooks
// into ReferencesSearch to scan the PSI tree directly, so the inherited implementation produces
// confirmed, correctly-grouped results without any override here.
//
// Registered in plugin.xml as a com.intellij.findUsagesHandlerFactory extension.
class GolrFindUsagesHandlerFactory : FindUsagesHandlerFactory() {

    override fun canFindUsages(element: PsiElement): Boolean = element is GolrSymbolDefinition

    override fun createFindUsagesHandler(element: PsiElement, forHighlightUsages: Boolean): FindUsagesHandler =
        GolrFindUsagesHandler(element as GolrSymbolDefinition)
}

// Thin subclass — exists only so FindUsagesHandlerFactory can return a concrete instance.
// All real work is done by GolrReferencesSearcher via the inherited ReferencesSearch path.
private class GolrFindUsagesHandler(definition: GolrSymbolDefinition) : FindUsagesHandler(definition)
