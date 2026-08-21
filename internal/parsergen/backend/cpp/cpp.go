package cpp

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/trace"
	"strings"
	"text/template"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

//go:embed parser.hpp.template
var parserTemplate string

var parsedTemplate = template.Must(template.New("parser.hpp.template").Funcs(template.FuncMap{
	"nonterminalName": nonterminalName,
}).Parse(parserTemplate))

// DefaultScannerInclude is the header the generated parser includes the token type from when the caller names none. It
// resolves to the header beside the parser, which is where the scanner sits both in the examples and in the backend
// test harness.
const DefaultScannerInclude = "scanner.hpp"

type Config struct {
	// Namespace is the C++ namespace the generated parser is declared in. It has to be the one the generated scanner
	// was given, because the parser declares the token type and the scanner interface to be members of its own
	// namespace.
	Namespace string

	// ScannerInclude is the header the generated parser includes the token type and the scanner interface from. It is
	// written into a quoted include, so it is resolved relative to the parser itself.
	//
	// The flag is not called a module, the way the JavaScript, TypeScript, Rust and Python ones are, because a module
	// is a different thing in C++ and this is a header.
	ScannerInclude string
}

type TemplateContext struct {
	Config Config
	Parser backend.Parser
	Tables Tables
}

// FromParser writes the parser as C++ source code to the given writer. Returns an error if the C++ source code can not
// be encoded successfully.
func FromParser(writer io.Writer, parser backend.Parser, config Config) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Backends: Cpp: FromParser").End()

	if len(parser.States) == 0 {
		return errors.New("the parser does not have any state")
	}
	if config.ScannerInclude == "" {
		config.ScannerInclude = DefaultScannerInclude
	}

	var buffer bytes.Buffer
	if err := parsedTemplate.Execute(&buffer, TemplateContext{
		Config: config,
		Parser: parser,
		Tables: NewTables(parser),
	}); err != nil {
		return fmt.Errorf("rendering the template: %w", err)
	}

	if _, err := writer.Write(buffer.Bytes()); err != nil {
		return err
	}
	return nil
}

// ParserToFile writes the parser as C++ source code to the given file path. Returns an error if the file can not be
// written or the C++ source code can not be encoded successfully.
func ParserToFile(filePath string, parser backend.Parser, config Config) (err error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the C++ file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return FromParser(file, parser, config)
}

// ParserToString returns the parser as C++ source code. Returns an error if the C++ source code can not be encoded
// successfully.
func ParserToString(parser backend.Parser, config Config) (string, error) {
	var builder strings.Builder
	if err := FromParser(&builder, parser, config); err != nil {
		return "", err
	}
	return builder.String(), nil
}

// terminalName returns the name of the enumerator which stands for the given terminal. The enumerators themselves are
// written by the scanner backend, so this has to name them the same way the scanner does, see its tokenName helper.
func terminalName(symbol frontend.Symbol) string {
	if symbol.Name == "$end" {
		return "EndToken"
	}
	if symbol.Name == frontend.SymbolError.Name {
		// The error symbol reaches the generated scanner as a rule which matches nothing, where it is named the same
		// way.
		return "ErrorToken"
	}
	return "Token" + utils.GoIdentifier(symbol.Name)
}

// nonterminalName returns the name of the enumerator which stands for the given nonterminal.
//
// A nonterminal of the grammar is prefixed, which is what keeps a grammar free to name a nonterminal anything without
// colliding with a nonterminal the generator needs for itself. The augmented start symbol is such a generator owned
// nonterminal, and it is suffixed instead, the way terminalName suffixes the ones the grammar cannot name either. The
// two shapes are what tells them apart: a grammar which does have a nonterminal called accept gets NonterminalAccept,
// which is not the AcceptNonterminal returned here.
func nonterminalName(symbol frontend.Symbol) string {
	if symbol.Name == "$accept" {
		return "AcceptNonterminal"
	}
	return "Nonterminal" + utils.GoIdentifier(symbol.Name)
}
