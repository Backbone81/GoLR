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
	"errors"
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

// appendNodeTrace appends the post-order walk of the subtree, which is the shift and reduce sequence of the parse.
//
// The symbol is taken apart rather than printed, because the trace names a symbol and Symbol.String does not: it is the
// number the parser stores which comes out of it, and the name is what a golden holds.
func appendNodeTrace(lines []string, node *parser.Node) []string {
	for childIdx := range node.Children {
		lines = appendNodeTrace(lines, &node.Children[childIdx])
	}

	if nonterminal, ok := node.Symbol.Nonterminal(); ok {
		return append(lines, fmt.Sprintf("REDUCE %s %d", nonterminal, len(node.Children)))
	}

	token, _ := node.Symbol.Terminal()
	if token == parser.ErrorToken {
		// The leaf the recovery pushed where it resumed. It stands for the dropped input and names no token.
		return append(lines, "RESYNC")
	}
	return append(lines, fmt.Sprintf("SHIFT %s", token))
}

// parseErrors takes apart the error Parse returns and hands back the syntax errors it carries.
//
// Parse joins its errors even when there is only one, so a join is what a parse with any error at all returns and nil
// is what it returns without one. Anything else is a defect of the generated parser: every error a parse reports is
// supposed to name the position it happened at, and the trace has no line to describe one which does not.
func parseErrors(err error) []*parser.Error {
	if err == nil {
		return nil
	}

	joined, ok := err.(interface{ Unwrap() []error })
	if !ok {
		panic(fmt.Sprintf("the parser returned an error which is not a join: %v", err))
	}

	var result []*parser.Error
	for _, single := range joined.Unwrap() {
		var parseError *parser.Error
		if !errors.As(single, &parseError) {
			panic(fmt.Sprintf("the parser returned an error without a position: %v", single))
		}
		result = append(result, parseError)
	}
	return result
}

// appendParserTrace parses the whole input and appends its errors, the walk of the tree and the accept event.
//
// The errors come first because a shift carries no offset to interleave them by. A parse which was given up returns no
// tree, so its trace is the errors alone and the missing accept event is what says the two outcomes apart.
func appendParserTrace(lines []string, source []byte, inputPath string) []string {
	// The TokenSkipper here, because a skipped rule never reaches the parser.
	tree, err := parser.NewParser().Parse(parser.NewTokenSkipper(parser.NewScanner(source, inputPath)))

	for _, parseError := range parseErrors(err) {
		lines = append(lines, fmt.Sprintf("ERROR %d", parseError.ByteOffset))
	}

	// The root of a tree is the start symbol and therefore a nonterminal. A parse which was given up returns the zero
	// Node instead, whose symbol is a terminal, which is how the two are told apart without a second return value.
	if _, ok := tree.Symbol.Nonterminal(); !ok {
		return lines
	}

	lines = appendNodeTrace(lines, &tree)
	return append(lines, "ACCEPT")
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
