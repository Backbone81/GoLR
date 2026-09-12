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

	// emitTight reports if the next emit should not write a whitespace to separate the previous token from the next
	// one.
	emitTight bool

	// context is a stack which describes the current nesting
	context []golrparser.Token

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
	f.emitTight = false
	f.context = f.context[:0]

	f.output = bytes.NewBuffer(make([]byte, 0, len(source)))
	f.scanner = golrparser.NewScanner(source, filePath)
	for f.scanner.Next() {
		//nolint:exhaustive // We are only interested in a few special tokens
		switch f.scanner.Token() {
		case golrparser.TokenWhitespace:
		case golrparser.TokenComment:
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
			f.linebreak()
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
	}
	return f.output.Bytes()
}

func (f *Formatter) indentInc() {
	f.indentLevel++
}

func (f *Formatter) indentDec() {
	f.indentLevel = max(f.indentLevel-1, 0)
}

func (f *Formatter) linebreak() {
	f.output.Write([]byte("\n"))
	f.indentNext = true
}

func (f *Formatter) emit(data []byte) {
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
