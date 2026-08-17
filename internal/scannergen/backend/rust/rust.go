package rust

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

//go:embed scanner.rs.template
var scannerTemplate string

var parsedTemplate = template.Must(template.New("scanner.rs.template").Funcs(template.FuncMap{
	"tokenName": tokenName,
}).Parse(scannerTemplate))

type TemplateContext struct {
	Tables Tables
	DFA    backend.DFA
}

// FromDFA writes the DFA as Rust source code to the given writer. Returns an error if the Rust source code can not be
// encoded successfully.
func FromDFA(writer io.Writer, dfa backend.DFA) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Backend: Rust: FromDFA").End()

	if len(dfa.States) == 0 {
		return errors.New("the DFA does not have any state")
	}

	var buffer bytes.Buffer
	if err := parsedTemplate.Execute(&buffer, TemplateContext{
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

// DFAToFile writes the DFA as Rust source code to the given file path. Returns an error if the file can not be
// written or the Rust source code can not be encoded successfully.
func DFAToFile(filePath string, dfa backend.DFA) (err error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the Rust file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return FromDFA(file, dfa)
}

// tokenName returns the name of the token variant the generated scanner declares for the given rule. Rust spells an
// enum variant in the upper camel case the Go backends use, so a token is named the same in both.
func tokenName(ruleIdx int, rule frontend.Rule) string {
	return "Token" + utils.GoIdentifier(rule.Name)
}
