package dot

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/trace"
	"strings"
	"text/template"

	"github.com/backbone81/golr/internal/scannergen/backend"
	"github.com/backbone81/golr/internal/scannergen/frontend"
)

//go:embed scanner.dot.template
var scannerTemplate string

var parsedTemplate = template.Must(template.New("scanner.dot.template").Funcs(template.FuncMap{
	"dotLabel": dotLabel,
}).Parse(scannerTemplate))

type TemplateContext struct {
	DFA backend.DFA
}

// FromDFA writes the DFA as DOT document to the given writer. Returns an error if the DOT document can not be
// encoded successfully.
func FromDFA(writer io.Writer, dfa backend.DFA) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Backend: DOT: FromDFA").End()

	if err := parsedTemplate.Execute(writer, TemplateContext{
		DFA: dfa,
	}); err != nil {
		return fmt.Errorf("rendering the template: %w", err)
	}
	return nil
}

// DFAToFile writes the DFA as DOT document to the given file path. Returns an error if the file can not be
// written or the DOT source code can not be encoded successfully.
func DFAToFile(filePath string, dfa backend.DFA) (err error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the DOT file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return FromDFA(file, dfa)
}

// DFAToString returns the parser as DOT document. Returns an error if the DOT document can not be encoded
// successfully.
func DFAToString(dfa backend.DFA) (string, error) {
	var builder strings.Builder
	if err := FromDFA(&builder, dfa); err != nil {
		return "", err
	}
	return builder.String(), nil
}

func dotLabel(charRange frontend.CharRange) string {
	s := charRange.String()
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
