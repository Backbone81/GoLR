package csharp

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

//go:embed scanner.cs.template
var scannerTemplate string

var parsedTemplate = template.Must(template.New("scanner.cs.template").Funcs(template.FuncMap{
	"tokenName":  tokenName,
	"tokenValue": tokenValue,
}).Parse(scannerTemplate))

// DefaultNamespace is the C# namespace the generated scanner is declared in when the caller names none.
const DefaultNamespace = "Parser"

type Config struct {
	// Namespace is the C# namespace the generated scanner is declared in.
	//
	// The generated parser has to be given the same one. C# resolves a type by its namespace and not by the file it
	// is in, so that is what lets the parser use the token constants and the scanner interface declared here without
	// naming a file at all.
	Namespace string
}

type TemplateContext struct {
	Config Config
	Tables Tables
	DFA    backend.DFA
}

// FromDFA writes the DFA as C# source code to the given writer. Returns an error if the C# source code can not be
// encoded successfully.
func FromDFA(writer io.Writer, dfa backend.DFA, config Config) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Backend: CSharp: FromDFA").End()

	if len(dfa.States) == 0 {
		return errors.New("the DFA does not have any state")
	}

	if config.Namespace == "" {
		config.Namespace = DefaultNamespace
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

// DFAToFile writes the DFA as C# source code to the given file path. Returns an error if the file can not be written
// or the C# source code can not be encoded successfully.
func DFAToFile(filePath string, dfa backend.DFA, config Config) (err error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the C# file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return FromDFA(file, dfa, config)
}

// tokenName returns the name of the token the generated scanner declares for the given rule. The naming follows the Go
// backends rather than any C# convention, so that a token is called the same in every language a scanner is generated
// for.
func tokenName(ruleIdx int, rule frontend.Rule) string {
	name := utils.GoIdentifier(rule.Name)
	return "Token" + name
}

// tokenValue returns the value the generated scanner gives the token of the rule with the given index. The rules follow
// after the synthetic tokens, which take the lowest values.
func tokenValue(ruleIdx int) int {
	return syntheticTokenCount + ruleIdx
}
