package com.backbone81.golr

import com.intellij.codeInsight.completion.CompletionContributor
import com.intellij.codeInsight.completion.CompletionParameters
import com.intellij.codeInsight.completion.CompletionProvider
import com.intellij.codeInsight.completion.CompletionResultSet
import com.intellij.codeInsight.completion.CompletionType
import com.intellij.codeInsight.lookup.LookupElementBuilder
import com.intellij.patterns.PlatformPatterns
import com.intellij.psi.util.PsiTreeUtil
import com.intellij.util.ProcessingContext

// Provides basic ("smart") auto-completion for GoLR files: when the user is typing an
// identifier, every symbol defined anywhere in the same file is offered as a suggestion.
//
// How completion works in IntelliJ:
//   When completion is invoked, the platform inserts a synthetic dummy identifier at the
//   caret, re-parses the file, and then walks the contributors whose pattern matches the
//   leaf at the caret. Because GolrPsiParser wraps a body identifier as SYMBOL_REFERENCE
//   and a leading identifier as NAME_ELEMENT, the caret leaf is always the raw IDENTIFIER
//   token underneath one of those composite nodes. Matching on the IDENTIFIER token type
//   therefore fires in exactly the positions where a symbol name is meaningful — rule
//   bodies, precedence lines, the @start declaration, and (harmlessly) definition sites.
//
// What we suggest:
//   All names declared by GolrSymbolDefinition nodes in the file — terminals from the
//   @scanner section and nonterminals from the @parser section. The same model that backs
//   Go to Definition / Find Usages / Rename (GolrSymbolDefinition) is the source of truth
//   here, so completion never drifts out of sync with resolution.
class GolrCompletionContributor : CompletionContributor() {
    init {
        extend(
            CompletionType.BASIC,
            // Match the IDENTIFIER leaf token in GoLR files. withLanguage keeps this from
            // firing inside other languages that might embed .golr fragments.
            PlatformPatterns.psiElement(GolrTokenTypes.IDENTIFIER).withLanguage(GolrLanguage),
            GolrSymbolCompletionProvider,
        )
    }

    private object GolrSymbolCompletionProvider : CompletionProvider<CompletionParameters>() {
        override fun addCompletions(
            parameters: CompletionParameters,
            context: ProcessingContext,
            result: CompletionResultSet,
        ) {
            // parameters.originalFile is the real file (not the synthetic completion copy),
            // which is what the user sees and what holds all the definitions.
            val file = parameters.originalFile

            // Collect every defined symbol name. A name may legitimately appear only once as
            // a definition, but we de-duplicate defensively in case a file declares the same
            // symbol twice (which the resolver also tolerates).
            val seen = HashSet<String>()
            for (definition in PsiTreeUtil.findChildrenOfType(file, GolrSymbolDefinition::class.java)) {
                val name = definition.name ?: continue
                if (!seen.add(name)) continue

                // "terminal" / "nonterminal" tail text mirrors the label GolrFindUsagesProvider
                // uses, so the same vocabulary shows up across all features.
                val typeText = if (definition.isTerminal()) "terminal" else "nonterminal"
                result.addElement(
                    LookupElementBuilder.create(name)
                        .withTypeText(typeText, true),
                )
            }
        }
    }
}
