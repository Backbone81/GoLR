package com.backbone81.golr

import com.intellij.psi.tree.IElementType

// IntelliJ represents source code as a tree (the PSI tree) whose nodes each carry a type tag.
// There are two kinds of types:
//   - Token types  (leaves): produced by the lexer, e.g. IDENTIFIER, COLON, SEMICOLON.
//     These live in GolrTokenTypes.
//   - Element types (inner nodes): produced by the parser wrapping a span of tokens.
//     These live here.
//
// The parser (GolrPsiParser) creates inner nodes by calling builder.mark() / marker.done(type).
// GolrParserDefinition.createElement() then maps each element type to the Kotlin class that
// represents it as a PSI object (e.g. SYMBOL_DEFINITION -> GolrSymbolDefinition).
object GolrElementTypes {

    // A complete rule in either section, e.g.:
    //   INTEGER: /[0-9]+/;            (@scanner)
    //   expression : term "+" term ;  (@parser)
    //
    // Mapped to GolrSymbolDefinition, which is the node IntelliJ lands on for:
    //   - "Go to Definition" targets
    //   - Rename refactoring (the element whose name is changed)
    //   - "Find Usages" starting points (the definition being searched for)
    val SYMBOL_DEFINITION: IElementType = IElementType("SYMBOL_DEFINITION", GolrLanguage)

    // The identifier token that names a rule, i.e. the left-hand side of the colon.
    // It is always a direct child of SYMBOL_DEFINITION.
    //
    // Mapped to GolrNameElement. IntelliJ uses this sub-node to know exactly which
    // characters to highlight when the caret sits on a definition, and to know which
    // text range the rename dialog should pre-fill.
    val NAME_ELEMENT: IElementType = IElementType("NAME_ELEMENT", GolrLanguage)

    // An identifier used as a reference to another symbol (right-hand side of a parser
    // rule body, operand in a precedence line, or start-symbol declaration).
    //
    // Mapped to GolrSymbolReference, which overrides getReference() to return a
    // PsiReference object. That object is what IntelliJ calls for:
    //   - Ctrl+B  "Go to Declaration" (PsiReference.resolve())
    //   - Alt+F7  "Find Usages" (reverse-resolving from a definition)
    //   - Shift+F6 rename (PsiReference.handleElementRename())
    val SYMBOL_REFERENCE: IElementType = IElementType("SYMBOL_REFERENCE", GolrLanguage)

    // Structural wrapper for a precedence line such as:   @left : PLUS MINUS ;
    // The SYMBOL_REFERENCE children inside it are the interesting nodes; the wrapper
    // itself exists only so the tree has a meaningful node for the whole line.
    val PRECEDENCE_DECLARATION: IElementType = IElementType("PRECEDENCE_DECLARATION", GolrLanguage)
}
