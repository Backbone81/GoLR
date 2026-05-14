package yaml

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/trace"
	"strings"

	"github.com/goccy/go-yaml"

	"golr/internal/parsergen/backend"
)

// ToParser reads the parser as YAML document from the given reader. Returns an error if the YAML document can not be
// decoded successfully.
func ToParser(reader io.Reader) (backend.Parser, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Backends: YAML: ToParser").End()

	var result backend.Parser
	if err := yaml.NewDecoder(reader).Decode(&result); err != nil {
		return backend.Parser{}, fmt.Errorf("decoding YAML to parser: %w", err)
	}
	return result, nil
}

// FromParser writes the parser as YAML document to the given writer. Returns an error if the YAML document can not be
// encoded successfully.
func FromParser(writer io.Writer, parser backend.Parser) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Backends: YAML: FromParser").End()

	if err := yaml.NewEncoder(writer).Encode(&parser); err != nil {
		return err
	}
	return nil
}

// ParserFromFile reads the parser as YAML document from the given file path. Returns an error if the file can not be
// read or the YAML document can not be decoded successfully.
func ParserFromFile(filePath string) (parser backend.Parser, err error) { //nolint:nonamedreturns // Required for defer
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Open(filePath)
	if err != nil {
		return backend.Parser{}, fmt.Errorf("opening the YAML file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return ToParser(file)
}

// ParserToFile writes the parser as YAML document to the given file path. Returns an error if the file can not be
// written or the YAML document can not be encoded successfully.
func ParserToFile(filePath string, parser backend.Parser) (err error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the YAML file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return FromParser(file, parser)
}

// ParserFromString reads the parser as YAML document from the given string. Returns an error if the YAML document can
// not be decoded successfully.
func ParserFromString(yamlDocument string) (backend.Parser, error) {
	return ToParser(strings.NewReader(yamlDocument))
}

// ParserToString returns the parser as YAML document. Returns an error if the YAML document can not be encoded
// successfully.
func ParserToString(parser backend.Parser) (string, error) {
	var builder strings.Builder
	if err := FromParser(&builder, parser); err != nil {
		return "", err
	}
	return builder.String(), nil
}
