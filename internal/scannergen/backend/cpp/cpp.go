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
	"text/template"

	"github.com/backbone81/golr/internal/scannergen/backend"
	"github.com/backbone81/golr/internal/scannergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

//go:embed scanner.hpp.template
var scannerTemplate string

var parsedTemplate = template.Must(template.New("scanner.hpp.template").Funcs(template.FuncMap{
	"tokenName": tokenName,
}).Parse(scannerTemplate))

type Config struct {
	// Namespace is the C++ namespace the generated scanner is declared in.
	//
	// The generated parser has to be given the same one, so that the token type and the scanner interface it includes
	// from here are the ones it declares itself to use.
	Namespace string
}

type TemplateContext struct {
	Config Config
	Tables Tables
	DFA    backend.DFA
}

// FromDFA writes the DFA as C++ source code to the given writer. Returns an error if the C++ source code can not be
// encoded successfully.
func FromDFA(writer io.Writer, dfa backend.DFA, config Config) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Backend: Cpp: FromDFA").End()

	if len(dfa.States) == 0 {
		return errors.New("the DFA does not have any state")
	}

	var buffer bytes.Buffer
	if err := parsedTemplate.Execute(&buffer, TemplateContext{
		Config: config,
		Tables: NewTables(dfa),
		DFA:    dfa,
	}); err != nil {
		return fmt.Errorf("rendering the template: %w", err)
	}

	if _, err := writer.Write(buffer.Bytes()); err != nil {
		return err
	}
	return nil
}

// DFAToFile writes the DFA as C++ source code to the given file path. Returns an error if the file can not be written
// or the C++ source code can not be encoded successfully.
func DFAToFile(filePath string, dfa backend.DFA, config Config) (err error) {
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

	return FromDFA(file, dfa, config)
}

// tokenName returns the name of the enumerator the generated scanner declares for the given rule. C++ has no single
// convention for the enumerators of a scoped enumeration, so the naming follows the Go backends and a token is called
// the same as in Go, Rust, C# and TypeScript.
func tokenName(ruleIdx int, rule frontend.Rule) string {
	return "Token" + utils.GoIdentifier(rule.Name)
}
