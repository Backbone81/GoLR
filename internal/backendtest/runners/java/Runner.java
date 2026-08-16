// The Java runner for the backend test corpus. It reads an input file, runs the generated scanner and the generated
// parser over it, and writes the canonical scanner trace and parser trace the harness diffs against.
//
// This file has no dependencies, and it must not grow any. It runs in the image with no network, which is what proves
// that generated GoLR code needs nothing but the bare language.

import java.io.IOException;
import java.nio.ByteBuffer;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import java.util.Locale;

import parser.Parser;
import parser.Scanner;

public class Runner {
    private static final String SCANNER_TRACE_FILE_NAME = "scanner.actual";
    private static final String PARSER_TRACE_FILE_NAME = "parser.actual";

    // The bytes a trace line carries as they are. Everything outside of it is escaped.
    private static final int PRINTABLE_LOW = 0x20;
    private static final int PRINTABLE_HIGH = 0x7e;

    public static void main(String[] args) throws IOException {
        if (args.length != 1) {
            throw new IllegalArgumentException("the runner takes the input file as its only argument");
        }
        String inputPath = args[0];

        // readAllBytes and not readString, because the generated scanner wants the bytes. Handing it a decoded string
        // would be the classic mistake this harness exists to catch.
        byte[] source = Files.readAllBytes(Path.of(inputPath));

        writeTrace(SCANNER_TRACE_FILE_NAME, lines -> appendScannerTrace(lines, source, inputPath));
        writeTrace(PARSER_TRACE_FILE_NAME, lines -> appendParserTrace(lines, source, inputPath));
    }

    // escapeLexeme escapes the bytes of a lexeme. The caller writes the quotes around the result.
    //
    // The lexeme is escaped byte by byte, never character by character, so a multi byte UTF-8 sequence becomes one
    // \xHH escape per byte. Decoding it first would report UTF-16 code unit offsets and disagree with every other
    // backend.
    private static String escapeLexeme(ByteBuffer lexeme) {
        StringBuilder result = new StringBuilder();
        while (lexeme.hasRemaining()) {
            int value = lexeme.get() & 0xff;
            switch (value) {
                case 0x5c -> result.append("\\\\"); // backslash
                case 0x22 -> result.append("\\\""); // quote
                case 0x0a -> result.append("\\n");
                case 0x0d -> result.append("\\r");
                case 0x09 -> result.append("\\t");
                default -> {
                    if (PRINTABLE_LOW <= value && value <= PRINTABLE_HIGH) {
                        result.append((char) value);
                    } else {
                        // The trace asks for a zero padded pair of lower case hex digits.
                        result.append(String.format(Locale.ROOT, "\\x%02x", value));
                    }
                }
            }
        }
        return result.toString();
    }

    // appendScannerTrace scans the whole input and appends one line per event.
    private static void appendScannerTrace(List<String> lines, byte[] source, String inputPath) {
        // The plain Scanner and not the TokenSkipper: a skipped rule matched like any other, and the offsets of the
        // tokens around it are only checkable when it is in the trace.
        Scanner scanner = new Scanner(source, inputPath);

        while (scanner.next()) {
            // For a failed match this is the start of the attempt, not the byte which could not be consumed.
            int start = scanner.byteOffset();

            if (scanner.token() == Scanner.Token.INVALID_TOKEN) {
                lines.add("ERROR " + start);
                continue;
            }

            // The end comes from the length of the lexeme rather than from an offset the scanner reports, so that a
            // lexeme which disagrees with the offsets cannot pass unnoticed.
            ByteBuffer lexeme = scanner.lexeme();
            int end = start + lexeme.remaining();
            lines.add("TOKEN " + scanner.token() + " " + start + " " + end + " \"" + escapeLexeme(lexeme) + "\"");
        }

        // The offset the scanner reports after it ran out of input, not source.length. The two agree only when the
        // scanner consumed everything.
        lines.add("EOF " + scanner.byteOffset());
    }

    // appendNodeTrace appends the post-order walk of the subtree, which is the shift and reduce sequence of the parse.
    private static void appendNodeTrace(List<String> lines, Parser.ParseNode node) {
        for (Parser.ParseNode child : node.children()) {
            appendNodeTrace(lines, child);
        }

        if (node.symbol() instanceof Parser.NonterminalSymbol symbol) {
            lines.add("REDUCE " + symbol.nonterminal() + " " + node.children().size());
            return;
        }

        // Everything which is no nonterminal is a terminal, so the outcome is known here.
        Parser.TerminalSymbol symbol = (Parser.TerminalSymbol) node.symbol();
        if (symbol.token() == Scanner.Token.ERROR_TOKEN) {
            // The leaf the recovery pushed where it resumed. It stands for the dropped input and names no token.
            lines.add("RESYNC");
            return;
        }
        lines.add("SHIFT " + symbol.token());
    }

    // appendParserTrace parses the whole input and appends its errors, the walk of the tree and the accept event.
    //
    // The errors come first because a shift carries no offset to interleave them by. A parse which was given up
    // returns no tree, so its trace is the errors alone and the missing accept event is what says the two outcomes
    // apart.
    private static void appendParserTrace(List<String> lines, byte[] source, String inputPath) {
        // The TokenSkipper here, because a skipped rule never reaches the parser.
        Parser.ParseResult result = new Parser().parse(new Scanner.TokenSkipper(new Scanner(source, inputPath)));

        for (Parser.ParseError error : result.errors()) {
            lines.add("ERROR " + error.byteOffset());
        }

        if (result.tree() == null) {
            return;
        }
        appendNodeTrace(lines, result.tree());
        lines.add("ACCEPT");
    }

    // writeTrace produces one trace and writes it to its file. Whatever was produced before a failure is written all
    // the same, so a runner which breaks half way still says how far it got, and the other trace is still produced.
    private static void writeTrace(String fileName, TraceProducer producer) throws IOException {
        List<String> lines = new ArrayList<>();
        try {
            producer.produce(lines);
        } catch (Throwable error) {
            // A Throwable and not an Exception, so that a stack overflow on a deep tree leaves the trace it got to.
            System.err.println("producing " + fileName + " failed:");
            error.printStackTrace();
        }

        // Every line is terminated, and an empty trace is an empty file rather than a bare newline. The line ending is
        // spelled out, because the trace is LF whatever the platform would use.
        StringBuilder text = new StringBuilder();
        for (String line : lines) {
            text.append(line).append('\n');
        }
        Files.writeString(Path.of(fileName), text.toString());
    }

    // TraceProducer is one half of the run, so that both halves are written the same way.
    @FunctionalInterface
    private interface TraceProducer {
        void produce(List<String> lines);
    }
}
