// The Go runner for the backend test corpus. It reads an input file, runs the generated scanner and the generated
// parser over it, and writes the canonical scanner trace, parser trace and tree trace the harness diffs against.
//
// This file has no dependencies, and it must not grow any. It is built by the official golang image with no network,
// which is what proves that generated GoLR code needs nothing but the bare language.
//
// Nothing is caught here. Whatever goes wrong ends the run with its stack trace, which is what a panic below is for:
// this is a test, so a run which went wrong has nothing worth salvaging, and the harness reports the case by the trace
// it is missing. Every trace is written before the next one is produced, so a parse which panics still leaves the
// traces in front of it behind and a case which fails all three stays distinguishable from one which fails only the
// last.
package main

import (
	"fmt"
	"os"
	"strings"

	"runner/parser"
)

// The names the harness reads the three traces back from. They are fixed by the harness and spelled out in every
// runner.
const (
	scannerTraceFileName = "scanner.actual"
	parserTraceFileName  = "parser.actual"
	treeTraceFileName    = "tree.actual"
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

// appendScannerTrace scans the whole input and appends one line per event: the position the token or the failed match
// starts at, a keyword, and for a token its rule and lexeme, for a failed match the bytes it could not match.
func appendScannerTrace(lines []string, source []byte, inputPath string) []string {
	// The plain Scanner and not the TokenSkipper: a skipped rule matched like any other, and the position of the
	// tokens around it is only checkable when it is in the trace.
	scanner := parser.NewScanner(source, inputPath)

	for scanner.Next() {
		position := scanner.Position(scanner.ByteOffset())
		location := fmt.Sprintf("%d:%d", position.Line, position.Column)
		lexeme := escapeLexeme(scanner.Lexeme())

		if scanner.Token() == parser.InvalidToken {
			lines = append(lines, fmt.Sprintf(`%-7s %-7s "%s"`, location, "ERROR", lexeme))
			continue
		}
		lines = append(lines, fmt.Sprintf(`%-7s %-7s %s "%s"`, location, "TOKEN", scanner.Token(), lexeme))
	}

	// The position after the scanner ran out of input, which is one past the last byte only when it consumed
	// everything. It is the offset every off by one in a line table lands on.
	position := scanner.Position(scanner.ByteOffset())
	return append(lines, fmt.Sprintf("%-7s %s", fmt.Sprintf("%d:%d", position.Line, position.Column), "EOF"))
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

// appendTreeTrace parses the whole input and appends one line per node of the tree the parse built, in pre-order. A
// parse which was given up builds no tree and appends nothing, which is the empty trace the harness expects for it.
func appendTreeTrace(lines []string, source []byte, inputPath string) []string {
	// The scanner stays at hand after the parse, because a node carries the span of the source it covers and not the
	// source itself, so the trace resolves every node through Position and Text.
	scanner := parser.NewTokenSkipper(parser.NewScanner(source, inputPath))

	root, _ := parser.NewParser().Parse(scanner)

	// The root of a tree is the start symbol, so a root which is no nonterminal is the zero node Parse returns when
	// the recovery gave up in the middle of the input.
	if _, ok := root.Symbol.Nonterminal(); !ok {
		return lines
	}
	return appendTreeNode(lines, scanner, &root, 0)
}

// appendTreeNode appends the line of the given node and the lines of everything below it, which is the pre-order the
// tree trace is read in: a node, then what it was built from. The payload carries the indentation and the position and
// span columns do not, so they stay in the same place however deep a node sits.
func appendTreeNode(lines []string, scanner *parser.TokenSkipper, node *parser.Node, depth int) []string {
	position := scanner.Position(node.ByteOffset)
	location := fmt.Sprintf("%d:%d", position.Line, position.Column)
	span := fmt.Sprintf("%d+%d", node.ByteOffset, node.ByteLength)

	payload := strings.Repeat("  ", depth)
	switch terminal, isTerminal := node.Symbol.Terminal(); {
	case isTerminal && terminal == parser.ErrorToken:
		// The error node stands for no token of its own, so its span is all it carries.
		payload += terminalTraceName(terminal)
	case isTerminal:
		// The text is read off the source through the span and never carried along from the token, which is what
		// makes the trace state that the span is right.
		payload += terminalTraceName(terminal) + ` "` + escapeLexeme(scanner.Text(node.ByteOffset, node.ByteLength)) + `"`
	default:
		nonterminal, _ := node.Symbol.Nonterminal()
		payload += reduceTracePayload(nonterminal, node.Children)
	}

	lines = append(lines, fmt.Sprintf("%-7s %-7s %s", location, span, payload))
	for i := range node.Children {
		lines = appendTreeNode(lines, scanner, &node.Children[i], depth+1)
	}
	return lines
}

// reduceTracePayload renders a node as "lhs => rhs", the way the REDUCE line of a parser trace names the production it
// was reduced from, or as "lhs => ε" for a production with an empty right hand side.
func reduceTracePayload(lhs parser.Nonterminal, rhs []parser.Node) string {
	payload := lhs.String() + " =>"
	if len(rhs) == 0 {
		return payload + " ε"
	}
	for i := range rhs {
		payload += " " + symbolTraceName(rhs[i].Symbol)
	}
	return payload
}

// symbolTraceName is the bare grammar name of a symbol, which for a nonterminal is what String returns and for a
// terminal is the name the traces spell it with.
func symbolTraceName(symbol parser.Symbol) string {
	if nonterminal, ok := symbol.Nonterminal(); ok {
		return nonterminal.String()
	}
	terminal, _ := symbol.Terminal()
	return terminalTraceName(terminal)
}

// terminalTraceName names a terminal for a trace line, giving the three tokens the grammar cannot spell a dollar name.
func terminalTraceName(terminal parser.Token) string {
	switch terminal {
	case parser.EndToken:
		return "$end"
	case parser.ErrorToken:
		return "$error"
	case parser.InvalidToken:
		return "$invalid"
	default:
		return terminal.String()
	}
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
	writeTrace(treeTraceFileName, appendTreeTrace(nil, source, inputPath))
}
