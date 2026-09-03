package java

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

//go:embed parser.java.template
var parserTemplate string

var parsedTemplate = template.Must(template.New("parser.java.template").Funcs(template.FuncMap{
	"nonterminalName":     nonterminalName,
	"isAcceptNonterminal": isAcceptNonterminal,
	"productionName":      productionName,
}).Parse(parserTemplate))

// DefaultPackageName is the Java package the generated parser is declared in when the caller names none.
const DefaultPackageName = "parser"

type Config struct {
	// PackageName is the Java package the generated parser is declared in.
	//
	// It has to be the one the scanner was generated into. The parser needs the token constants the scanner declares
	// and the scanner interface it reads them through, and the interface is not public, because a Java file holds only
	// one public type.
	PackageName string
}

type TemplateContext struct {
	Config Config
	Parser backend.Parser
	Tables Tables
}

// FromParser writes the parser as Java source code to the given writer. Returns an error if the Java source code can
// not be encoded successfully.
func FromParser(writer io.Writer, parser backend.Parser, config Config) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Backends: Java: FromParser").End()

	if len(parser.States) == 0 {
		return errors.New("the parser does not have any state")
	}

	if config.PackageName == "" {
		config.PackageName = DefaultPackageName
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

// ParserToFile writes the parser as Java source code to the given file path. Returns an error if the file can not be
// written or the Java source code can not be encoded successfully.
func ParserToFile(filePath string, parser backend.Parser, config Config) (err error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the Java file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return FromParser(file, parser, config)
}

// ParserToString returns the parser as Java source code. Returns an error if the Java source code can not be encoded
// successfully.
func ParserToString(parser backend.Parser, config Config) (string, error) {
	var builder strings.Builder
	if err := FromParser(&builder, parser, config); err != nil {
		return "", err
	}
	return builder.String(), nil
}

// terminalName returns the name of the token constant which stands for the given terminal. The constants themselves
// are written by the scanner backend, so this has to name them the same way the scanner does, see its tokenName
// helper.
func terminalName(symbol frontend.Symbol) string {
	if symbol.Name == "$end" {
		return "END_TOKEN"
	}
	if symbol.Name == frontend.SymbolError.Name {
		// The error symbol reaches the generated scanner as a rule which matches nothing, where it is named the same
		// way.
		return "ERROR_TOKEN"
	}
	return utils.JavaConstantName("Token" + utils.GoIdentifier(symbol.Name))
}

// nonterminalName returns the name of the constant which stands for the given nonterminal.
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
	return utils.JavaConstantName("Nonterminal" + utils.GoIdentifier(symbol.Name))
}

// isAcceptNonterminal reports whether the given nonterminal is the augmented start symbol, which the generator owns
// rather than the grammar. It is the one nonterminal the generated parser documents.
func isAcceptNonterminal(symbol frontend.Symbol) bool {
	return symbol.Name == "$accept"
}

// productionName returns the name of the constant which stands for the production of the given name.
func productionName(name string) string {
	return utils.JavaConstantName("Production" + utils.GoIdentifier(name))
}
