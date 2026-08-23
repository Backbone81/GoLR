package c

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

//go:embed parser.h.template
var parserTemplate string

var parsedTemplate = template.Must(template.New("parser.h.template").Parse(parserTemplate))

// DefaultPrefix is the prefix the generated names carry when the caller names none.
const DefaultPrefix = "parser"

// DefaultScannerInclude is the header the generated parser includes the token type from when the caller names none. It
// resolves to the header beside the parser, which is where the scanner sits both in the examples and in the backend
// test harness.
const DefaultScannerInclude = "scanner.h"

type Config struct {
	// Prefix is put in front of every name the generated parser declares. It has to be the one the generated scanner
	// was given, because the parser reads the token type from there under that prefix.
	//
	// C has no namespaces, so the prefix is what keeps two generated parsers in one program apart.
	Prefix string

	// ScannerInclude is the header the generated parser includes the token type and the scanner from. It is written
	// into a quoted include, so it is resolved relative to the parser itself.
	ScannerInclude string
}

type TemplateContext struct {
	Config Config
	Parser backend.Parser
	Tables Tables
}

// Type returns the name the given type goes by in the generated code.
func (c TemplateContext) Type(name string) string {
	return utils.CTypeName(c.Config.Prefix, name)
}

// Func returns the name the given function goes by in the generated code.
func (c TemplateContext) Func(name string) string {
	return utils.CFunctionName(c.Config.Prefix, name)
}

// Const returns the name the given constant or enumerator goes by in the generated code.
func (c TemplateContext) Const(name string) string {
	return utils.CConstantName(c.Config.Prefix, name)
}

// NonterminalName returns the name of the enumerator which stands for the given nonterminal.
func (c TemplateContext) NonterminalName(symbol frontend.Symbol) string {
	return nonterminalName(c.Config.Prefix, symbol)
}

// FromParser writes the parser as C source code to the given writer. Returns an error if the C source code can not be
// encoded successfully.
func FromParser(writer io.Writer, parser backend.Parser, config Config) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Backends: C: FromParser").End()

	if len(parser.States) == 0 {
		return errors.New("the parser does not have any state")
	}
	if config.Prefix == "" {
		config.Prefix = DefaultPrefix
	}
	if config.ScannerInclude == "" {
		config.ScannerInclude = DefaultScannerInclude
	}

	var buffer bytes.Buffer
	if err := parsedTemplate.Execute(&buffer, TemplateContext{
		Config: config,
		Parser: parser,
		Tables: NewTables(parser, config),
	}); err != nil {
		return fmt.Errorf("rendering the template: %w", err)
	}

	if _, err := writer.Write(buffer.Bytes()); err != nil {
		return err
	}
	return nil
}

// ParserToFile writes the parser as C source code to the given file path. Returns an error if the file can not be
// written or the C source code can not be encoded successfully.
func ParserToFile(filePath string, parser backend.Parser, config Config) (err error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the C file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return FromParser(file, parser, config)
}

// ParserToString returns the parser as C source code. Returns an error if the C source code can not be encoded
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
func terminalName(prefix string, symbol frontend.Symbol) string {
	if symbol.Name == "$end" {
		return utils.CConstantName(prefix, "TokenEndToken")
	}
	if symbol.Name == frontend.SymbolError.Name {
		// The error symbol reaches the generated scanner as a rule which matches nothing, where it is named the same
		// way.
		return utils.CConstantName(prefix, "TokenErrorToken")
	}
	return utils.CConstantName(prefix, "Token"+utils.GoIdentifier(symbol.Name))
}

// nonterminalName returns the name of the enumerator which stands for the given nonterminal.
//
// A nonterminal of the grammar is prefixed, which is what keeps a grammar free to name a nonterminal anything without
// colliding with a nonterminal the generator needs for itself. The augmented start symbol is such a generator owned
// nonterminal, and it is suffixed instead, the way terminalName suffixes the ones the grammar cannot name either. The
// two shapes are what tells them apart: a grammar which does have a nonterminal called accept gets
// NONTERMINAL_ACCEPT, which is not the ACCEPT_NONTERMINAL returned here.
func nonterminalName(prefix string, symbol frontend.Symbol) string {
	if symbol.Name == "$accept" {
		return utils.CConstantName(prefix, "AcceptNonterminal")
	}
	return utils.CConstantName(prefix, "Nonterminal"+utils.GoIdentifier(symbol.Name))
}
