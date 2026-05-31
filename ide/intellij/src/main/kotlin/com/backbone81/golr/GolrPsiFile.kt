package com.backbone81.golr

import com.intellij.extapi.psi.PsiFileBase
import com.intellij.psi.FileViewProvider

class GolrPsiFile(viewProvider: FileViewProvider) : PsiFileBase(viewProvider, GolrLanguage) {
    override fun getFileType() = GolrLanguageFileType.INSTANCE
}
