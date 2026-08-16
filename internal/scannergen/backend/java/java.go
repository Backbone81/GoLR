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
	"text/template"

	"github.com/backbone81/golr/internal/scannergen/backend"
	"github.com/backbone81/golr/internal/scannergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

//go:embed scanner.java.template
var scannerTemplate string

var parsedTemplate = template.Must(template.New("scanner.java.template").Funcs(template.FuncMap{
	"tokenName": tokenName,
}).Parse(scannerTemplate))

type Config struct {
	// PackageName is the Java package the generated scanner is declared in.
	//
	// The generated parser has to be given the same one. Java resolves a type by its package and the parser needs both
	// the token constants and the scanner interface declared here.
	PackageName string
}

type TemplateContext struct {
	Config Config
	Tables Tables
	DFA    backend.DFA
}

// FromDFA writes the DFA as Java source code to the given writer. Returns an error if the Java source code can not be
// encoded successfully.
func FromDFA(writer io.Writer, dfa backend.DFA, config Config) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Backend: Java: FromDFA").End()

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

// DFAToFile writes the DFA as Java source code to the given file path. Returns an error if the file can not be written
// or the Java source code can not be encoded successfully.
func DFAToFile(filePath string, dfa backend.DFA, config Config) (err error) {
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

	return FromDFA(file, dfa, config)
}

// tokenName returns the name of the token the generated scanner declares for the given rule. The name a token is known
// by follows the Go backends, so that it is the same in every language, spelled the way Java spells a constant.
func tokenName(ruleIdx int, rule frontend.Rule) string {
	return utils.JavaConstantName("Token" + utils.GoIdentifier(rule.Name))
}
