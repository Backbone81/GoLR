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
	"text/template"

	"github.com/backbone81/golr/internal/scannergen/backend"
	"github.com/backbone81/golr/internal/scannergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

//go:embed scanner.py.template
var scannerTemplate string

var parsedTemplate = template.Must(template.New("scanner.py.template").Funcs(template.FuncMap{
	"tokenName":  tokenName,
	"tokenValue": tokenValue,
}).Parse(scannerTemplate))

type TemplateContext struct {
	Tables Tables
	DFA    backend.DFA
}

// FromDFA writes the DFA as Python source code to the given writer. Returns an error if the Python source code can not
// be encoded successfully.
func FromDFA(writer io.Writer, dfa backend.DFA) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Backend: Python: FromDFA").End()

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

// DFAToFile writes the DFA as Python source code to the given file path. Returns an error if the file can not be
// written or the Python source code can not be encoded successfully.
func DFAToFile(filePath string, dfa backend.DFA) (err error) {
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

	return FromDFA(file, dfa)
}

// tokenName returns the name of the token member the generated scanner declares for the given rule. Python spells an
// enumeration member in upper case with underscores, so a token is called TOKEN_NUMBER the way the Java backend names
// it rather than TokenNumber.
func tokenName(ruleIdx int, rule frontend.Rule) string {
	return utils.PythonConstantName("Token" + utils.GoIdentifier(rule.Name))
}

// tokenValue returns the value the generated scanner gives the token of the rule with the given index. The values are
// written out rather than left to the enumeration, so that the rules follow after the synthetic tokens which take the
// lowest values.
func tokenValue(ruleIdx int) int {
	return syntheticTokenCount + ruleIdx
}
