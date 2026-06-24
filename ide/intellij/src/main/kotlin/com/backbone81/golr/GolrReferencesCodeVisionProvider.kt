package com.backbone81.golr

import com.intellij.codeInsight.codeVision.CodeVisionRelativeOrdering
import com.intellij.codeInsight.hints.codeVision.ReferencesCodeVisionProvider
import com.intellij.psi.PsiElement
import com.intellij.psi.PsiFile
import com.intellij.psi.search.searches.ReferencesSearch

// Renders the "N usages" Code Vision indicator above every GoLR symbol definition and
// makes it clickable to open the usages popup.
//
// --- Why this class is required ---
//
// The "N usages" inlay is NOT a free side effect of having Find Usages / ReferencesSearch
// working. It is a separate IntelliJ feature called Code Vision. The platform does not show
// usage counts for arbitrary languages automatically: each language registers its own
// provider. Java has JavaReferencesCodeVisionProvider, Kotlin has its own, and so on. Until
// a provider is registered for GoLR, no usage count is ever drawn — which is exactly why the
// indicator was missing before this class existed.
//
// --- How it works ---
//
// ReferencesCodeVisionProvider is the platform base class for "N usages" Code Vision. It
// already supplies everything except the language-specific decisions:
//   - name            → inherited.
//   - handleClick()   → inherited; calls GotoDeclarationAction.startFindUsages(), which reuses
//                       our Find Usages stack (GolrFindUsagesHandlerFactory +
//                       GolrReferencesSearcher) to show the usages popup on click.
//
// We implement the language-specific hooks:
//   - acceptsFile()    → restrict to .golr files.
//   - acceptsElement() → place the hint on each GolrSymbolDefinition. The hint is anchored
//                        above the line where the definition starts (its name).
//   - getHint()        → compute the count via ReferencesSearch.search(), which is routed
//                        through GolrReferencesSearcher and therefore returns all confirmed
//                        reference sites.
//
// --- Why a dedicated group instead of the platform "Usages" group ---
//
// The base class defaults groupId to the platform "Usages" group (key "references") that Java,
// Kotlin, etc. share. The daemon gates each provider on whether its GROUP is enabled
// (CodeVisionHost: `if (!settings.isProviderEnabled(provider.groupId)) continue`). A developer
// who turns off the noisy Java "N usages" hints therefore also silences GoLR's — the symbol
// counts silently disappear for a reason that has nothing to do with GoLR.
//
// To keep GoLR independent and on by default, we override groupId to our own group and back it
// with GolrCodeVisionGroupSettingProvider so it appears as its own "GoLR usages" toggle in
// Settings | Editor | Inlay Hints | Code Vision. Unknown group ids default to enabled
// (CodeVisionSettings.isProviderEnabled falls back to true), so the hint shows out of the box.
//
// Being a DaemonBoundCodeVisionProvider (via the base class), the count is recomputed by the
// daemon whenever the PSI changes, so it stays in sync as the file is edited.
//
// --- Requires an open project, not a single file ---
//
// The "N usages" count only renders when the .golr file is opened as part of a project (i.e.
// the repository / folder is opened in the IDE). When a lone .golr file is opened on its own,
// the IDE runs in single-file / LightEdit mode, which has no project model and no indexing, so
// the daemon code-insight pass that drives Code Vision never runs and no count appears. This is
// standard JetBrains behavior (the same is true for Java/Kotlin), not a GoLR-specific bug. Open
// the project root rather than an individual file to see usage counts.
//
// Registered in plugin.xml as a com.intellij.codeInsight.daemonBoundCodeVisionProvider.
class GolrReferencesCodeVisionProvider : ReferencesCodeVisionProvider() {

    override fun acceptsFile(file: PsiFile): Boolean = file.language == GolrLanguage

    override fun acceptsElement(element: PsiElement): Boolean = element is GolrSymbolDefinition

    override fun getHint(element: PsiElement, file: PsiFile): String? {
        if (element !is GolrSymbolDefinition) return null
        val count = ReferencesSearch.search(element).findAll().size
        return when (count) {
            0 -> "no usages"
            1 -> "1 usage"
            else -> "$count usages"
        }
    }

    // Unique id for this provider. Used by IntelliJ to persist per-provider enablement and
    // to order providers relative to each other.
    override val id: String get() = "golr.references"

    // Our own Code Vision group (see the class comment), so GoLR usage counts are not affected
    // by the shared platform "Usages" toggle. Backed by GolrCodeVisionGroupSettingProvider.
    override val groupId: String get() = GROUP_ID

    // GoLR has only this one Code Vision provider, so there is nothing to order it against.
    override val relativeOrderings: List<CodeVisionRelativeOrdering> get() = emptyList()

    companion object {
        const val GROUP_ID: String = "golr.usages"
    }
}
