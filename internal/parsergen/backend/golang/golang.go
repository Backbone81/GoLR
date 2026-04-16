package golang

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"go/format"
	"golr/internal/parsergen/backend"
	"io"
	"os"
	"runtime/trace"
	"strings"
	"text/template"
)

//go:embed parser.go.template
var parserTemplate string

var parsedTemplate = template.Must(template.New("parser.go.template").Parse(parserTemplate))

type Config struct {
	PackageName string
}

type TemplateContext struct {
	Config Config
	Parser backend.Parser
}

// FromParser writes the parser as Go source code to the given writer. Returns an error if the Go source code can not be
// encoded successfully.
func FromParser(writer io.Writer, parser backend.Parser, config Config) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Backends: Golang: FromParser").End()

	var buffer bytes.Buffer
	if err := parsedTemplate.Execute(&buffer, TemplateContext{
		Config: config,
		Parser: parser,
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

// ParserToFile writes the parser as Go source code to the given file path. Returns an error if the file can not be
// written or the Go source code can not be encoded successfully.
func ParserToFile(filePath string, parser backend.Parser, config Config) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the Go file %q: %w", filePath, err)
	}
	defer file.Close()

	return FromParser(file, parser, config)
}

// ParserToString returns the parser as Go source code. Returns an error if the Go source code can not be encoded
// successfully.
func ParserToString(parser backend.Parser, config Config) (string, error) {
	var builder strings.Builder
	if err := FromParser(&builder, parser, config); err != nil {
		return "", err
	}
	return builder.String(), nil
}
