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
	"strings"
	"text/template"

	"github.com/backbone81/golr/internal/scannergen/backend"
	"github.com/backbone81/golr/internal/scannergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

//go:embed scanner.java.template
var scannerTemplate string

var parsedTemplate = template.Must(template.New("scanner.java.template").Funcs(template.FuncMap{
	"printByte": printByte,
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

// FromDFA writes the DFA as Java source code to the given writer. Returns an error if the Java source code can not be
// encoded successfully.
func FromDFA(writer io.Writer, dfa backend.DFA, config Config) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Backend: Java: FromDFA").End()

	var buffer bytes.Buffer
	if err := parsedTemplate.Execute(&buffer, TemplateContext{
		Config: config,
		DFA:    dfa,
	}); err != nil {
		return fmt.Errorf("rendering the template: %w", err)
	}
	source := buffer.Bytes()

	if _, err := writer.Write(source); err != nil {
		return err
	}
	return nil
}

// DFAToFile writes the DFA as Go source code to the given file path. Returns an error if the file can not be
// written or the Go source code can not be encoded successfully.
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

// printByte returns a string for a rune which is safe to use in Java source. Standard ASCII characters are printed
// as is to be human-readable. Special characters which are not printable or any Unicode codepoint is printed as
// its hexadecimal value. That way the direct coded scanner can be easily inspected and debugged by a human.
func printByte(r byte) string {
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
	return strings.ToUpper(rule.Name)
}
