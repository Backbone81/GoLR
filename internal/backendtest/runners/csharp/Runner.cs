// The C# runner for the backend test corpus. It reads an input file, runs the generated scanner and the generated
// parser over it, and writes the canonical scanner trace and parser trace the harness diffs against.
//
// This file has no dependencies, and it must not grow any. It runs in the image with no network, which is what proves
// that generated GoLR code needs nothing but the bare language.

using System;
using System.Collections.Generic;
using System.Globalization;
using System.IO;
using System.Text;

using Parser;

internal static class Runner
{
    private const string ScannerTraceFileName = "scanner.actual";
    private const string ParserTraceFileName = "parser.actual";

    // The bytes a trace line carries as they are. Everything outside of it is escaped.
    private const byte PrintableLow = 0x20;
    private const byte PrintableHigh = 0x7e;

    private static void Main(string[] args)
    {
        if (args.Length != 1)
        {
            throw new ArgumentException("the runner takes the input file as its only argument");
        }
        string inputPath = args[0];

        // ReadAllBytes and not ReadAllText, because the generated scanner wants the bytes. Handing it a decoded string
        // would be the classic mistake this harness exists to catch.
        byte[] source = File.ReadAllBytes(inputPath);

        WriteTrace(ScannerTraceFileName, lines => AppendScannerTrace(lines, source, inputPath));
        WriteTrace(ParserTraceFileName, lines => AppendParserTrace(lines, source, inputPath));
    }

    // EscapeLexeme escapes the bytes of a lexeme. The caller writes the quotes around the result.
    //
    // The lexeme is escaped byte by byte, never rune by rune, so a multi byte UTF-8 sequence becomes one \xHH escape
    // per byte. Decoding it first would report UTF-16 code unit offsets and disagree with every other backend.
    private static string EscapeLexeme(ReadOnlySpan<byte> lexeme)
    {
        StringBuilder result = new StringBuilder();
        foreach (byte value in lexeme)
        {
            switch (value)
            {
                case 0x5c: // backslash
                    result.Append("\\\\");
                    continue;
                case 0x22: // quote
                    result.Append("\\\"");
                    continue;
                case 0x0a:
                    result.Append("\\n");
                    continue;
                case 0x0d:
                    result.Append("\\r");
                    continue;
                case 0x09:
                    result.Append("\\t");
                    continue;
            }

            if (PrintableLow <= value && value <= PrintableHigh)
            {
                result.Append((char)value);
                continue;
            }

            // The trace asks for a zero padded pair of lower case hex digits.
            result.Append("\\x").Append(value.ToString("x2", CultureInfo.InvariantCulture));
        }
        return result.ToString();
    }

    // AppendScannerTrace scans the whole input and appends one line per event.
    private static void AppendScannerTrace(List<string> lines, byte[] source, string inputPath)
    {
        // The plain Scanner and not the TokenSkipper: a skipped rule matched like any other, and the offsets of the
        // tokens around it are only checkable when it is in the trace.
        Scanner scanner = new Scanner(source, inputPath);

        while (scanner.Next())
        {
            // For a failed match this is the start of the attempt, not the byte which could not be consumed.
            int start = scanner.ByteOffset;

            if (scanner.Token == Token.InvalidToken)
            {
                lines.Add($"ERROR {start}");
                continue;
            }

            // The end comes from the length of the lexeme rather than from an offset the scanner reports, so that a
            // lexeme which disagrees with the offsets cannot pass unnoticed.
            ReadOnlySpan<byte> lexeme = scanner.Lexeme.Span;
            string name = scanner.Token.ToDisplayString();
            lines.Add($"TOKEN {name} {start} {start + lexeme.Length} \"{EscapeLexeme(lexeme)}\"");
        }

        // The offset the scanner reports after it ran out of input, not source.Length. The two agree only when the
        // scanner consumed everything.
        lines.Add($"EOF {scanner.ByteOffset}");
    }

    // AppendNodeTrace appends the post-order walk of the subtree, which is the shift and reduce sequence of the parse.
    private static void AppendNodeTrace(List<string> lines, ParseNode node)
    {
        foreach (ParseNode child in node.Children)
        {
            AppendNodeTrace(lines, child);
        }

        if (node.Symbol.TryGetNonterminal(out Nonterminal nonterminal))
        {
            lines.Add($"REDUCE {nonterminal.ToDisplayString()} {node.Children.Count}");
            return;
        }

        // Everything which is no nonterminal is a terminal, so the outcome is known here.
        _ = node.Symbol.TryGetTerminal(out Token token);
        if (token == Token.ErrorToken)
        {
            // The leaf the recovery pushed where it resumed. It stands for the dropped input and names no token.
            lines.Add("RESYNC");
            return;
        }
        lines.Add($"SHIFT {token.ToDisplayString()}");
    }

    // AppendParserTrace parses the whole input and appends its errors, the walk of the tree and the accept event.
    //
    // The errors come first because a shift carries no offset to interleave them by. A parse which was given up returns
    // no tree, so its trace is the errors alone and the missing accept event is what says the two outcomes apart.
    private static void AppendParserTrace(List<string> lines, byte[] source, string inputPath)
    {
        // The TokenSkipper here, because a skipped rule never reaches the parser.
        ParseResult result = new Parser.Parser().Parse(new TokenSkipper(new Scanner(source, inputPath)));

        foreach (ParseError error in result.Errors)
        {
            lines.Add($"ERROR {error.ByteOffset}");
        }

        if (result.Tree == null)
        {
            return;
        }
        AppendNodeTrace(lines, result.Tree);
        lines.Add("ACCEPT");
    }

    // WriteTrace produces one trace and writes it to its file. Whatever was produced before a throw is written all the
    // same, so a runner which breaks half way still says how far it got, and the other trace is still produced.
    private static void WriteTrace(string fileName, Action<List<string>> produce)
    {
        List<string> lines = new List<string>();
        try
        {
            produce(lines);
        }
        catch (Exception error)
        {
            Console.Error.WriteLine($"producing {fileName} failed: {error}");
        }

        // Every line is terminated, and an empty trace is an empty file rather than a bare newline. The line ending is
        // spelled out, because the trace is LF whatever the platform would use.
        StringBuilder text = new StringBuilder();
        foreach (string line in lines)
        {
            text.Append(line).Append('\n');
        }
        File.WriteAllText(fileName, text.ToString());
    }
}
