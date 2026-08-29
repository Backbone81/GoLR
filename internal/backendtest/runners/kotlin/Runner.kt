// The Kotlin runner for the backend test corpus. It reads an input file, runs the generated scanner and the generated
// parser over it, and writes the canonical scanner trace and parser trace the harness diffs against.
//
// This file has no dependencies, and it must not grow any. It runs in the image with no network, which is what proves
// that generated GoLR code needs nothing but the bare language.

import java.io.File
import java.nio.ByteBuffer
import java.util.Locale
import parser.NonterminalSymbol
import parser.ParseNode
import parser.Parser
import parser.Scanner
import parser.TerminalSymbol
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

// appendScannerTrace scans the whole input and appends one line per event.
private fun appendScannerTrace(lines: MutableList<String>, source: ByteArray, inputPath: String) {
    // The plain Scanner and not the TokenSkipper: a skipped rule matched like any other, and the offsets of the tokens
    // around it are only checkable when it is in the trace.
    val scanner = Scanner(source, inputPath)

    while (scanner.next()) {
        // For a failed match this is the start of the attempt, not the byte which could not be consumed.
        val start = scanner.byteOffset

        if (scanner.token == Token.INVALID_TOKEN) {
            lines.add("ERROR $start")
            continue
        }

        // The end comes from the length of the lexeme rather than from an offset the scanner reports, so that a lexeme
        // which disagrees with the offsets cannot pass unnoticed.
        val lexeme = scanner.lexeme
        val end = start + lexeme.remaining()
        lines.add("TOKEN ${scanner.token} $start $end \"${escapeLexeme(lexeme)}\"")
    }

    // The offset the scanner reports after it ran out of input, not source.size. The two agree only when the scanner
    // consumed everything.
    lines.add("EOF ${scanner.byteOffset}")
}

// appendNodeTrace appends the post-order walk of the subtree, which is the shift and reduce sequence of the parse.
private fun appendNodeTrace(lines: MutableList<String>, node: ParseNode) {
    for (child in node.children) {
        appendNodeTrace(lines, child)
    }

    when (val symbol = node.symbol) {
        is NonterminalSymbol -> lines.add("REDUCE ${symbol.nonterminal} ${node.children.size}")

        // The leaf the recovery pushed where it resumed stands for the dropped input and names no token.
        is TerminalSymbol ->
            if (symbol.token == Token.ERROR_TOKEN) {
                lines.add("RESYNC")
            } else {
                lines.add("SHIFT ${symbol.token}")
            }
    }
}

// appendParserTrace parses the whole input and appends its errors, the walk of the tree and the accept event.
//
// The errors come first because a shift carries no offset to interleave them by. A parse which was given up returns no
// tree, so its trace is the errors alone and the missing accept event is what says the two outcomes apart.
private fun appendParserTrace(lines: MutableList<String>, source: ByteArray, inputPath: String) {
    // The TokenSkipper here, because a skipped rule never reaches the parser.
    val result = Parser().parse(TokenSkipper(Scanner(source, inputPath)))

    for (error in result.errors) {
        lines.add("ERROR ${error.byteOffset}")
    }

    val tree = result.tree ?: return
    appendNodeTrace(lines, tree)
    lines.add("ACCEPT")
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
