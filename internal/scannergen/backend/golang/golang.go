package golang

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"go/format"
	"golr/internal/scannergen/backend"
	"golr/internal/scannergen/frontend"
	"io"
	"os"
	"runtime/trace"
	"strings"
	"text/template"
	"unicode"
)

//go:embed scanner.go.template
var scannerTemplate string

var parsedTemplate = template.Must(template.New("scanner.go.template").Funcs(template.FuncMap{
	"printRune":    printRune,
	"terminalName": terminalName,
	"stateName":    stateName,
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
func DFAToFile(filePath string, dfa backend.DFA, config Config) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the Go file %q: %w", filePath, err)
	}
	defer file.Close()

	return FromDFA(file, dfa, config)
}

// printRune returns a string for a rune which is safe to use in go source. Standard ASCII characters are printed
// as is to be human-readable. Special characters which are not printable or any Unicode codepoint is printed as
// its hexadecimal value. That way the direct coded scanner can be easily inspected and debugged by a human.
func printRune(r rune) string {
	switch {
	case r == ' ':
		return "' '"
	case r == '\t':
		return "'\\t'"
	case r == '\r':
		return "'\\r'"
	case r == '\n':
		return "'\\n'"
	case r == '\'':
		return "'\\''"
	case r == '\\':
		return "'\\\\'"
	default:
		if 32 <= r && r < 127 {
			return fmt.Sprintf("'%c'", r)
		}
		return fmt.Sprintf("0x%x", r)
	}
}

func stateName(stateIdx int, rule frontend.Rule) string {
	name := goName(rule.Name)
	return fmt.Sprintf("state%d%s", stateIdx, name)
}

func terminalName(ruleIdx int, rule frontend.Rule) string {
	name := goName(rule.Name)
	return fmt.Sprintf("Terminal%s", name)
}

func goName(text string) string {
	words := strings.FieldsFunc(text, func(r rune) bool {
		return r == '_' || r == ' ' || r == '\t'
	})

	var builder strings.Builder
	for _, word := range words {
		if len(word) == 0 {
			continue
		}

		cleaned := replaceSpecialCharacters(word)
		capitalized := capitalizeFirstChar(cleaned)
		builder.WriteString(capitalized)
	}
	return builder.String()
}

func replaceSpecialCharacters(text string) string {
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

func capitalizeFirstChar(text string) string {
	var builder strings.Builder
	for i, r := range text {
		if i == 0 {
			builder.WriteRune(unicode.ToUpper(r))
		} else {
			builder.WriteRune(unicode.ToLower(r))
		}
	}
	return builder.String()
}
