package javascript

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

//go:embed scanner.js.template
var scannerTemplate string

var parsedTemplate = template.Must(template.New("scanner.js.template").Funcs(template.FuncMap{
	"tokenName":  tokenName,
	"tokenValue": tokenValue,
}).Parse(scannerTemplate))

type TemplateContext struct {
	Tables Tables
	DFA    backend.DFA
}

// FromDFA writes the DFA as JavaScript source code to the given writer. Returns an error if the JavaScript source code
// can not be encoded successfully.
func FromDFA(writer io.Writer, dfa backend.DFA) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Backend: JavaScript: FromDFA").End()

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

// DFAToFile writes the DFA as JavaScript source code to the given file path. Returns an error if the file can not be
// written or the JavaScript source code can not be encoded successfully.
func DFAToFile(filePath string, dfa backend.DFA) (err error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the JavaScript file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return FromDFA(file, dfa)
}

// tokenName returns the name of the token the generated scanner declares for the given rule. The naming follows the Go
// backends rather than any JavaScript convention, so that a token is called the same in every language a scanner is
// generated for.
func tokenName(ruleIdx int, rule frontend.Rule) string {
	name := utils.GoIdentifier(rule.Name)
	return "Token" + name
}

// tokenValue returns the value the generated scanner gives the token of the rule with the given index. JavaScript has
// no counterpart to the iota the Go backends number their tokens with, so the value has to be written out, and the
// rules follow after the synthetic tokens which take the lowest values.
func tokenValue(ruleIdx int) int {
	return syntheticTokenCount + ruleIdx
}
