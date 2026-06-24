package com.backbone81.golr

import com.intellij.codeInsight.codeVision.settings.CodeVisionGroupSettingProvider

// Declares GoLR's own Code Vision group so the usage-count inlay shows up as its own toggle in
// Settings | Editor | Inlay Hints | Code Vision, labelled "GoLR usages", independent of the
// platform-wide "Usages" group that Java/Kotlin share.
//
// Without this, GolrReferencesCodeVisionProvider's custom groupId would still work (unknown
// groups default to enabled), but the settings entry would fall back to the raw resource-bundle
// key "codeLens.golr.usages.name" for its label because no bundle defines it. Providing the
// name and description here gives a clean, self-describing settings entry.
//
// Registered in plugin.xml as a com.intellij.config.codeVisionGroupSettingProvider.
class GolrCodeVisionGroupSettingProvider : CodeVisionGroupSettingProvider {

    override val groupId: String
        get() = GolrReferencesCodeVisionProvider.GROUP_ID

    override val groupName: String
        get() = "GoLR usages"

    override val description: String
        get() = "Shows the number of usages of a GoLR grammar symbol above its definition."
}
