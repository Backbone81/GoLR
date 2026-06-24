package com.backbone81.golr

import com.intellij.psi.PsiElement
import com.intellij.usages.impl.rules.UsageType
import com.intellij.usages.impl.rules.UsageTypeProvider

// Assigns a display group name to GoLR symbol usages in the Find Usages panel.
//
// IntelliJ categorises each found usage into a group using registered UsageTypeProvider
// extensions. Each provider's getUsageType() is called with the PSI element at the usage
// site. The first non-null result wins; if no provider recognises the element, IntelliJ
// falls back to "Unclassified".
//
// Without this provider, every GolrSymbolReference usage appears under "Unclassified"
// because IntelliJ has no built-in knowledge of GoLR nodes.
//
// Registered in plugin.xml as a com.intellij.usageTypeProvider extension.
class GolrUsageTypeProvider : UsageTypeProvider {

    override fun getUsageType(element: PsiElement): UsageType? =
        if (element is GolrSymbolReference) SYMBOL_REFERENCE else null

    companion object {
        // The string passed to UsageType becomes the group header text shown in the
        // Find Usages panel (e.g. "Symbol reference  2 usages").
        val SYMBOL_REFERENCE = UsageType("Symbol reference")
    }
}
