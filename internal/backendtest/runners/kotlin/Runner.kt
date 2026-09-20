// The Kotlin runner for the backend test corpus. It reads an input file, runs the generated scanner and the generated
// parser over it, and writes the canonical scanner trace, parser trace and tree trace the harness diffs against.
//
// This file has no dependencies, and it must not grow any. It runs in the image with no network, which is what proves
// that generated GoLR code needs nothing but the bare language.

import java.io.File
import java.nio.ByteBuffer
import java.util.Locale
import kotlin.system.exitProcess
import parser.Nonterminal
import parser.NonterminalSymbol
import parser.ParseNode
import parser.ParseSymbol
import parser.Parser
import parser.Scanner
import parser.TerminalSymbol
import parser.Token
import parser.TokenSkipper

private const val SCANNER_TRACE_FILE_NAME = "scanner.actual"
private const val PARSER_TRACE_FILE_NAME = "parser.actual"
private const val TREE_TRACE_FILE_NAME = "tree.actual"

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
    writeTrace(TREE_TRACE_FILE_NAME) { lines -> appendTreeTrace(lines, source, inputPath) }
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

// checkPosition holds the offset based position and text against the line, column and lexeme of the token the scanner
// currently sits on. The two ways of asking exist side by side until the release which drops line and column, and the
// corpus is where they have to agree: every case of it is far more input than a hand written test covers.
private fun checkPosition(scanner: Scanner) {
    val position = scanner.position(scanner.byteOffset)
    if (position.line != scanner.line || position.column != scanner.column || position.filePath != scanner.filePath) {
        System.err.println(
            "position(${scanner.byteOffset}) is ${position.filePath} ${position.line}:${position.column}, " +
                "but the scanner reports ${scanner.filePath} ${scanner.line}:${scanner.column}",
        )
        exitProcess(1)
    }

    val lexeme = scanner.lexeme
    if (scanner.text(scanner.byteOffset, lexeme.remaining()) != lexeme) {
        System.err.println("text(${scanner.byteOffset}, ${lexeme.remaining()}) differs from the lexeme")
        exitProcess(1)
    }
}

// appendScannerTrace scans the whole input and appends one line per event: the position the token or the failed match
// starts at, a keyword, and for a token its rule and lexeme, for a failed match the bytes it could not match.
private fun appendScannerTrace(lines: MutableList<String>, source: ByteArray, inputPath: String) {
    // The plain Scanner and not the TokenSkipper: a skipped rule matched like any other, and the position of the
    // tokens around it is only checkable when it is in the trace.
    val scanner = Scanner(source, inputPath)

    while (scanner.next()) {
        checkPosition(scanner)

        val location = "${scanner.line}:${scanner.column}"
        val lexeme = escapeLexeme(scanner.lexeme)

        if (scanner.token == Token.INVALID_TOKEN) {
            lines.add(String.format(Locale.ROOT, "%-7s %-7s \"%s\"", location, "ERROR", lexeme))
            continue
        }
        lines.add(String.format(Locale.ROOT, "%-7s %-7s %s \"%s\"", location, "TOKEN", scanner.token, lexeme))
    }

    // The position after the scanner ran out of input, which is one past the last byte only when it consumed
    // everything. It is the offset every off by one in a line table lands on, so it is checked like a token.
    checkPosition(scanner)
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

// terminalTraceName names a terminal for a trace line, giving the three tokens the grammar cannot spell a dollar
// name.
private fun terminalTraceName(terminal: Token): String = when (terminal) {
    Token.END_TOKEN -> "\$end"
    Token.ERROR_TOKEN -> "\$error"
    Token.INVALID_TOKEN -> "\$invalid"
    else -> terminal.toString()
}

// symbolTraceName is the bare grammar name of a symbol, which for a nonterminal is what toString returns and for a
// terminal is the name the traces spell it with.
private fun symbolTraceName(symbol: ParseSymbol): String = when (symbol) {
    is NonterminalSymbol -> symbol.nonterminal.toString()
    is TerminalSymbol -> terminalTraceName(symbol.token)
}

// reduceTracePayload renders a node as "lhs => rhs", the way the REDUCE line of a parser trace names the production it
// was reduced from, or as "lhs => ε" for a production with an empty right hand side.
private fun reduceTracePayload(lhs: Nonterminal, rhs: List<ParseNode>): String {
    val payload = StringBuilder(lhs.toString()).append(" =>")
    if (rhs.isEmpty()) {
        return payload.append(" ε").toString()
    }
    for (child in rhs) {
        payload.append(' ').append(symbolTraceName(child.symbol))
    }
    return payload.toString()
}

// appendTreeNode appends the line of the given node and the lines of everything below it, which is the pre-order the
// tree trace is read in: a node, then what it was built from. The payload carries the indentation and the position and
// span columns do not, so they stay in the same place however deep a node sits.
private fun appendTreeNode(lines: MutableList<String>, scanner: Scanner, node: ParseNode, depth: Int) {
    val position = scanner.position(node.byteOffset)
    val location = "${position.line}:${position.column}"
    val span = "${node.byteOffset}+${node.byteLength}"

    val payload = StringBuilder("  ".repeat(depth))
    when (val symbol = node.symbol) {
        is NonterminalSymbol -> payload.append(reduceTracePayload(symbol.nonterminal, node.children))
        is TerminalSymbol -> if (symbol.token == Token.ERROR_TOKEN) {
            // The error node stands for no token of its own, so its span is all it carries.
            payload.append(terminalTraceName(symbol.token))
        } else {
            // The text is read off the source through the span and never carried along from the token, which is what
            // makes the trace state that the span is right.
            payload.append(terminalTraceName(symbol.token))
                .append(" \"")
                .append(escapeLexeme(scanner.text(node.byteOffset, node.byteLength)))
                .append('"')
        }
    }

    lines.add(String.format(Locale.ROOT, "%-7s %-7s %s", location, span, payload))
    for (child in node.children) {
        appendTreeNode(lines, scanner, child, depth + 1)
    }
}

// appendTreeTrace parses the whole input and appends one line per node of the tree the parse built, in pre-order. A
// parse which was given up builds no tree and appends nothing, which is the empty trace the harness expects for it.
private fun appendTreeTrace(lines: MutableList<String>, source: ByteArray, inputPath: String) {
    // The scanner stays at hand after the parse, because a node carries the span of the source it covers and not the
    // source itself, so the trace resolves every node through position and text.
    val scanner = Scanner(source, inputPath)

    val result = Parser().parse(TokenSkipper(scanner))
    val tree = result.tree
    if (tree != null) {
        appendTreeNode(lines, scanner, tree, 0)
    }
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
