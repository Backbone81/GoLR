package com.backbone81.golr

import com.intellij.codeInsight.codeVision.CodeVisionEntry
import com.intellij.codeInsight.codeVision.ui.model.CodeVisionListData
import com.intellij.codeInsight.codeVision.ui.renderers.CodeVisionInlayRenderer
import com.intellij.openapi.editor.Editor
import com.intellij.testFramework.PlatformTestUtil

// Collects the code vision lenses in the editor. Which inlay kind carries them depends on the
// anchor, which is a global setting other plugins may change, so all three kinds are read.
fun Editor.codeVisionLenses(length: Int): List<CodeVisionEntry> =
    (inlayModel.getBlockElementsInRange(0, length) +
        inlayModel.getAfterLineEndElementsInRange(0, length) +
        inlayModel.getInlineElementsInRange(0, length))
        .filter { it.renderer is CodeVisionInlayRenderer }
        .mapNotNull { it.getUserData(CodeVisionListData.KEY) }
        .flatMap { it.visibleLens }

// CodeVisionHost.calculateCodeVisionSync returns void on 2025.3 but a future on 2026.2, where the
// lenses land after it returns. Its result therefore cannot be awaited from code which compiles
// against both, so wait on the inlay model instead.
fun Editor.awaitCodeVisionLenses(length: Int, expected: Int): List<CodeVisionEntry> {
    PlatformTestUtil.waitWithEventsDispatching(
        "expected $expected code vision lenses to appear",
        { codeVisionLenses(length).size >= expected },
        10,
    )
    return codeVisionLenses(length)
}
