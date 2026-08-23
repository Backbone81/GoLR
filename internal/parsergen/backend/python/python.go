package python

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

//go:embed parser.py.template
var parserTemplate string

var parsedTemplate = template.Must(template.New("parser.py.template").Funcs(template.FuncMap{
	"nonterminalName":     nonterminalName,
	"isAcceptNonterminal": isAcceptNonterminal,
}).Parse(parserTemplate))

// DefaultScannerModule is the module the generated parser imports the token constants from when the caller names none.
// It is the file name the scanner backend writes to in the examples and in the backend test harness, without the
// extension.
const DefaultScannerModule = "scanner"

type Config struct {
	// ScannerModule is the module the generated parser imports the token constants from.
	//
	// The parser needs the values the scanner gave the terminals and never learns them itself. Python has a module
	// boundary between the two, so the module is the one thing this backend has to be told. A leading dot makes the
	// import relative to the package the two files sit in.
	ScannerModule string
}

type TemplateContext struct {
	Config Config
	Parser backend.Parser
	Tables Tables
}

// FromParser writes the parser as Python source code to the given writer. Returns an error if the Python source code
// can not be encoded successfully.
func FromParser(writer io.Writer, parser backend.Parser, config Config) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Backends: Python: FromParser").End()

	if len(parser.States) == 0 {
		return errors.New("the parser does not have any state")
	}

	if config.ScannerModule == "" {
		config.ScannerModule = DefaultScannerModule
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

// ParserToFile writes the parser as Python source code to the given file path. Returns an error if the file can not be
// written or the Python source code can not be encoded successfully.
func ParserToFile(filePath string, parser backend.Parser, config Config) (err error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the Python file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return FromParser(file, parser, config)
}

// ParserToString returns the parser as Python source code. Returns an error if the Python source code can not be
// encoded successfully.
func ParserToString(parser backend.Parser, config Config) (string, error) {
	var builder strings.Builder
	if err := FromParser(&builder, parser, config); err != nil {
		return "", err
	}
	return builder.String(), nil
}

// terminalName returns the name of the token member which stands for the given terminal. The members themselves are
// written by the scanner backend, so this has to name them the same way the scanner does, see its tokenName helper.
func terminalName(symbol frontend.Symbol) string {
	if symbol.Name == "$end" {
		return "END_TOKEN"
	}
	if symbol.Name == frontend.SymbolError.Name {
		// The error symbol reaches the generated scanner as a rule which matches nothing, where it is named the same
		// way.
		return "ERROR_TOKEN"
	}
	return utils.PythonConstantName("Token" + utils.GoIdentifier(symbol.Name))
}

// nonterminalName returns the name of the member which stands for the given nonterminal.
//
// A nonterminal of the grammar is prefixed, which is what keeps a grammar free to name a nonterminal anything without
// colliding with a nonterminal the generator needs for itself. The augmented start symbol is such a generator owned
// nonterminal, and it is suffixed instead, the way terminalName suffixes the ones the grammar cannot name either. The
// two shapes are what tells them apart: a grammar which does have a nonterminal called accept gets
// NONTERMINAL_ACCEPT, which is not the ACCEPT_NONTERMINAL returned here.
func nonterminalName(symbol frontend.Symbol) string {
	if isAcceptNonterminal(symbol) {
		return "ACCEPT_NONTERMINAL"
	}
	return utils.PythonConstantName("Nonterminal" + utils.GoIdentifier(symbol.Name))
}

// isAcceptNonterminal reports whether the given nonterminal is the augmented start symbol, which the generator owns
// rather than the grammar. It is the one nonterminal the generated parser documents.
func isAcceptNonterminal(symbol frontend.Symbol) bool {
	return symbol.Name == "$accept"
}
