// The Kotlin runner for the backend test corpus. It reads an input file, runs the generated scanner and the generated
// parser over it, and writes the canonical scanner trace and parser trace the harness diffs against.
//
// This file has no dependencies, and it must not grow any. It runs in the image with no network, which is what proves
// that generated GoLR code needs nothing but the bare language.

import java.io.File
import java.nio.ByteBuffer
import java.util.Locale
import parser.Parser
import parser.Scanner
import parser.Token
import parser.TokenSkipper

private const val SCANNER_TRACE_FILE_NAME = "scanner.actual"
private const val PARSER_TRACE_FILE_NAME = "parser.actual"

// The bytes a trace line carries as they are. Everything outside of it is escaped.
private const val PRINTABLE_LOW = 0x20
private const val PRINTABLE_HIGH = 0x7e

fun main(args: Array<String>) {
    require(args.size == 1) { "the runner takes the input file as its only argument" }
    val inputPath = args[0]

    // readBytes and not readText, because the generated scanner wants the bytes. Handing it a decoded string would be
    // the classic mistake this harness exists to catch.
    val source = File(inputPath).readBytes()

    writeTrace(SCANNER_TRACE_FILE_NAME) { lines -> appendScannerTrace(lines, source, inputPath) }
    writeTrace(PARSER_TRACE_FILE_NAME) { lines -> appendParserTrace(lines, source, inputPath) }
}

// escapeLexeme escapes the bytes of a lexeme. The caller writes the quotes around the result.
//
// The lexeme is escaped byte by byte, never character by character, so a multi byte UTF-8 sequence becomes one \xHH
// escape per byte. Decoding it first would report UTF-16 code unit offsets and disagree with every other backend.
private fun escapeLexeme(lexeme: ByteBuffer): String {
    val result = StringBuilder()
    while (lexeme.hasRemaining()) {
        val value = lexeme.get().toInt() and 0xff
        when (value) {
            0x5c -> result.append("\\\\") // backslash
            0x22 -> result.append("\\\"") // quote
            0x0a -> result.append("\\n")
            0x0d -> result.append("\\r")
            0x09 -> result.append("\\t")
            in PRINTABLE_LOW..PRINTABLE_HIGH -> result.append(value.toChar())
            // The trace asks for a zero padded pair of lower case hex digits.
            else -> result.append(String.format(Locale.ROOT, "\\x%02x", value))
        }
    }
    return result.toString()
}

// appendScannerTrace scans the whole input and appends one line per event: the position the token or the failed match
// starts at, a keyword, and for a token its rule and lexeme, for a failed match the bytes it could not match.
private fun appendScannerTrace(lines: MutableList<String>, source: ByteArray, inputPath: String) {
    // The plain Scanner and not the TokenSkipper: a skipped rule matched like any other, and the position of the
    // tokens around it is only checkable when it is in the trace.
    val scanner = Scanner(source, inputPath)

    while (scanner.next()) {
        val location = "${scanner.line}:${scanner.column}"
        val lexeme = escapeLexeme(scanner.lexeme)

        if (scanner.token == Token.INVALID_TOKEN) {
            lines.add(String.format(Locale.ROOT, "%-7s %-7s \"%s\"", location, "ERROR", lexeme))
            continue
        }
        lines.add(String.format(Locale.ROOT, "%-7s %-7s %s \"%s\"", location, "TOKEN", scanner.token, lexeme))
    }

    // The position after the scanner ran out of input, which is one past the last byte only when it consumed
    // everything.
    lines.add(String.format(Locale.ROOT, "%-7s %s", "${scanner.line}:${scanner.column}", "EOF"))
}

// appendParserTrace parses the whole input and appends the line the parser's trace hook emits for every action. The
// hook reports the error recovery steps too, which the returned tree does not, so the tree and the error are ignored.
private fun appendParserTrace(lines: MutableList<String>, source: ByteArray, inputPath: String) {
    val parser = Parser()
    parser.trace = { line -> lines.add(line) }

    // The TokenSkipper here, because a skipped rule never reaches the parser.
    parser.parse(TokenSkipper(Scanner(source, inputPath)))
}

// writeTrace produces one trace and writes it to its file. Whatever was produced before a failure is written all the
// same, so a runner which breaks half way still says how far it got, and the other trace is still produced.
private fun writeTrace(fileName: String, produce: (MutableList<String>) -> Unit) {
    val lines = mutableListOf<String>()
    try {
        produce(lines)
    } catch (error: Throwable) {
        // A Throwable and not an Exception, so that a stack overflow on a deep tree leaves the trace it got to.
        System.err.println("producing $fileName failed:")
        error.printStackTrace()
    }

    // Every line is terminated, and an empty trace is an empty file rather than a bare newline. The line ending is
    // spelled out, because the trace is LF whatever the platform would use.
    File(fileName).writeText(lines.joinToString(separator = "") { line -> line + "\n" })
}
