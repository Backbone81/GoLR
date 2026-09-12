package fmt

import (
	"bytes"

	golrparser "github.com/backbone81/golr/internal/parsergen/frontend/golr/parser"
)

type Formatter struct {
	config Config

	scanner     *golrparser.Scanner
	indentLevel int

	// lastScannerIdentifierLen is the length of the most recently emitted identifier inside a @scanner section,
	// i.e. the left hand side of the scanner rule whose ":" comes next.
	lastScannerIdentifierLen int

	// scannerRuleGroups collects scanner rule marks seen during the first pass, grouped by contiguous runs of
	// rules not separated by an explicit blank line. It always has at least one (possibly empty) group; a new
	// one is started as soon as an explicit blank line is seen inside a @scanner section, and it stays empty if
	// no rule follows before the next one. Consumed by alignScannerRules afterwards.
	scannerRuleGroups [][]scannerRuleRHS

	// indentNext reports if the next output needs to be indented or not.
	indentNext bool

	// pendingLinebreak reports if a linebreak was requested but not yet written to the output. Requesting a
	// linebreak is lazy so that a same-line trailing comment can still be emitted before it.
	pendingLinebreak bool

	// emitTight reports if the next emit should not write a whitespace to separate the previous token from the next
	// one.
	emitTight bool

	// context is a stack which describes the current nesting
	context []golrparser.Token

	// explicitBlankLine reports if the whitespace just consumed contained a blank line, i.e. the user separated the
	// surrounding tokens by an empty line of their own.
	explicitBlankLine bool

	// explicitNewline reports if the whitespace just consumed contained a linebreak, i.e. the next token starts on a
	// new source line instead of trailing the previous token.
	explicitNewline bool

	output *bytes.Buffer
}

// scannerRuleRHS notes down where a scanner rule's right hand side starts, so a second pass can align it with
// the other rules in the same group.
type scannerRuleRHS struct {
	// offset is the byte position in the output right after the rule's ":", where padding is inserted.
	offset int

	// ruleNameLen is the length of the rule's identifier.
	ruleNameLen int
}

func NewFormatter() *Formatter {
	return &Formatter{
		config: DefaultConfig,
	}
}

func (f *Formatter) Format(source []byte, filePath string) []byte {
	f.indentLevel = 0
	f.indentNext = true
	f.pendingLinebreak = false
	f.emitTight = false
	f.context = f.context[:0]
	f.explicitBlankLine = false
	f.explicitNewline = false
	f.lastScannerIdentifierLen = 0
	f.scannerRuleGroups = [][]scannerRuleRHS{nil}

	f.output = bytes.NewBuffer(make([]byte, 0, len(source)))
	f.scanner = golrparser.NewScanner(source, filePath)
	for f.scanner.Next() {
		//nolint:exhaustive // We are only interested in a few special tokens
		switch f.scanner.Token() {
		case golrparser.TokenWhitespace:
			// We are looking for linebreaks and explicit blank lines by the user.
			newlines := bytes.Count(f.scanner.Lexeme(), []byte("\n"))
			if newlines >= 1 {
				f.explicitNewline = true
			}
			if newlines >= 2 {
				f.explicitBlankLine = true
				if f.currentContext() == golrparser.TokenScanner {
					// Start a new alignment group. It stays empty if no rule follows before the next blank
					// line, which is harmless.
					f.scannerRuleGroups = append(f.scannerRuleGroups, nil)
				}
			}
			continue

		case golrparser.TokenComment:
			isLineComment := bytes.HasPrefix(f.scanner.Lexeme(), []byte("//"))
			switch {
			case f.explicitNewline:
				// The comment starts on its own line rather than trailing the previous token.
				f.pendingLinebreak = false
				f.linebreak()
				f.emit(f.scanner.Lexeme())
			case f.pendingLinebreak:
				// The comment trails the previous token, but a linebreak is already scheduled to run after that
				// token. Emit the comment before that linebreak instead of flushing it early.
				f.pendingLinebreak = false
				f.emit(f.scanner.Lexeme())
				f.pendingLinebreak = true
			default:
				f.emit(f.scanner.Lexeme())
			}
			if isLineComment {
				// Anything after "//" on the same line would otherwise be swallowed into the comment.
				f.linebreak()
			}

		case golrparser.TokenColon:
			switch f.currentContext() {
			case golrparser.TokenScanner:
				// Note down where the rule's right hand side starts, so the second pass can align it.
				mark := scannerRuleRHS{
					offset:      f.output.Len() + 1,
					ruleNameLen: f.lastScannerIdentifierLen,
				}
				lastGroup := len(f.scannerRuleGroups) - 1
				f.scannerRuleGroups[lastGroup] = append(f.scannerRuleGroups[lastGroup], mark)

				f.emitTight = true
				f.emit(f.scanner.Lexeme())
			case golrparser.TokenParser:
				f.indentInc()
				f.linebreak()
				f.emit(f.scanner.Lexeme())
			default:
				f.emitTight = true
				f.emit(f.scanner.Lexeme())
			}

		case golrparser.TokenPipe:
			if !f.pendingLinebreak {
				f.linebreak()
			}
			f.emit(f.scanner.Lexeme())

		case golrparser.TokenSemi:
			switch f.currentContext() {
			case golrparser.TokenParser:
				f.linebreak()
				f.emit(f.scanner.Lexeme())
				f.indentDec()
				f.linebreak()
				f.linebreak()
			default:
				f.emitTight = true
				f.emit(f.scanner.Lexeme())
				f.linebreak()
			}

		case golrparser.TokenLbrace:
			f.emit(f.scanner.Lexeme())
			f.linebreak()
			f.indentInc()

		case golrparser.TokenRbrace:
			f.indentLevel = max(f.indentLevel-1, 0)
			if !f.indentNext {
				f.linebreak()
			}
			f.emit(f.scanner.Lexeme())
			f.linebreak()
			switch f.currentContext() {
			case golrparser.TokenScanner:
				// We need an additional linebreak between @scanner and @parser section.
				f.linebreak()
			}

			// We remove any context we did push onto the context stack.
			f.popContext()

		case golrparser.TokenLparen:
			// Left parenthesis always sit close to the previous token.
			f.emitTight = true
			f.emit(f.scanner.Lexeme())

			// The following token also needs to sit close to the left parenthesis.
			f.emitTight = true

		case golrparser.TokenRparen:
			// Right parenthesis always sit close to the previous token.
			f.emitTight = true
			f.emit(f.scanner.Lexeme())

		case golrparser.TokenIdentifier:
			switch f.currentContext() {
			case golrparser.TokenScanner:
				// This identifier is the left hand side of a scanner rule; remember its length for the ":"
				// that follows right after.
				f.lastScannerIdentifierLen = len(f.scanner.Lexeme())
				if f.explicitBlankLine {
					// As the user did provide explicit blank lines to separate scanner rules,
					// we emit a linebreak as well.
					f.linebreak()
				}
			}
			f.emit(f.scanner.Lexeme())

		case golrparser.TokenScanner:
			// When we see a @scanner token, we note it down in our context. Expecting a { } block next.
			// On the next } we remove the @scanner token again from our context.
			f.emit(f.scanner.Lexeme())
			f.pushContext(golrparser.TokenScanner)

		case golrparser.TokenParser:
			// When we see a @parser token, we note it down in our context. Expecting a { } block next.
			// On the next } we remove the @parser token again from our context.
			f.emit(f.scanner.Lexeme())
			f.pushContext(golrparser.TokenParser)

		default:
			f.emit(f.scanner.Lexeme())
		}

		f.explicitBlankLine = false
		f.explicitNewline = false
	}
	f.linebreak()
	return f.alignScannerRules(f.output.Bytes())
}

func (f *Formatter) indentInc() {
	f.indentLevel++
}

func (f *Formatter) indentDec() {
	f.indentLevel = max(f.indentLevel-1, 0)
}

// linebreak requests a linebreak before the next emitted content. It is lazy: the actual "\n" is written by emit,
// so that a same-line trailing comment can still be inserted before it. If a linebreak is already pending, this one
// is written immediately, keeping the pending flag around for the one emit will still owe.
func (f *Formatter) linebreak() {
	if f.pendingLinebreak {
		f.output.Write([]byte("\n"))
	}
	f.pendingLinebreak = true
	f.indentNext = true
}

func (f *Formatter) emit(data []byte) {
	if f.pendingLinebreak {
		f.output.Write([]byte("\n"))
		f.pendingLinebreak = false
	}
	if f.indentNext {
		for range f.indentLevel {
			f.output.Write([]byte(f.config.Indentation))
		}
		f.emitTight = true
	}
	if !f.emitTight {
		f.output.Write([]byte(" "))
	}
	f.indentNext = false
	f.emitTight = false
	f.output.Write(data)
}

func (f *Formatter) pushContext(token golrparser.Token) {
	f.context = append(f.context, token)
}

func (f *Formatter) popContext() {
	f.context = f.context[:max(len(f.context)-1, 0)]
}

func (f *Formatter) currentContext() golrparser.Token {
	if len(f.context) == 0 {
		return golrparser.InvalidToken
	}
	return f.context[len(f.context)-1]
}

// alignScannerRules is the second pass over the pretty printed output. It pads the ":" of scanner rules within
// each group (a run of rules not separated by an explicit blank line) so their right hand sides start in the same
// column.
func (f *Formatter) alignScannerRules(data []byte) []byte {
	result := make([]byte, 0, len(data))
	lastOffset := 0
	for _, group := range f.scannerRuleGroups {
		maxPrefixLen := 0
		for _, mark := range group {
			maxPrefixLen = max(maxPrefixLen, mark.ruleNameLen)
		}

		for _, mark := range group {
			result = append(result, data[lastOffset:mark.offset]...)
			for range maxPrefixLen - mark.ruleNameLen {
				result = append(result, ' ')
			}
			lastOffset = mark.offset
		}
	}
	result = append(result, data[lastOffset:]...)
	return result
}
