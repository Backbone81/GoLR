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

//go:embed scanner.go.template
var scannerTemplate string

var parsedTemplate = template.Must(template.New("scanner.go.template").Funcs(template.FuncMap{
	"printRune": printRune,
	"tokenName": tokenName,
	"stateName": stateName,
}).Parse(scannerTemplate))

type Config struct {
	PackageName string
}

type TemplateContext struct {
	Config Config
	DFA    backend.DFA
}

// FromDFA writes the DFA as Go source code to the given writer. Returns an error if the Go source code can not be
// encoded successfully.
func FromDFA(writer io.Writer, dfa backend.DFA, config Config) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Backend: Golang: FromDFA").End()

	var buffer bytes.Buffer
	if err := parsedTemplate.Execute(&buffer, TemplateContext{
		Config: config,
		DFA:    dfa,
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

// printRune returns a string for a rune which is safe to use in go source. Standard ASCII characters are printed
// as is to be human-readable. Special characters which are not printable or any Unicode codepoint is printed as
// its hexadecimal value. That way the direct coded scanner can be easily inspected and debugged by a human.
func printRune(r rune) string {
	switch r {
	case ' ':
		return "' '"
	case '\t':
		return "'\\t'"
	case '\r':
		return "'\\r'"
	case '\n':
		return "'\\n'"
	case '\'':
		return "'\\''"
	case '\\':
		return "'\\\\'"
	default:
		if 32 <= r && r < 127 {
			return fmt.Sprintf("'%c'", r)
		}
		return fmt.Sprintf("0x%x", r)
	}
}

func stateName(stateIdx int, rule frontend.Rule) string {
	name := utils.GoIdentifier(rule.Name)
	return fmt.Sprintf("state%d%s", stateIdx, name)
}

func tokenName(ruleIdx int, rule frontend.Rule) string {
	name := utils.GoIdentifier(rule.Name)
	return "Token" + name
}
