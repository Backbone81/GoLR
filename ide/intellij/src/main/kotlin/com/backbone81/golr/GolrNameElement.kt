package com.backbone81.golr

import com.intellij.extapi.psi.ASTWrapperPsiElement
import com.intellij.lang.ASTNode

// Wraps the single IDENTIFIER token that sits at the definition site of a rule, i.e. the
// "NAME" in NAME : body ;
//
// ASTWrapperPsiElement is IntelliJ's convenience base class for composite PSI nodes: it
// holds the underlying ASTNode and delegates all standard PsiElement operations to it.
//
// This class is a child of GolrSymbolDefinition and its role is narrow:
//   - GolrSymbolDefinition.getNameIdentifier() returns it so IntelliJ knows which text
//     range to highlight when the caret sits on a definition.
//   - Its text is the symbol name used by rename and reference resolution.
class GolrNameElement(node: ASTNode) : ASTWrapperPsiElement(node)
