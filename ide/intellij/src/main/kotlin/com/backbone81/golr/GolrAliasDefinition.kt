package com.backbone81.golr

import com.intellij.extapi.psi.ASTWrapperPsiElement
import com.intellij.lang.ASTNode
import com.intellij.psi.PsiElement
import com.intellij.psi.PsiNameIdentifierOwner

// Represents the string alias of a terminal, i.e. the "+" in
//   PLUS : "+" ;
// Productions and precedence lines may reference the terminal by this string instead of by
// its name. Such references resolve here, so Go to Definition lands on the string and a
// rename started on the string renames the alias rather than the terminal.
//
// The name is the string including its quotes, as that is the text of every reference to it.
class GolrAliasDefinition(node: ASTNode) : ASTWrapperPsiElement(node), PsiNameIdentifierOwner {

    // The STRING token, which spans the whole element.
    override fun getNameIdentifier(): PsiElement? = firstChild

    override fun getName(): String = text

    override fun setName(name: String): PsiElement =
        replace(GolrPsiFactory.createAliasDefinition(project, name))
}
