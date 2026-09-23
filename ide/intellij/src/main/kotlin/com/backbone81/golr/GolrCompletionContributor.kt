package com.backbone81.golr

import com.intellij.codeInsight.completion.CompletionContributor
import com.intellij.codeInsight.completion.CompletionParameters
import com.intellij.codeInsight.completion.CompletionProvider
import com.intellij.codeInsight.completion.CompletionResultSet
import com.intellij.codeInsight.completion.CompletionType
import com.intellij.codeInsight.lookup.LookupElementBuilder
import com.intellij.patterns.PlatformPatterns
import com.intellij.psi.tree.TokenSet
import com.intellij.psi.util.PsiTreeUtil
import com.intellij.util.ProcessingContext

// Provides basic auto-completion for GoLR files: symbol names where a symbol may be referenced,
// and the keywords the grammar allows at the caret, e.g. @scanner and @parser at the top level
// or @name and @error in a production.
//
// How completion works in IntelliJ:
//   When completion is invoked, the platform inserts a synthetic dummy identifier at the
//   caret, re-parses the file, and then walks the contributors whose pattern matches the
//   leaf at the caret. We match every GoLR leaf and decide from the text before the caret
//   what to offer (see GolrCompletionPosition), as the coarse PSI does not tell a rule body
//   from a precedence line, and a half-typed keyword like "@na" is not a token of its own.
//
// What we suggest:
//   - Keywords valid at the caret position.
//   - In reference positions (rule bodies, precedence lines, @start, @precedence(...)), all
//     names declared by GolrSymbolDefinition nodes in the file — terminals from the @scanner
//     section and nonterminals from the @parser section. The same model that backs Go to
//     Definition / Find Usages / Rename is the source of truth here, so completion never
//     drifts out of sync with resolution.
class GolrCompletionContributor : CompletionContributor() {
    init {
        extend(
            CompletionType.BASIC,
            // withLanguage keeps this from firing inside other languages that might embed
            // .golr fragments.
            PlatformPatterns.psiElement().withLanguage(GolrLanguage),
            GolrCompletionProvider,
        )
    }

    private object GolrCompletionProvider : CompletionProvider<CompletionParameters>() {
        // Leaves whose content is free text, where nothing is completed.
        private val TEXT_TOKENS = TokenSet.create(
            GolrTokenTypes.COMMENT_LINE,
            GolrTokenTypes.COMMENT_BLOCK,
            GolrTokenTypes.STRING,
            GolrTokenTypes.REGEX,
        )

        override fun addCompletions(
            parameters: CompletionParameters,
            context: ProcessingContext,
            result: CompletionResultSet,
        ) {
            if (parameters.position.node.elementType in TEXT_TOKENS) return

            // The platform derives the prefix from the identifier at the caret, which leaves out
            // the "@" of a keyword, so we compute the prefix ourselves.
            val text = parameters.editor.document.charsSequence
            val caret = parameters.offset
            var start = caret
            while (start > 0 && isIdentifierChar(text[start - 1])) start--
            if (start > 0 && text[start - 1] == '@') start--
            val prefix = text.subSequence(start, caret).toString()
            val matcher = result.withPrefixMatcher(prefix)

            val position = GolrCompletionPosition.at(text, start)
            for (keyword in position.keywords) {
                matcher.addElement(LookupElementBuilder.create(keyword).bold())
            }
            if (position.symbols && !prefix.startsWith("@")) {
                addSymbols(parameters, matcher)
            }
        }

        private fun addSymbols(parameters: CompletionParameters, result: CompletionResultSet) {
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

        private fun isIdentifierChar(c: Char) = c.isLetterOrDigit() || c == '_'
    }
}

// Where the caret is in terms of the GoLR grammar, with the keywords valid there and whether
// symbol names may be referenced there.
internal enum class GolrCompletionPosition(val keywords: List<String>, val symbols: Boolean) {
    TOP_LEVEL(listOf("@scanner", "@parser"), false),
    SCANNER_RULE_BODY(listOf("@empty", "@skip", "@fragment"), false),
    PARSER_STATEMENT_START(listOf("@start", "@precedence"), false),
    START_BODY(emptyList(), true),
    PRECEDENCE_LINE_START(listOf("@left", "@right", "@none", "@precedence"), false),
    PRECEDENCE_LINE_BODY(listOf("@error"), true),
    RULE_BODY(listOf("@empty", "@error", "@precedence", "@name"), true),
    PRECEDENCE_ARGUMENT(emptyList(), true),

    // Positions where a new name is written or nothing can be completed, e.g. the name of a
    // rule, between a rule name and its ":", or inside @name(...).
    OTHER(emptyList(), false);

    companion object {
        // Determines the position at `offset` by lexing text[0, offset). Tracking the tokens
        // instead of the PSI keeps this working in the half-typed code completion runs on.
        fun at(text: CharSequence, offset: Int): GolrCompletionPosition {
            var section: String? = null         // "@scanner" or "@parser" once its "{" is seen
            var pendingSection: String? = null  // a section keyword waiting for its "{"
            var inPrecedenceBlock = false
            var statement: String? = null       // first token of the current statement
            var afterColon = false
            var parenKeyword: String? = null    // the keyword before an open "("
            var previous: String? = null

            val lexer = GolrLexer()
            lexer.start(text, 0, offset, 0)
            while (lexer.tokenType != null) {
                val type = lexer.tokenType
                val token = text.subSequence(lexer.tokenStart, lexer.tokenEnd).toString()
                lexer.advance()
                when (type) {
                    GolrTokenTypes.WHITE_SPACE, GolrTokenTypes.COMMENT_LINE, GolrTokenTypes.COMMENT_BLOCK -> continue
                    GolrTokenTypes.KEYWORD_SECTION -> if (section == null) pendingSection = token
                    GolrTokenTypes.LBRACE -> when {
                        section == null && pendingSection != null -> {
                            section = pendingSection
                            pendingSection = null
                        }
                        section == "@parser" && statement == "@precedence" && !afterColon -> {
                            inPrecedenceBlock = true
                            statement = null
                        }
                    }
                    GolrTokenTypes.RBRACE -> {
                        if (inPrecedenceBlock) inPrecedenceBlock = false else section = null
                        statement = null
                        afterColon = false
                    }
                    GolrTokenTypes.SEMICOLON -> {
                        statement = null
                        afterColon = false
                    }
                    GolrTokenTypes.COLON -> afterColon = true
                    GolrTokenTypes.LPAREN -> parenKeyword = previous
                    GolrTokenTypes.RPAREN -> parenKeyword = null
                    else -> if (statement == null) statement = token
                }
                previous = token
            }

            return when {
                section == null -> if (pendingSection == null) TOP_LEVEL else OTHER
                parenKeyword == "@precedence" -> PRECEDENCE_ARGUMENT
                parenKeyword != null -> OTHER
                section == "@scanner" -> if (afterColon) SCANNER_RULE_BODY else OTHER
                inPrecedenceBlock -> when {
                    statement == null -> PRECEDENCE_LINE_START
                    afterColon -> PRECEDENCE_LINE_BODY
                    else -> OTHER
                }
                statement == null -> PARSER_STATEMENT_START
                !afterColon -> OTHER
                statement == "@start" -> START_BODY
                else -> RULE_BODY
            }
        }
    }
}
