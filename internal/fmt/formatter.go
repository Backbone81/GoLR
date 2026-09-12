package fmt

import (
	"bytes"

	golrparser "github.com/backbone81/golr/internal/parsergen/frontend/golr/parser"
)

type Formatter struct {
	config Config

	scanner     *golrparser.Scanner
	indentLevel int

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
			if f.currentContext() == golrparser.TokenScanner {
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
			if f.explicitBlankLine && f.currentContext() == golrparser.TokenScanner {
				// As the user did provide explicit blank lines to separate scanner rules,
				// we emit a linebreak as well.
				f.linebreak()
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
	return f.output.Bytes()
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
