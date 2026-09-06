// The Go runner for the backend test corpus. It reads an input file, runs the generated scanner and the generated
// parser over it, and writes the canonical scanner trace and parser trace the harness diffs against.
//
// This file has no dependencies, and it must not grow any. It is built by the official golang image with no network,
// which is what proves that generated GoLR code needs nothing but the bare language.
//
// Nothing is caught here. Whatever goes wrong ends the run with its stack trace, which is what a panic below is for:
// this is a test, so a run which went wrong has nothing worth salvaging, and the harness reports the case by the trace
// it is missing. The scanner trace is written before the parser runs, so a parser which panics still leaves the
// scanner trace behind and a case which fails both stays distinguishable from one which fails only the parser.
package main

import (
	"fmt"
	"os"
	"strings"

	"runner/parser"
)

// The names the harness reads the two traces back from. They are fixed by the harness and spelled out in every runner.
const (
	scannerTraceFileName = "scanner.actual"
	parserTraceFileName  = "parser.actual"
)

// printableLow and printableHigh delimit the range of bytes a trace line carries as they are. Everything outside of it
// is escaped.
const (
	printableLow  = 0x20
	printableHigh = 0x7E
)

// escapeLexeme escapes the bytes of a lexeme. The caller writes the quotes around the result.
//
// The lexeme is escaped byte by byte and never rune by rune, so a multi byte UTF-8 sequence becomes one \xHH escape
// per byte. Go's %q is no alternative: it writes such a sequence as the character it stands for, and it reaches for
// escapes which are not in the trace format.
func escapeLexeme(lexeme []byte) string {
	var result strings.Builder
	result.Grow(len(lexeme))

	for _, value := range lexeme {
		switch {
		case value == '\\':
			result.WriteString(`\\`)
		case value == '"':
			result.WriteString(`\"`)
		case value == '\n':
			result.WriteString(`\n`)
		case value == '\r':
			result.WriteString(`\r`)
		case value == '\t':
			result.WriteString(`\t`)
		case printableLow <= value && value <= printableHigh:
			result.WriteByte(value)
		default:
			// %02x is lower case and zero padded, which is what the escape sequence asks for. A strings.Builder never
			// fails.
			_, _ = fmt.Fprintf(&result, `\x%02x`, value)
		}
	}

	return result.String()
}

// appendScannerTrace scans the whole input and appends one line per event.
func appendScannerTrace(lines []string, source []byte, inputPath string) []string {
	// The plain Scanner and not the TokenSkipper: a skipped rule matched like any other, and the offsets of the tokens
	// around it are only checkable when it is in the trace.
	scanner := parser.NewScanner(source, inputPath)

	for scanner.Next() {
		// For a failed match this is the start of the attempt, not the byte which could not be consumed.
		start := scanner.ByteOffset()

		if scanner.Token() == parser.InvalidToken {
			lines = append(lines, fmt.Sprintf("ERROR %d", start))
			continue
		}

		// The end comes from the length of the lexeme rather than from an offset the scanner reports, so that a lexeme
		// which disagrees with the offsets cannot pass unnoticed.
		lexeme := scanner.Lexeme()
		lines = append(lines, fmt.Sprintf(
			`TOKEN %s %d %d "%s"`,
			scanner.Token(),
			start,
			start+len(lexeme),
			escapeLexeme(lexeme),
		))
	}

	// The offset the scanner reports after it ran out of input, not the length of the source. The two agree only when
	// the scanner consumed everything.
	return append(lines, fmt.Sprintf("EOF %d", scanner.ByteOffset()))
}

// appendParserTrace parses the whole input and appends the line the parser's trace hook emits for every action. The
// hook reports the error recovery steps too, which the returned tree does not, so the tree and the error are ignored.
func appendParserTrace(lines []string, source []byte, inputPath string) []string {
	p := parser.NewParser()
	p.Trace = func(line string) {
		lines = append(lines, line)
	}

	// The TokenSkipper here, because a skipped rule never reaches the parser.
	p.Parse(parser.NewTokenSkipper(parser.NewScanner(source, inputPath)))

	return lines
}

// writeTrace writes one trace to its file. Every line is terminated, and an empty trace is an empty file rather than a
// bare newline.
func writeTrace(fileName string, lines []string) {
	var content strings.Builder
	for _, line := range lines {
		content.WriteString(line)
		content.WriteString("\n")
	}

	if err := os.WriteFile(fileName, []byte(content.String()), 0o644); err != nil {
		panic(err)
	}
}

// main runs both traces over the input file named on the command line.
func main() {
	inputPath := os.Args[1]

	// The bytes of the input and not text, because the generated scanner works on bytes and the offsets in a trace
	// count bytes.
	source, err := os.ReadFile(inputPath)
	if err != nil {
		panic(err)
	}

	writeTrace(scannerTraceFileName, appendScannerTrace(nil, source, inputPath))
	writeTrace(parserTraceFileName, appendParserTrace(nil, source, inputPath))
}
