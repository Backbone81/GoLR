package com.backbone81.golr

import com.backbone81.golr.generated.Scanner
import com.backbone81.golr.generated.Token
import com.intellij.lexer.LexerBase
import com.intellij.psi.tree.IElementType

// Adapts the GoLR-generated DFA scanner (generated/Scanner.kt, from
// internal/parsergen/frontend/golr/spec/golr.golr via `make generate`) to IntelliJ's Lexer.
//
// The scanner works on a ByteArray with byte offsets; IntelliJ works on a CharSequence with
// character offsets, so `start` UTF-8 encodes the buffer and builds offset maps both ways.
//
// getState is always 0: the DFA restarts at every token boundary, so any boundary is a safe
// resume point for an incremental relex.
class GolrLexer : LexerBase() {
    private var buffer: CharSequence = ""
    private var bufferEnd: Int = 0

    private val scanner = Scanner(ByteArray(0), "")

    // buffer[0, bufferEnd) as UTF-8, plus offset maps both ways (each one slot longer than
    // its input so the end offset maps too).
    private var bytes: ByteArray = ByteArray(0)
    private var byteToChar: IntArray = IntArray(1)
    private var charToByte: IntArray = IntArray(1)

    private var tokenType: IElementType? = null
    private var tokenStart: Int = 0
    private var tokenEnd: Int = 0

    override fun start(buffer: CharSequence, startOffset: Int, endOffset: Int, initialState: Int) {
        this.buffer = buffer
        this.bufferEnd = endOffset

        val text = buffer.subSequence(0, endOffset).toString()
        bytes = text.toByteArray(Charsets.UTF_8)
        buildOffsetMaps(text)

        scanner.reset(bytes, charToByte[startOffset])
        advance()
    }

    override fun getState(): Int = 0
    override fun getTokenType(): IElementType? = tokenType
    override fun getTokenStart(): Int = tokenStart
    override fun getTokenEnd(): Int = tokenEnd
    override fun getBufferSequence(): CharSequence = buffer
    override fun getBufferEnd(): Int = bufferEnd

    override fun advance() {
        if (!scanner.next()) {
            // Nothing left to lex.
            tokenType = null
            tokenStart = bufferEnd
            tokenEnd = bufferEnd
            return
        }

        val startByte = scanner.byteOffset
        val endByte = startByte + scanner.lexeme.remaining()
        tokenStart = byteToChar[startByte]
        tokenEnd = byteToChar[endByte]
        tokenType = mapToken(scanner.token, startByte)
    }

    // Maps a generated token to the plugin's IElementType. `startByte` is used to split a
    // COMMENT into line vs. block by its opening bytes.
    private fun mapToken(token: Token, startByte: Int): IElementType = when (token) {
        Token.TOKEN_WHITESPACE -> GolrTokenTypes.WHITE_SPACE

        Token.TOKEN_COMMENT ->
            if (startByte + 1 < bytes.size && bytes[startByte] == SLASH && bytes[startByte + 1] == SLASH) {
                GolrTokenTypes.COMMENT_LINE
            } else {
                GolrTokenTypes.COMMENT_BLOCK
            }

        Token.TOKEN_SCANNER, Token.TOKEN_PARSER -> GolrTokenTypes.KEYWORD_SECTION

        Token.TOKEN_PRECEDENCE, Token.TOKEN_START, Token.TOKEN_LEFT, Token.TOKEN_RIGHT,
        Token.TOKEN_NONE, Token.TOKEN_SKIP, Token.TOKEN_EMPTY, Token.TOKEN_ERROR,
        Token.TOKEN_FRAGMENT -> GolrTokenTypes.KEYWORD_CONTROL

        Token.TOKEN_LBRACE -> GolrTokenTypes.LBRACE
        Token.TOKEN_RBRACE -> GolrTokenTypes.RBRACE
        Token.TOKEN_LPAREN -> GolrTokenTypes.LPAREN
        Token.TOKEN_RPAREN -> GolrTokenTypes.RPAREN
        Token.TOKEN_COLON -> GolrTokenTypes.COLON
        Token.TOKEN_SEMI -> GolrTokenTypes.SEMICOLON
        Token.TOKEN_PIPE -> GolrTokenTypes.PIPE

        Token.TOKEN_NAME -> GolrTokenTypes.IDENTIFIER
        Token.TOKEN_REGEX -> GolrTokenTypes.REGEX
        Token.TOKEN_STRING -> GolrTokenTypes.STRING

        // COMMA is defined but unused in the grammar; INVALID_TOKEN is any byte that starts no
        // token. Both are errors. ERROR_TOKEN and END_TOKEN never come from input.
        Token.TOKEN_COMMA, Token.INVALID_TOKEN, Token.ERROR_TOKEN, Token.END_TOKEN ->
            GolrTokenTypes.BAD_CHARACTER
    }

    // One pass over the code points: every char of a code point maps to its first byte and
    // every byte maps back to its first char. Token boundaries never fall inside a code point,
    // so that resolution is enough.
    private fun buildOffsetMaps(text: String) {
        charToByte = IntArray(text.length + 1)
        byteToChar = IntArray(bytes.size + 1)

        var charIdx = 0
        var byteIdx = 0
        while (charIdx < text.length) {
            val codePoint = text.codePointAt(charIdx)
            val charCount = Character.charCount(codePoint)
            val byteCount = utf8ByteCount(codePoint)
            for (k in 0 until charCount) charToByte[charIdx + k] = byteIdx
            for (k in 0 until byteCount) byteToChar[byteIdx + k] = charIdx
            charIdx += charCount
            byteIdx += byteCount
        }
        charToByte[text.length] = bytes.size
        byteToChar[bytes.size] = text.length
    }

    private fun utf8ByteCount(codePoint: Int): Int = when {
        codePoint < 0x80 -> 1
        codePoint < 0x800 -> 2
        codePoint < 0x10000 -> 3
        else -> 4
    }

    private companion object {
        private const val SLASH: Byte = '/'.code.toByte()
    }
}
