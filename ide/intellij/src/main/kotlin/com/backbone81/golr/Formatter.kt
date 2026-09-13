package com.backbone81.golr

import com.backbone81.golr.generated.Scanner
import com.backbone81.golr.generated.Token
import java.io.ByteArrayOutputStream
import java.nio.ByteBuffer
import java.nio.charset.StandardCharsets

// Structural port of internal/fmt/formatter.go. Kept as close as possible to that file's shape
// (same fields, same onTokenX methods, same two-pass structure) so a fix made there has a
// same-named counterpart here to apply it to. The only intentional deviations are things Kotlin
// itself dictates: currentContext() returns a nullable Token instead of using an "invalid token"
// sentinel, the generated Scanner exposes properties (scanner.token/scanner.lexeme) rather than
// methods, and lexeme bytes arrive as a ByteBuffer rather than a byte slice (see lexemeBytes()).
//
// Format pretty prints source in two passes: token-by-token emission with lazily scheduled line
// breaks (see linebreak/blankLine/emit), followed by alignScannerRules, which pads scanner rule
// ":" columns using the marks collected during the first pass.
class Formatter(private val config: Config = DefaultConfig) {
    private lateinit var scanner: Scanner
    private var indentLevel: Int = 0

    // The length of the most recently emitted identifier inside a @scanner section, i.e. the
    // left hand side of the scanner rule whose ":" comes next.
    private var lastScannerIdentifierLen: Int = 0

    // Collects scanner rule marks seen during the first pass, grouped by contiguous runs of
    // rules not separated by an explicit blank line. It always has at least one (possibly empty)
    // group; a new one is started as soon as an explicit blank line is seen inside a @scanner
    // section, and it stays empty if no rule follows before the next one. Consumed by
    // alignScannerRules afterwards.
    private var scannerRuleGroups: MutableList<MutableList<ScannerRuleRHS>> = mutableListOf()

    // Reports if the next output needs to be indented or not.
    private var indentNext: Boolean = false

    // Reports if a linebreak was requested but not yet written to the output. Requesting a
    // linebreak is lazy so that a same-line trailing comment can still be emitted before it.
    private var pendingLinebreak: Boolean = false

    // Reports if a blank line before the next emitted content has already been scheduled, so a
    // second, independent reason to want one (e.g. automatic spacing between rules and a user's
    // own blank line around a comment coinciding) does not stack into two. Cleared once real
    // content is emitted.
    private var pendingBlankLine: Boolean = false

    // Reports if the next emit should not write a whitespace to separate the previous token from
    // the next one.
    private var emitTight: Boolean = false

    // A stack which describes the current nesting.
    private val context: MutableList<Token> = mutableListOf()

    // Reports if the whitespace just consumed contained a blank line, i.e. the user separated the
    // surrounding tokens by an empty line of their own.
    private var explicitBlankLine: Boolean = false

    // Reports if the whitespace just consumed contained a linebreak, i.e. the next token starts
    // on a new source line instead of trailing the previous token.
    private var explicitNewline: Boolean = false

    // Reports if the token just emitted was a comment, so a following explicit blank line can be
    // attributed to it instead of to whatever token comes after the whitespace.
    private var explicitComment: Boolean = false

    private lateinit var output: ByteArrayOutputStream

    // Notes down where a scanner rule's right hand side starts, so a second pass can align it
    // with the other rules in the same group.
    private data class ScannerRuleRHS(
        // The byte position in the output right after the rule's ":", where padding is inserted.
        val offset: Int,
        // The length of the rule's identifier.
        val ruleNameLen: Int,
    )

    fun format(source: ByteArray, filePath: String): ByteArray {
        indentLevel = 0
        indentNext = true
        pendingLinebreak = false
        pendingBlankLine = false
        emitTight = false
        context.clear()
        explicitBlankLine = false
        explicitNewline = false
        explicitComment = false
        lastScannerIdentifierLen = 0
        scannerRuleGroups = mutableListOf(mutableListOf())

        output = ByteArrayOutputStream(source.size)
        scanner = Scanner(source, filePath)
        while (scanner.next()) {
            // explicitComment must be reset before every token, not just at the bottom of the
            // loop like explicitBlankLine/explicitNewline, so it reflects only the token from the
            // immediately preceding iteration; TOKEN_WHITESPACE below still needs that old value,
            // so grab a copy first.
            val explicitCommentBefore = explicitComment
            explicitComment = false

            when (scanner.token) {
                Token.TOKEN_WHITESPACE -> {
                    onTokenWhitespace(explicitCommentBefore)
                    continue
                }
                Token.TOKEN_COMMENT -> onTokenComment()
                Token.TOKEN_COLON -> onTokenColon()
                Token.TOKEN_PIPE -> onTokenPipe()
                Token.TOKEN_SEMI -> onTokenSemi()
                Token.TOKEN_LBRACE -> onTokenLbrace()
                Token.TOKEN_RBRACE -> onTokenRbrace()
                Token.TOKEN_LPAREN -> onTokenLparen()
                Token.TOKEN_RPAREN -> onTokenRparen()
                Token.TOKEN_IDENTIFIER -> onTokenIdentifier()
                Token.TOKEN_SCANNER -> onTokenScanner()
                Token.TOKEN_PARSER -> onTokenParser()
                Token.TOKEN_START -> onTokenStart()
                Token.TOKEN_PRECEDENCE -> onTokenPrecedence()
                else -> emit(lexemeBytes())
            }

            explicitBlankLine = false
            explicitNewline = false
            // explicitComment is intentionally not reset here; see the reset at the top of the loop.
        }
        ensureTrailingNewline()
        return alignScannerRules(output.toByteArray())
    }

    // Guarantees that non-empty output ends with exactly one "\n", regardless of which token
    // ended the input (a well-formed "}" leaves a blank line pending that would otherwise
    // contribute a second, unwanted trailing "\n", while a file that just stops mid-rule leaves
    // nothing pending at all).
    private fun ensureTrailingNewline() {
        val bytes = output.toByteArray()
        if (bytes.isEmpty()) {
            return
        }
        if (bytes[bytes.size - 1] != NEWLINE_BYTE) {
            output.write(NEWLINE_BYTE.toInt())
        }
    }

    private fun onTokenWhitespace(explicitCommentBefore: Boolean) {
        // We are looking for linebreaks and explicit blank lines by the user.
        val newlines = lexemeBytes().count { it == NEWLINE_BYTE }
        if (newlines >= 1) {
            explicitNewline = true
        }
        if (newlines >= 2) {
            explicitBlankLine = true
            if (explicitCommentBefore) {
                // The user separated the comment we just emitted from what follows with a blank
                // line; keep it instead of collapsing the comment onto the next token.
                blankLine()
            }
            if (currentContext() == Token.TOKEN_SCANNER) {
                // Start a new alignment group. It stays empty if no rule follows before the next
                // blank line, which is harmless.
                scannerRuleGroups.add(mutableListOf())
            }
        }
    }

    private fun onTokenComment() {
        val lexeme = lexemeBytes()
        val isLineComment = lexeme.size >= 2 && lexeme[0] == SLASH_BYTE && lexeme[1] == SLASH_BYTE
        when {
            explicitNewline -> {
                // The comment starts on its own line rather than trailing the previous token.
                if (explicitBlankLine) {
                    // The user separated the comment from the previous content with a blank
                    // line; keep it.
                    blankLine()
                } else {
                    linebreak()
                }
                emit(lexeme)
            }
            pendingLinebreak -> {
                // The comment trails the previous token, but a linebreak (and possibly a blank
                // line) is already scheduled to run after that token. Emit the comment before
                // that instead of flushing it early. indentNext is forced false for the emit: it
                // is only true here because the scheduled linebreak hasn't fired yet, and leaving
                // it true would make emit treat the comment as opening a fresh indented line
                // (writing the indent string as a separator) instead of trailing the current one.
                val hadBlankLine = pendingBlankLine
                pendingLinebreak = false
                pendingBlankLine = false
                indentNext = false
                emit(lexeme)
                pendingLinebreak = true
                pendingBlankLine = hadBlankLine
                indentNext = true
            }
            else -> emit(lexeme)
        }
        if (isLineComment || explicitNewline) {
            // Anything after "//" on the same line would otherwise be swallowed into the
            // comment. A comment that started on its own line must not have the following
            // content glued onto it either, even for a block comment which doesn't force this
            // lexically.
            linebreak()
        }
        // Read by the top of the loop on the next token, to detect a blank line right after this
        // comment.
        explicitComment = true
    }

    private fun onTokenColon() {
        when (currentContext()) {
            Token.TOKEN_SCANNER -> {
                // Note down where the rule's right hand side starts, so the second pass can
                // align it.
                val mark = ScannerRuleRHS(
                    offset = output.size() + 1,
                    ruleNameLen = lastScannerIdentifierLen,
                )
                scannerRuleGroups.last().add(mark)

                emitTight = true
                emit(lexemeBytes())
            }
            Token.TOKEN_PARSER -> {
                // Indent the alternatives under the rule name; matching indentDec runs at the
                // rule's ";".
                indentInc()
                linebreak()
                emit(lexemeBytes())
            }
            else -> {
                emitTight = true
                emit(lexemeBytes())
            }
        }
    }

    private fun onTokenPipe() {
        linebreak()
        emit(lexemeBytes())
    }

    private fun onTokenSemi() {
        when (currentContext()) {
            Token.TOKEN_PARSER -> {
                linebreak()
                emit(lexemeBytes())
                indentDec()
                // Separate top-level rules with a blank line; idempotent, so it composes with a
                // blank line a trailing comment on the next rule also wants (see onTokenComment).
                blankLine()
            }
            Token.TOKEN_START -> {
                // "@start: X;" stays on one line rather than being spread out like a regular rule.
                emitTight = true
                emit(lexemeBytes())
                popContext()
                blankLine()
            }
            else -> {
                emitTight = true
                emit(lexemeBytes())
                linebreak()
            }
        }
    }

    private fun onTokenLbrace() {
        emit(lexemeBytes())
        linebreak()
        indentInc()
    }

    private fun onTokenRbrace() {
        indentLevel = maxOf(indentLevel - 1, 0)
        // A blank line may still be pending to separate the previous rule from a following one
        // (see onTokenSemi); it must not leak into a blank line before the closing brace itself.
        pendingBlankLine = false
        if (!indentNext) {
            // Skip if we're already at the start of a fresh line (e.g. an empty "{}" block),
            // otherwise this would add a spurious blank line before "}".
            linebreak()
        }
        emit(lexemeBytes())
        blankLine()

        // We remove any context we did push onto the context stack.
        popContext()
    }

    private fun onTokenLparen() {
        if (currentContext() == Token.TOKEN_PRECEDENCE) {
            // "@precedence" is ambiguous: "@precedence { ... }" opens a block, but
            // "@precedence(...)" is an inline alternative annotation with no matching "}".
            // onTokenPrecedence below pushes eagerly as if it were the block form; seeing a "("
            // instead of a "{" proves it wasn't, so undo it.
            popContext()
        }

        // Left parenthesis always sit close to the previous token.
        emitTight = true
        emit(lexemeBytes())

        // The following token also needs to sit close to the left parenthesis.
        emitTight = true
    }

    private fun onTokenRparen() {
        // Right parenthesis always sit close to the previous token.
        emitTight = true
        emit(lexemeBytes())
    }

    private fun onTokenIdentifier() {
        when (currentContext()) {
            Token.TOKEN_SCANNER -> {
                // This identifier is the left hand side of a scanner rule; remember its length
                // for the ":" that follows right after.
                lastScannerIdentifierLen = scanner.lexeme.remaining()
                if (explicitBlankLine) {
                    // Preserve the user's blank line between scanner rules.
                    blankLine()
                }
            }
            else -> {}
        }
        emit(lexemeBytes())
    }

    private fun onTokenScanner() {
        // Track that we're inside @scanner; popped again on the matching "}".
        emit(lexemeBytes())
        pushContext(Token.TOKEN_SCANNER)
    }

    private fun onTokenParser() {
        // Track that we're inside @parser; popped again on the matching "}".
        emit(lexemeBytes())
        pushContext(Token.TOKEN_PARSER)
    }

    private fun onTokenStart() {
        // Track that we're inside "@start ... ;"; popped again on the matching ";" so the colon
        // and semicolon are formatted inline instead of like a regular rule's.
        emit(lexemeBytes())
        pushContext(Token.TOKEN_START)
    }

    private fun onTokenPrecedence() {
        // Pushed eagerly as if opening a "@precedence { ... }" block; onTokenLparen above repairs
        // this if it turns out to be the inline "@precedence(...)" annotation instead.
        emit(lexemeBytes())
        pushContext(Token.TOKEN_PRECEDENCE)
    }

    private fun indentInc() {
        indentLevel++
    }

    private fun indentDec() {
        indentLevel = maxOf(indentLevel - 1, 0)
    }

    // Requests a linebreak before the next emitted content. Like blankLine, it is lazy (the
    // actual "\n" is written by emit, so that a same-line trailing comment can still be inserted
    // before it) and idempotent, so a second, independent reason to want one (e.g. a trailing
    // comment already scheduling one, followed by a token that unconditionally wants one too)
    // does not stack into an unwanted blank line.
    private fun linebreak() {
        pendingLinebreak = true
        indentNext = true
    }

    // Requests a blank line before the next emitted content. Like linebreak, it is lazy: the
    // actual "\n\n" is written by emit. It is idempotent, so a second, independent reason to want
    // a blank line (e.g. automatic spacing between rules and a user's own blank line around a
    // comment coinciding) does not stack into two.
    private fun blankLine() {
        pendingLinebreak = true
        pendingBlankLine = true
        indentNext = true
    }

    // Flushes any pending linebreak or blank line, indents if needed, and writes data.
    private fun emit(data: ByteArray) {
        when {
            // A pending blank line implies a pending linebreak too (see blankLine), so check it
            // first.
            pendingBlankLine -> output.write("\n\n".toByteArray(StandardCharsets.UTF_8))
            pendingLinebreak -> output.write("\n".toByteArray(StandardCharsets.UTF_8))
        }
        pendingLinebreak = false
        pendingBlankLine = false
        if (indentNext) {
            repeat(indentLevel) {
                output.write(config.indentation.toByteArray(StandardCharsets.UTF_8))
            }
            emitTight = true
        }
        if (!emitTight) {
            output.write(" ".toByteArray(StandardCharsets.UTF_8))
        }
        indentNext = false
        emitTight = false
        output.write(data)
    }

    private fun pushContext(token: Token) {
        context.add(token)
    }

    private fun popContext() {
        if (context.isNotEmpty()) {
            context.removeAt(context.size - 1)
        }
    }

    private fun currentContext(): Token? = context.lastOrNull()

    // The second pass over the pretty printed output. It pads the ":" of scanner rules within
    // each group (a run of rules not separated by an explicit blank line) so their right hand
    // sides start in the same column.
    private fun alignScannerRules(data: ByteArray): ByteArray {
        val result = ByteArrayOutputStream(data.size)
        var lastOffset = 0
        for (group in scannerRuleGroups) {
            var maxPrefixLen = 0
            for (mark in group) {
                maxPrefixLen = maxOf(maxPrefixLen, mark.ruleNameLen)
            }

            for (mark in group) {
                result.write(data, lastOffset, mark.offset - lastOffset)
                repeat(maxPrefixLen - mark.ruleNameLen) {
                    result.write(' '.code)
                }
                lastOffset = mark.offset
            }
        }
        result.write(data, lastOffset, data.size - lastOffset)
        return result.toByteArray()
    }

    // The generated Scanner exposes the current lexeme as a ByteBuffer view rather than a byte
    // slice; this copies it out so it can be handed to emit() the same way formatter.go hands
    // f.scanner.Lexeme() to f.emit() directly.
    private fun lexemeBytes(): ByteArray {
        val buffer: ByteBuffer = scanner.lexeme
        val bytes = ByteArray(buffer.remaining())
        buffer.get(bytes)
        return bytes
    }

    private companion object {
        private val NEWLINE_BYTE: Byte = '\n'.code.toByte()
        private val SLASH_BYTE: Byte = '/'.code.toByte()
    }
}
