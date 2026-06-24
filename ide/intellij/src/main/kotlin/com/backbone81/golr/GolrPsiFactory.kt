package com.backbone81.golr

import com.intellij.openapi.project.Project
import com.intellij.psi.PsiElement
import com.intellij.psi.PsiFileFactory
import com.intellij.psi.util.PsiTreeUtil

// A factory that constructs individual PSI nodes by parsing short in-memory snippets.
//
// Why is a factory needed?
//   PSI nodes are not plain objects you can instantiate with a constructor. Every node
//   must live inside a parsed PSI tree.  When a rename operation needs a new name node
//   (GolrNameElement) or a new reference node (GolrSymbolReference), we create a minimal
//   throwaway .golr file in memory, parse it with the real GolrPsiParser, then extract
//   the specific node from the resulting tree and hand it to the caller. The caller then
//   replaces the old node in the real file's PSI tree using PsiElement.replace().
//
// This is the standard IntelliJ pattern for PSI node creation; it is used the same way
// in the Java, Kotlin, and Python language plugins, among others.
object GolrPsiFactory {

    // Creates a GolrNameElement whose text is `name`.
    // Called by GolrSymbolDefinition.setName() to replace the old name node.
    fun createNameElement(project: Project, name: String): PsiElement {
        // Build a minimal valid @parser rule.  The "x" is a throwaway body symbol.
        val file = createFile(project, "@parser { $name : x ; }")
        val def = PsiTreeUtil.findChildOfType(file, GolrSymbolDefinition::class.java)
            ?: error("GolrPsiFactory: failed to parse name element for '$name'")
        return def.nameIdentifier
            ?: error("GolrPsiFactory: no name identifier found for '$name'")
    }

    // Creates a GolrSymbolReference whose text is `name`.
    // Called by GolrSymbolReference.GolrRef.handleElementRename() to replace the old
    // reference node after a rename.
    fun createSymbolReference(project: Project, name: String): PsiElement {
        // "dummy" is the rule name; `name` ends up as the first body symbol, i.e. a reference.
        val file = createFile(project, "@parser { dummy : $name ; }")
        return PsiTreeUtil.findChildOfType(file, GolrSymbolReference::class.java)
            ?: error("GolrPsiFactory: failed to parse symbol reference for '$name'")
    }

    // Parses `text` as a complete .golr file using the normal language infrastructure.
    // PsiFileFactory.createFileFromText() runs the lexer + parser on the text and returns
    // a PsiFile whose PSI tree we can then mine for specific nodes.
    private fun createFile(project: Project, text: String) =
        PsiFileFactory.getInstance(project)
            .createFileFromText("dummy.golr", GolrLanguageFileType.INSTANCE, text)
}
