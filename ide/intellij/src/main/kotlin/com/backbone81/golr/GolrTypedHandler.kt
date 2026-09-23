package com.backbone81.golr

import com.intellij.codeInsight.AutoPopupController
import com.intellij.codeInsight.editorActions.TypedHandlerDelegate
import com.intellij.openapi.editor.Editor
import com.intellij.openapi.project.Project
import com.intellij.psi.PsiFile

// Opens the completion popup when "@" is typed, so keywords are offered right away. The
// platform only does so on its own for identifier characters.
//
// Registered in plugin.xml as a com.intellij.typedHandler extension.
class GolrTypedHandler : TypedHandlerDelegate() {

    override fun checkAutoPopup(charTyped: Char, project: Project, editor: Editor, file: PsiFile): Result {
        if (file !is GolrPsiFile || charTyped != '@') return Result.CONTINUE
        AutoPopupController.getInstance(project).scheduleAutoPopup(editor)
        return Result.STOP
    }
}
