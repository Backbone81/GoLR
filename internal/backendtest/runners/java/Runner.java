// The Java runner for the backend test corpus. It reads an input file, runs the generated scanner and the generated
// parser over it, and writes the canonical scanner trace, parser trace and tree trace the harness diffs against.
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
    private static final String TREE_TRACE_FILE_NAME = "tree.actual";

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
        writeTrace(TREE_TRACE_FILE_NAME, lines -> appendTreeTrace(lines, source, inputPath));
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

    // checkPosition holds the offset based position() and text() against the line(), column() and lexeme() of the token
    // the scanner currently sits on. The two ways of asking exist side by side until the release which drops line() and
    // column(), and the corpus is where they have to agree: every case of it is far more input than a hand written test
    // covers.
    private static void checkPosition(Scanner scanner) {
        Scanner.Position position = scanner.position(scanner.byteOffset());
        if (position.line() != scanner.line() || position.column() != scanner.column()
                || !position.filePath().equals(scanner.filePath())) {
            System.err.printf(Locale.ROOT, "position(%d) is %s %d:%d, but the scanner reports %s %d:%d%n",
                    scanner.byteOffset(), position.filePath(), position.line(), position.column(), scanner.filePath(),
                    scanner.line(), scanner.column());
            System.exit(1);
        }

        ByteBuffer lexeme = scanner.lexeme();
        if (!scanner.text(scanner.byteOffset(), lexeme.remaining()).equals(lexeme)) {
            System.err.printf(Locale.ROOT, "text(%d, %d) differs from the lexeme%n", scanner.byteOffset(),
                    lexeme.remaining());
            System.exit(1);
        }
    }

    // appendScannerTrace scans the whole input and appends one line per event: the position the token or the failed
    // match starts at, a keyword, and for a token its rule and lexeme, for a failed match the bytes it could not match.
    private static void appendScannerTrace(List<String> lines, byte[] source, String inputPath) {
        // The plain Scanner and not the TokenSkipper: a skipped rule matched like any other, and the position of the
        // tokens around it is only checkable when it is in the trace.
        Scanner scanner = new Scanner(source, inputPath);

        while (scanner.next()) {
            checkPosition(scanner);

            String location = scanner.line() + ":" + scanner.column();
            String lexeme = escapeLexeme(scanner.lexeme());

            if (scanner.token() == Scanner.Token.INVALID_TOKEN) {
                lines.add(String.format(Locale.ROOT, "%-7s %-7s \"%s\"", location, "ERROR", lexeme));
                continue;
            }
            lines.add(String.format(Locale.ROOT, "%-7s %-7s %s \"%s\"", location, "TOKEN", scanner.token(), lexeme));
        }

        // The position after the scanner ran out of input, which is one past the last byte only when it consumed
        // everything. It is the offset every off by one in a line table lands on, so it is checked like a token.
        checkPosition(scanner);
        lines.add(String.format(Locale.ROOT, "%-7s %s", scanner.line() + ":" + scanner.column(), "EOF"));
    }

    // appendParserTrace parses the whole input and appends the line the parser's trace hook emits for every action.
    // The hook reports the error recovery steps too, which the returned tree does not, so the tree and the error are
    // ignored.
    private static void appendParserTrace(List<String> lines, byte[] source, String inputPath) {
        Parser parser = new Parser();
        parser.trace = lines::add;

        // The TokenSkipper here, because a skipped rule never reaches the parser.
        parser.parse(new Scanner.TokenSkipper(new Scanner(source, inputPath)));
    }

    // terminalTraceName names a terminal for a trace line, giving the three tokens the grammar cannot spell a dollar
    // name.
    private static String terminalTraceName(Scanner.Token terminal) {
        return switch (terminal) {
            case END_TOKEN -> "$end";
            case ERROR_TOKEN -> "$error";
            case INVALID_TOKEN -> "$invalid";
            default -> terminal.toString();
        };
    }

    // symbolTraceName is the bare grammar name of a symbol, which for a nonterminal is what toString returns and for a
    // terminal is the name the traces spell it with.
    private static String symbolTraceName(Parser.ParseSymbol symbol) {
        if (symbol instanceof Parser.NonterminalSymbol nonterminal) {
            return nonterminal.nonterminal().toString();
        }
        return terminalTraceName(((Parser.TerminalSymbol) symbol).token());
    }

    // reduceTracePayload renders a node as "lhs => rhs", the way the REDUCE line of a parser trace names the
    // production it was reduced from, or as "lhs => ε" for a production with an empty right hand side.
    private static String reduceTracePayload(Parser.Nonterminal lhs, List<Parser.ParseNode> rhs) {
        StringBuilder payload = new StringBuilder(lhs.toString()).append(" =>");
        if (rhs.isEmpty()) {
            return payload.append(" ε").toString();
        }
        for (Parser.ParseNode child : rhs) {
            payload.append(' ').append(symbolTraceName(child.symbol()));
        }
        return payload.toString();
    }

    // appendTreeNode appends the line of the given node and the lines of everything below it, which is the pre-order
    // the tree trace is read in: a node, then what it was built from. The payload carries the indentation and the
    // position and span columns do not, so they stay in the same place however deep a node sits.
    private static void appendTreeNode(List<String> lines, Scanner scanner, Parser.ParseNode node, int depth) {
        Scanner.Position position = scanner.position(node.byteOffset());
        String location = position.line() + ":" + position.column();
        String span = node.byteOffset() + "+" + node.byteLength();

        StringBuilder payload = new StringBuilder("  ".repeat(depth));
        if (node.symbol() instanceof Parser.NonterminalSymbol nonterminal) {
            payload.append(reduceTracePayload(nonterminal.nonterminal(), node.children()));
        } else {
            Scanner.Token terminal = ((Parser.TerminalSymbol) node.symbol()).token();
            if (terminal == Scanner.Token.ERROR_TOKEN) {
                // The error node stands for no token of its own, so its span is all it carries.
                payload.append(terminalTraceName(terminal));
            } else {
                // The text is read off the source through the span and never carried along from the token, which is
                // what makes the trace state that the span is right.
                payload.append(terminalTraceName(terminal))
                        .append(" \"")
                        .append(escapeLexeme(scanner.text(node.byteOffset(), node.byteLength())))
                        .append('"');
            }
        }

        lines.add(String.format(Locale.ROOT, "%-7s %-7s %s", location, span, payload));
        for (Parser.ParseNode child : node.children()) {
            appendTreeNode(lines, scanner, child, depth + 1);
        }
    }

    // appendTreeTrace parses the whole input and appends one line per node of the tree the parse built, in pre-order.
    // A parse which was given up builds no tree and appends nothing, which is the empty trace the harness expects for
    // it.
    private static void appendTreeTrace(List<String> lines, byte[] source, String inputPath) {
        // The scanner stays at hand after the parse, because a node carries the span of the source it covers and not
        // the source itself, so the trace resolves every node through position() and text().
        Scanner scanner = new Scanner(source, inputPath);

        Parser.ParseResult result = new Parser().parse(new Scanner.TokenSkipper(scanner));
        if (result.tree() != null) {
            appendTreeNode(lines, scanner, result.tree(), 0);
        }
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
