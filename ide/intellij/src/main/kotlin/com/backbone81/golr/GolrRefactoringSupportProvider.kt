package com.backbone81.golr

import com.intellij.lang.refactoring.RefactoringSupportProvider
import com.intellij.psi.PsiElement

// Declares which refactorings GoLR symbols support. For us that is just inline rename.
//
// --- Why this is the correct approach ---
//
// IntelliJ ships generic inline-rename handlers (VariableInplaceRenameHandler and
// MemberInplaceRenameHandler) that already know how to drive an in-editor rename: they set
// up the live template, link the definition to all its usage sites, and commit via
// PsiNamedElement.setName() / PsiReference.handleElementRename(). The only thing they ask of
// a language is a yes/no answer to "is inline rename available for this element?".
//
// They get that answer from LanguageRefactoringSupport.forContext(element), i.e. the
// RefactoringSupportProvider registered for the element's language. By registering this
// provider and returning true for GolrSymbolDefinition, the built-in handler activates for
// GoLR with no custom handler and no dependency on platform internals.
//
// --- Member rename, not variable rename ---
//
// There are two built-in inline handlers, tried in registration order:
//   1. VariableInplaceRenameHandler — for locals; gated by isInplaceRenameAvailable().
//      Its VariableInplaceRenamer is scoped to a local code block and CANNOT rename a
//      file-level symbol; when it can't, doRename() falls back to the modal dialog.
//   2. MemberInplaceRenameHandler — for members; gated by isMemberInplaceRenameAvailable().
//      Its MemberInplaceRenamer renames file/project-wide and handles our symbols correctly.
//
// GoLR symbols are file-level declarations, i.e. "members", so we opt into ONLY the member
// path: isInplaceRenameAvailable stays false (its default) so the variable handler declines
// and does not grab the rename, and isMemberInplaceRenameAvailable returns true so the member
// handler takes it. This mirrors what JavaRefactoringSupportProvider does for methods/fields.
// (Returning true from both is the bug that previously sent every rename to the dialog.)
//
// This replaces the previous GolrInplaceRenameHandler, which subclassed
// MemberInplaceRenameHandler and overrode isAvailable()/doRename() to work around the missing
// provider — even reasoning about a specific bytecode offset of the superclass. That approach
// was brittle against platform changes; this is the documented, supported path.
//
// How the rest of the rename flow is wired:
//   - The element to rename is resolved by the platform from the caret via TargetElementUtil:
//       * caret on a definition name  → the GolrSymbolDefinition (a PsiNameIdentifierOwner)
//       * caret on a reference        → its resolved GolrSymbolDefinition
//   - GolrRenamePsiElementProcessor then supplies the usage sites (findReferences) and opts
//     into inline rename (isInplaceRenameSupported = true).
//
// Registered in plugin.xml as a lang.refactoringSupport extension.
class GolrRefactoringSupportProvider : RefactoringSupportProvider() {

    // Intentionally NOT overriding isInplaceRenameAvailable: it must stay false so the
    // local-variable inline handler declines GoLR symbols (see the class comment).

    override fun isMemberInplaceRenameAvailable(element: PsiElement, context: PsiElement?): Boolean =
        element is GolrSymbolDefinition
}
