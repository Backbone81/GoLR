package com.backbone81.golr

import com.intellij.icons.AllIcons
import com.intellij.openapi.fileTypes.LanguageFileType
import javax.swing.Icon

class GolrLanguageFileType : LanguageFileType(GolrLanguage) {
    companion object {
        val INSTANCE = GolrLanguageFileType()
    }

    override fun getName() = "GoLR"
    override fun getDescription() = "GoLR grammar file"
    override fun getDefaultExtension() = "golr"
    override fun getIcon(): Icon = AllIcons.FileTypes.Text
}
