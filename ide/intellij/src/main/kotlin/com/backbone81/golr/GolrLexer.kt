package com.backbone81.golr

import com.intellij.lexer.LexerBase
import com.intellij.psi.tree.IElementType

class GolrLexer : LexerBase() {
    private var buffer: CharSequence = ""
    private var bufferEnd: Int = 0
    private var tokenStart: Int = 0
    private var tokenEnd: Int = 0
    private var tokenType: IElementType? = null
    private var state: Int = STATE_DEFAULT

    companion object {
        private const val STATE_DEFAULT = 0
        private const val STATE_IN_BLOCK_COMMENT = 1

        private val SECTION_KEYWORDS = setOf("scanner", "parser")
        private val CONTROL_KEYWORDS =
            setOf("skip", "fragment", "empty", "start", "left", "right", "none", "precedence")
    }

    override fun start(buffer: CharSequence, startOffset: Int, endOffset: Int, initialState: Int) {
        this.buffer = buffer
        this.bufferEnd = endOffset
        this.tokenStart = startOffset
        this.tokenEnd = startOffset
        this.state = initialState
        this.tokenType = null
        advance()
    }

    override fun getState(): Int = state
    override fun getTokenType(): IElementType? = tokenType
    override fun getTokenStart(): Int = tokenStart
    override fun getTokenEnd(): Int = tokenEnd
    override fun getBufferSequence(): CharSequence = buffer
    override fun getBufferEnd(): Int = bufferEnd

    override fun advance() {
        tokenStart = tokenEnd
        if (tokenStart >= bufferEnd) {
            tokenType = null
            return
        }
        tokenType = readNextToken()
    }

    private fun peekChar(offset: Int): Char = if (offset < bufferEnd) buffer[offset] else '\u0000'

    private fun readNextToken(): IElementType {
        if (state == STATE_IN_BLOCK_COMMENT) return readBlockCommentTail()

        val currChar = peekChar(tokenStart)

        return when (currChar) {
            ' ', '\t', '\r', '\n' -> readWhitespace()
            '/' -> when {
                peekChar(tokenStart + 1) == '/' -> readLineComment()
                peekChar(tokenStart + 1) == '*' -> readBlockComment()
                else -> readRegex()
            }

            '"' -> readString()
            '@' -> readAnnotation()
            else -> when {
                currChar.isLetter() || currChar == '_' -> readIdentifier()
                else -> readPunctuation(currChar)
            }
        }
    }

    private fun readWhitespace(): IElementType {
        tokenEnd = tokenStart + 1
        while (tokenEnd < bufferEnd && isWhitespaceChar(peekChar(tokenEnd))) tokenEnd++
        return GolrTokenTypes.WHITE_SPACE
    }

    private fun isWhitespaceChar(c: Char) = c == ' ' || c == '\t' || c == '\r' || c == '\n'

    private fun readLineComment(): IElementType {
        tokenEnd = tokenStart + 2
        while (tokenEnd < bufferEnd && peekChar(tokenEnd) != '\n') tokenEnd++
        return GolrTokenTypes.COMMENT_LINE
    }

    private fun readBlockComment(): IElementType {
        tokenEnd = tokenStart + 2
        return readBlockCommentTail()
    }

    private fun readBlockCommentTail(): IElementType {
        state = STATE_IN_BLOCK_COMMENT
        while (tokenEnd < bufferEnd) {
            if (peekChar(tokenEnd) == '*' && peekChar(tokenEnd + 1) == '/') {
                tokenEnd += 2
                state = STATE_DEFAULT
                break
            }
            tokenEnd++
        }
        return GolrTokenTypes.COMMENT_BLOCK
    }

    private fun readRegex(): IElementType {
        tokenEnd = tokenStart + 1
        while (tokenEnd < bufferEnd) {
            val rc = peekChar(tokenEnd)
            if (rc == '\\') {
                tokenEnd = minOf(tokenEnd + 2, bufferEnd); continue
            }
            if (rc == '/') {
                tokenEnd++; break
            }
            if (rc == '\n') break
            tokenEnd++
        }
        return GolrTokenTypes.REGEX
    }

    private fun readString(): IElementType {
        tokenEnd = tokenStart + 1
        while (tokenEnd < bufferEnd) {
            val rc = peekChar(tokenEnd)
            if (rc == '\\') {
                tokenEnd = minOf(tokenEnd + 2, bufferEnd); continue
            }
            if (rc == '"') {
                tokenEnd++; break
            }
            if (rc == '\n') break
            tokenEnd++
        }
        return GolrTokenTypes.STRING
    }

    private fun readAnnotation(): IElementType {
        tokenEnd = tokenStart + 1
        while (tokenEnd < bufferEnd && peekChar(tokenEnd).isLetter()) tokenEnd++
        val keyword = buffer.substring(tokenStart + 1, tokenEnd)
        return when (keyword) {
            in SECTION_KEYWORDS -> GolrTokenTypes.KEYWORD_SECTION
            in CONTROL_KEYWORDS -> GolrTokenTypes.KEYWORD_CONTROL
            else -> GolrTokenTypes.BAD_CHARACTER
        }
    }

    private fun readIdentifier(): IElementType {
        tokenEnd = tokenStart + 1
        while (tokenEnd < bufferEnd && (peekChar(tokenEnd).isLetterOrDigit() || peekChar(tokenEnd) == '_')) tokenEnd++
        return GolrTokenTypes.IDENTIFIER
    }

    private fun readPunctuation(currChar: Char): IElementType {
        tokenEnd = tokenStart + 1
        return when (currChar) {
            ':' -> GolrTokenTypes.COLON
            ';' -> GolrTokenTypes.SEMICOLON
            '|' -> GolrTokenTypes.PIPE
            '{' -> GolrTokenTypes.LBRACE
            '}' -> GolrTokenTypes.RBRACE
            '(' -> GolrTokenTypes.LPAREN
            ')' -> GolrTokenTypes.RPAREN
            else -> GolrTokenTypes.BAD_CHARACTER
        }
    }
}

