package com.github.backbone81.golr

import com.intellij.openapi.fileTypes.LanguageFileType
import com.intellij.openapi.util.IconLoader
import javax.swing.Icon

class GolrLanguageFileType : LanguageFileType(GolrLanguage) {
    companion object {
        val INSTANCE = GolrLanguageFileType()

        // IconLoader picks golr_dark.svg on dark themes.
        private val ICON = IconLoader.getIcon("/icons/golr.svg", GolrLanguageFileType::class.java)
    }

    override fun getName() = "GoLR"
    override fun getDescription() = "GoLR grammar file"
    override fun getDefaultExtension() = "golr"
    override fun getIcon(): Icon = ICON
}
