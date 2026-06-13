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
	"strings"
	"text/template"

	"github.com/backbone81/golr/internal/scannergen/backend"
	"github.com/backbone81/golr/internal/scannergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

//go:embed scanner.rs.template
var scannerTemplate string

var parsedTemplate = template.Must(template.New("scanner.rs.template").Funcs(template.FuncMap{
	"printByte": printByte,
	"tokenName": tokenName,
	"stateName": stateName,
}).Parse(scannerTemplate))

type TemplateContext struct {
	DFA backend.DFA
}

// FromDFA writes the DFA as Rust source code to the given writer. Returns an error if the Rust source code can not be
// encoded successfully.
func FromDFA(writer io.Writer, dfa backend.DFA) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Backend: Rust: FromDFA").End()

	var buffer bytes.Buffer
	if err := parsedTemplate.Execute(&buffer, TemplateContext{
		DFA: dfa,
	}); err != nil {
		return fmt.Errorf("rendering the template: %w", err)
	}
	source := buffer.Bytes()

	if _, err := writer.Write(source); err != nil {
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

// printByte returns a string for a rune which is safe to use in Rust source. Standard ASCII characters are printed
// as is to be human-readable. Special characters which are not printable or any Unicode codepoint is printed as
// its hexadecimal value. That way the direct coded scanner can be easily inspected and debugged by a human.
func printByte(r byte) string {
	switch r {
	case ' ':
		return "b' '"
	case '\t':
		return "b'\\t'"
	case '\r':
		return "b'\\r'"
	case '\n':
		return "b'\\n'"
	case '\'':
		return "b'\\''"
	case '\\':
		return "b'\\\\'"
	default:
		if 32 <= r && r < 127 {
			return fmt.Sprintf("b'%c'", r)
		}
		return fmt.Sprintf("0x%x", r)
	}
}

func stateName(stateIdx int, rule frontend.Rule) string {
	name := RustIdentifier(rule.Name)
	return fmt.Sprintf("state_%d_%s", stateIdx, name)
}

func tokenName(ruleIdx int, rule frontend.Rule) string {
	name := utils.GoIdentifier(rule.Name)
	return "Token" + name
}

// RustIdentifier creates a snake_case name suitable as a Rust identifier. Is used for code generation.
func RustIdentifier(text string) string {
	words := strings.FieldsFunc(text, func(r rune) bool {
		return r == '_' || r == ' ' || r == '\t'
	})

	var parts []string
	for _, word := range words {
		if len(word) == 0 {
			continue
		}
		cleaned := replaceSpecialCharactersRust(word)
		parts = append(parts, strings.ToLower(cleaned))
	}
	return strings.Join(parts, "_")
}

func replaceSpecialCharactersRust(text string) string {
	var builder strings.Builder
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
		} else {
			builder.WriteByte('_')
		}
	}
	return builder.String()
}
