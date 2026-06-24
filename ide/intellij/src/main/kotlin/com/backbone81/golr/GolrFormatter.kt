package com.backbone81.golr

import com.intellij.psi.tree.IElementType

// Pure, platform-independent reformatter for .golr files. It turns arbitrary (possibly messy)
// GoLR source into the canonical layout demonstrated by examples/golang/spec/golang.golr:
//
//   @scanner {
//       horizontal_whitespace: /[ \t]/ @fragment;     // bodies column-aligned within a group
//       vertical_whitespace:   /[\r\n]/ @fragment;
//   }
//
//   @parser {
//       @start: SourceFiles;                           // control directives stay single-line
//
//       Rule                                           // parser rule name on its own line
//           : alternative one                          // one alternative per line, ":"/"|"-led
//           | alternative two
//           ;
//   }
//
// The logic lives here as a plain function over text so it can be unit-tested directly and is
// kept independent of IntelliJ's block-based formatting engine. GolrPreFormatProcessor wires it
// into the IDE's "Reformat Code" action.
object GolrFormatter {
    private const val INDENT = "    "

    fun format(text: String): String {
        val tokens = lex(text)
        if (tokens.isEmpty()) return ""

        val out = Output()
        var i = 0
        var first = true
        while (i < tokens.size) {
            val token = tokens[i]
            if (!first && token.blankBefore) out.blank()
            first = false
            i = when {
                token.isComment -> { out.line(0, token.text); i + 1 }
                token.type == GolrTokenTypes.KEYWORD_SECTION -> emitSection(tokens, i, out)
                else -> i + 1 // stray top-level token — skip
            }
        }
        return out.build()
    }

    // ── tokenization ─────────────────────────────────────────────────────────────────────

    // A significant token (whitespace dropped) plus whether the whitespace preceding it
    // contained a blank line (>= 2 newlines), which drives blank-line preservation.
    private data class Tok(val type: IElementType, val text: String, val blankBefore: Boolean) {
        val isComment: Boolean
            get() = type == GolrTokenTypes.COMMENT_LINE || type == GolrTokenTypes.COMMENT_BLOCK
    }

    private fun lex(text: String): List<Tok> {
        val lexer = GolrLexer()
        lexer.start(text, 0, text.length, 0)
        val result = mutableListOf<Tok>()
        var newlines = 0
        while (lexer.tokenType != null) {
            val type = lexer.tokenType!!
            val raw = text.substring(lexer.tokenStart, lexer.tokenEnd)
            if (type == GolrTokenTypes.WHITE_SPACE) {
                newlines += raw.count { it == '\n' }
            } else {
                result.add(Tok(type, raw.trimEnd(), blankBefore = newlines >= 2))
                newlines = 0
            }
            lexer.advance()
        }
        return result
    }

    // ── item model ───────────────────────────────────────────────────────────────────────

    // A section body is parsed into a flat list of these before emission, so scanner rules can
    // be grouped for column alignment and blank lines can be reasoned about per item.
    private sealed class Item {
        abstract val blankBefore: Boolean
    }

    private class CommentItem(val text: String, override val blankBefore: Boolean) : Item()

    // A rule or directive ending in ';'. tokens holds everything up to (but excluding) the ';'.
    private class RuleItem(val tokens: List<Tok>, override val blankBefore: Boolean) : Item()

    // A nested "@precedence { ... }" block.
    private class BlockItem(
        val inner: List<Item>,
        override val blankBefore: Boolean,
        val closingBraceBlankBefore: Boolean,
    ) : Item()

    // Parses items until the RBRACE that closes the enclosing block. Returns the items and the
    // index of that RBRACE (not consumed), or the end index if none is found.
    private fun parseItems(tokens: List<Tok>, start: Int): Pair<List<Item>, Int> {
        val items = mutableListOf<Item>()
        var i = start
        while (i < tokens.size) {
            val token = tokens[i]
            when {
                token.type == GolrTokenTypes.RBRACE -> return items to i

                token.isComment -> {
                    items.add(CommentItem(token.text, token.blankBefore))
                    i++
                }

                token.type == GolrTokenTypes.KEYWORD_CONTROL &&
                    token.text == "@precedence" &&
                    tokens.getOrNull(i + 1)?.type == GolrTokenTypes.LBRACE -> {
                    val blankBefore = token.blankBefore
                    val (inner, afterInner) = parseItems(tokens, i + 2)
                    var next = afterInner
                    var braceBlank = false
                    if (next < tokens.size && tokens[next].type == GolrTokenTypes.RBRACE) {
                        braceBlank = tokens[next].blankBefore
                        next++
                    }
                    items.add(BlockItem(inner, blankBefore, braceBlank))
                    i = next
                }

                else -> {
                    val ruleTokens = mutableListOf<Tok>()
                    val blankBefore = token.blankBefore
                    while (i < tokens.size &&
                        tokens[i].type != GolrTokenTypes.SEMICOLON &&
                        tokens[i].type != GolrTokenTypes.RBRACE
                    ) {
                        ruleTokens.add(tokens[i])
                        i++
                    }
                    if (i < tokens.size && tokens[i].type == GolrTokenTypes.SEMICOLON) i++
                    if (ruleTokens.isNotEmpty()) items.add(RuleItem(ruleTokens, blankBefore))
                }
            }
        }
        return items to i
    }

    // ── emission ─────────────────────────────────────────────────────────────────────────

    private fun emitSection(tokens: List<Tok>, start: Int, out: Output): Int {
        val keyword = tokens[start].text // @scanner or @parser
        out.line(0, "$keyword {")

        var i = start + 1
        if (i < tokens.size && tokens[i].type == GolrTokenTypes.LBRACE) i++

        val (items, afterItems) = parseItems(tokens, i)
        i = afterItems

        if (keyword == "@scanner") emitScannerItems(items, out, 1) else emitParserItems(items, out, 1)

        val brace = tokens.getOrNull(i)
        if (brace != null && brace.blankBefore) out.blank()
        out.line(0, "}")
        return if (brace != null && brace.type == GolrTokenTypes.RBRACE) i + 1 else i
    }

    private fun emitScannerItems(items: List<Item>, out: Output, depth: Int) {
        // Assign each RuleItem a group id: a maximal run of directly consecutive rules (no blank
        // line and no comment between them). Bodies are column-aligned within a group.
        val groupOf = HashMap<RuleItem, Int>()
        var groupId = -1
        var prevWasRule = false
        for (item in items) {
            if (item is RuleItem) {
                if (!prevWasRule || item.blankBefore) groupId++
                groupOf[item] = groupId
                prevWasRule = true
            } else {
                prevWasRule = false
            }
        }
        val groupWidth = HashMap<Int, Int>()
        for (item in items) {
            if (item is RuleItem) {
                val width = nameOf(item).length + 1 // "+ 1" for the colon
                val g = groupOf.getValue(item)
                groupWidth[g] = maxOf(groupWidth[g] ?: 0, width)
            }
        }

        var first = true
        for (item in items) {
            if (!first && item.blankBefore) out.blank()
            first = false
            when (item) {
                is CommentItem -> out.line(depth, item.text)
                is BlockItem -> emitPrecedenceBlock(item, out, depth)
                is RuleItem -> {
                    val prefix = "${nameOf(item)}:".padEnd(groupWidth.getValue(groupOf.getValue(item)))
                    val body = joinSymbols(bodyOf(item))
                    out.line(depth, if (body.isEmpty()) "$prefix;" else "$prefix $body;")
                }
            }
        }
    }

    private fun emitParserItems(items: List<Item>, out: Output, depth: Int) {
        var first = true
        for (item in items) {
            if (!first && item.blankBefore) out.blank()
            first = false
            when (item) {
                is CommentItem -> out.line(depth, item.text)
                is BlockItem -> emitPrecedenceBlock(item, out, depth)
                is RuleItem ->
                    if (item.tokens.first().type == GolrTokenTypes.KEYWORD_CONTROL) {
                        emitDirective(item, out, depth)
                    } else {
                        emitParserRule(item, out, depth)
                    }
            }
        }
    }

    private fun emitParserRule(item: RuleItem, out: Output, depth: Int) {
        out.line(depth, nameOf(item))
        val alternatives = splitByPipe(bodyOf(item))
        alternatives.forEachIndexed { index, alternative ->
            val lead = if (index == 0) ":" else "|"
            val body = joinSymbols(alternative)
            out.line(depth + 1, if (body.isEmpty()) lead else "$lead $body")
        }
        out.line(depth + 1, ";")
    }

    // A single-line directive such as "@start: SourceFiles;" or "@left: \"+\" \"-\";".
    private fun emitDirective(item: RuleItem, out: Output, depth: Int) {
        val keyword = item.tokens.first().text
        val body = joinSymbols(bodyOf(item))
        out.line(depth, if (body.isEmpty()) "$keyword:;" else "$keyword: $body;")
    }

    private fun emitPrecedenceBlock(block: BlockItem, out: Output, depth: Int) {
        out.line(depth, "@precedence {")
        var first = true
        for (item in block.inner) {
            if (!first && item.blankBefore) out.blank()
            first = false
            when (item) {
                is CommentItem -> out.line(depth + 1, item.text)
                is RuleItem -> emitDirective(item, out, depth + 1)
                is BlockItem -> emitPrecedenceBlock(item, out, depth + 1)
            }
        }
        if (block.closingBraceBlankBefore) out.blank()
        out.line(depth, "}")
    }

    // ── helpers ──────────────────────────────────────────────────────────────────────────

    private fun nameOf(item: RuleItem): String = item.tokens.first().text

    // The tokens after the ':' (the rule/directive body).
    private fun bodyOf(item: RuleItem): List<Tok> {
        val colon = item.tokens.indexOfFirst { it.type == GolrTokenTypes.COLON }
        return if (colon >= 0) item.tokens.drop(colon + 1) else item.tokens.drop(1)
    }

    private fun splitByPipe(tokens: List<Tok>): List<List<Tok>> {
        val result = mutableListOf<MutableList<Tok>>(mutableListOf())
        for (token in tokens) {
            if (token.type == GolrTokenTypes.PIPE) {
                result.add(mutableListOf())
            } else {
                result.last().add(token)
            }
        }
        return result
    }

    // Joins body symbols with single spaces, keeping an inline "@precedence(NAME)" annotation
    // tight (no spaces around its parentheses).
    private fun joinSymbols(tokens: List<Tok>): String {
        val builder = StringBuilder()
        var i = 0
        while (i < tokens.size) {
            val token = tokens[i]
            if (token.type == GolrTokenTypes.KEYWORD_CONTROL &&
                token.text == "@precedence" &&
                tokens.getOrNull(i + 1)?.type == GolrTokenTypes.LPAREN
            ) {
                val inner = tokens.getOrNull(i + 2)?.text ?: ""
                if (builder.isNotEmpty()) builder.append(' ')
                builder.append("@precedence(").append(inner).append(')')
                i += 4 // @precedence ( NAME )
            } else {
                if (builder.isNotEmpty()) builder.append(' ')
                builder.append(token.text)
                i++
            }
        }
        return builder.toString()
    }

    // Accumulates output lines, normalizing blank lines (no leading or doubled blanks) and
    // guaranteeing exactly one trailing newline.
    private class Output {
        private val lines = mutableListOf<String>()

        fun line(depth: Int, text: String) {
            lines.add(if (text.isEmpty()) "" else INDENT.repeat(depth) + text)
        }

        fun blank() {
            if (lines.isNotEmpty() && lines.last().isNotEmpty()) lines.add("")
        }

        fun build(): String = lines.joinToString("\n").trimEnd('\n') + "\n"
    }
}
