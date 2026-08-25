package golang

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"go/format"
	"io"
	"os"
	"runtime/trace"
	"text/template"

	"github.com/backbone81/golr/internal/scannergen/backend"
	"github.com/backbone81/golr/internal/scannergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

// invalidTokenName is the name of the token constant the generated scanner uses when it has no token, which is what the
// accept table holds for a state which does not accept.
const invalidTokenName = "InvalidToken"

//go:embed scanner.go.template
var scannerTemplate string

var parsedTemplate = template.Must(template.New("scanner.go.template").Funcs(template.FuncMap{
	"tokenName": tokenName,
}).Parse(scannerTemplate))

// DefaultPackageName is the Go package the generated scanner is declared in when the caller names none.
const DefaultPackageName = "parser"

type Config struct {
	// PackageName is the Go package the generated scanner is declared in.
	//
	// The generated parser has to be given the same one. Go resolves a type by its package and the parser needs both
	// the token constants and the scanner interface declared here.
	PackageName string
}

type TemplateContext struct {
	Config Config
	DFA    backend.DFA
	Tables Tables
}

// FromDFA writes the DFA as Go source code to the given writer. Returns an error if the Go source code can not be
// encoded successfully.
func FromDFA(writer io.Writer, dfa backend.DFA, config Config) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Backend: Golang Table: FromDFA").End()

	if len(dfa.States) == 0 {
		return errors.New("the DFA does not have any state")
	}

	if config.PackageName == "" {
		config.PackageName = DefaultPackageName
	}

	var buffer bytes.Buffer
	if err := parsedTemplate.Execute(&buffer, TemplateContext{
		Config: config,
		DFA:    dfa,
		Tables: NewTables(dfa),
	}); err != nil {
		return fmt.Errorf("rendering the template: %w", err)
	}
	source := buffer.Bytes()

	var joinedErr error
	formatted, err := format.Source(source)
	if err != nil {
		joinedErr = errors.Join(joinedErr, err)
	} else {
		source = formatted
	}

	if _, err := writer.Write(source); err != nil {
		joinedErr = errors.Join(joinedErr, err)
	}
	return joinedErr
}

// DFAToFile writes the DFA as Go source code to the given file path. Returns an error if the file can not be
// written or the Go source code can not be encoded successfully.
func DFAToFile(filePath string, dfa backend.DFA, config Config) (err error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the Go file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return FromDFA(file, dfa, config)
}

func tokenName(ruleIdx int, rule frontend.Rule) string {
	name := utils.GoIdentifier(rule.Name)
	return "Token" + name
}
