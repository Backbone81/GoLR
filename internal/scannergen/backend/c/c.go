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
	"text/template"

	"github.com/backbone81/golr/internal/scannergen/backend"
	"github.com/backbone81/golr/internal/scannergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

//go:embed scanner.h.template
var scannerTemplate string

var parsedTemplate = template.Must(template.New("scanner.h.template").Parse(scannerTemplate))

// DefaultPrefix is the prefix the generated names carry when the caller names none.
const DefaultPrefix = "parser"

type Config struct {
	// Prefix is put in front of every name the generated scanner declares.
	//
	// C has no namespaces, so the prefix is what keeps two generated scanners in one program apart. The generated
	// parser has to be given the same one, so that the token type it reads from here is the one it declares itself
	// to use.
	Prefix string
}

type TemplateContext struct {
	Config Config
	Tables Tables
	DFA    backend.DFA
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

// TokenName returns the name of the enumerator the generated scanner declares for the given rule.
func (c TemplateContext) TokenName(rule frontend.Rule) string {
	return tokenName(c.Config.Prefix, rule)
}

// FromDFA writes the DFA as C source code to the given writer. Returns an error if the C source code can not be
// encoded successfully.
func FromDFA(writer io.Writer, dfa backend.DFA, config Config) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Backend: C: FromDFA").End()

	if len(dfa.States) == 0 {
		return errors.New("the DFA does not have any state")
	}

	if config.Prefix == "" {
		config.Prefix = DefaultPrefix
	}

	var buffer bytes.Buffer
	if err := parsedTemplate.Execute(&buffer, TemplateContext{
		Config: config,
		Tables: NewTables(dfa, config),
		DFA:    dfa,
	}); err != nil {
		return fmt.Errorf("rendering the template: %w", err)
	}

	if _, err := writer.Write(buffer.Bytes()); err != nil {
		return err
	}
	return nil
}

// DFAToFile writes the DFA as C source code to the given file path. Returns an error if the file can not be written
// or the C source code can not be encoded successfully.
func DFAToFile(filePath string, dfa backend.DFA, config Config) (err error) {
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

	return FromDFA(file, dfa, config)
}

// tokenName returns the name of the enumerator the generated scanner declares for the given rule. C spells an
// enumerator in upper case with underscores, so a token is called what the Java and Python backends call it rather
// than what the Go, Rust, C# and C++ ones do.
func tokenName(prefix string, rule frontend.Rule) string {
	return utils.CConstantName(prefix, "Token"+utils.GoIdentifier(rule.Name))
}
