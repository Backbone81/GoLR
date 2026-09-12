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

	output *bytes.Buffer
}

func NewFormatter() *Formatter {
	return &Formatter{
		config: DefaultConfig,
	}
}

func (f *Formatter) Format(source []byte, filePath string) []byte {
	f.output = bytes.NewBuffer(make([]byte, 0, len(source)))
	f.scanner = golrparser.NewScanner(source, filePath)
	for f.scanner.Next() {
		//nolint:exhaustive // We are only interested in a few special tokens
		switch f.scanner.Token() {
		case golrparser.TokenWhitespace:
		case golrparser.TokenComment:
		case golrparser.TokenSemi:
			f.emit(f.scanner.Lexeme())
			f.linebreak()
		case golrparser.TokenColon:
			f.emitTight = true
			f.emit(f.scanner.Lexeme())
		case golrparser.TokenLbrace:
			f.emit(f.scanner.Lexeme())
			f.linebreak()
			f.indentLevel++
		case golrparser.TokenRbrace:
			f.indentLevel = max(f.indentLevel-1, 0)
			f.linebreak()
			f.emit(f.scanner.Lexeme())
			f.linebreak()
		default:
			f.emit(f.scanner.Lexeme())
		}
	}
	return f.output.Bytes()
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
		f.indentNext = false
		f.emitTight = false
	}
	if !f.emitTight {
		f.output.Write([]byte(" "))
		f.emitTight = false
	}
	f.output.Write(data)
}
