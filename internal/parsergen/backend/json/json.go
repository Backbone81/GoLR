package json

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/trace"
	"strings"

	"golr/internal/parsergen/backend"
)

// ToParser reads the parser as JSON document from the given reader. Returns an error if the JSON document can not be
// decoded successfully.
func ToParser(reader io.Reader) (backend.Parser, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Backends: JSON: ToParser").End()

	var result backend.Parser
	if err := json.NewDecoder(reader).Decode(&result); err != nil {
		return backend.Parser{}, fmt.Errorf("decoding JSON to parser: %w", err)
	}
	return result, nil
}

// FromParser writes the parser as JSON document to the given writer. Returns an error if the JSON document can not be
// encoded successfully.
func FromParser(writer io.Writer, parser backend.Parser) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Backends: JSON: FromParser").End()

	if err := json.NewEncoder(writer).Encode(&parser); err != nil {
		return fmt.Errorf("encoding JSON from parser: %w", err)
	}
	return nil
}

// ParserFromFile reads the parser as JSON document from the given file path. Returns an error if the file can not be
// read or the JSON document can not be decoded successfully.
func ParserFromFile(filePath string) (parser backend.Parser, err error) { //nolint:nonamedreturns // Required for defer
	file, err := os.Open(filePath) //nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	if err != nil {
		return backend.Parser{}, fmt.Errorf("opening the JSON file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return ToParser(file)
}

// ParserToFile writes the parser as JSON document to the given file path. Returns an error if the file can not be
// written or the JSON document can not be encoded successfully.
func ParserToFile(filePath string, parser backend.Parser) (err error) {
	file, err := os.Create(filePath) //nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	if err != nil {
		return fmt.Errorf("creating the JSON file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return FromParser(file, parser)
}

// ParserFromString reads the parser as JSON document from the given string. Returns an error if the JSON document
// can not be decoded successfully.
func ParserFromString(jsonDocument string) (backend.Parser, error) {
	return ToParser(strings.NewReader(jsonDocument))
}

// ParserToString returns the parser as JSON document. Returns an error if the JSON document can not be encoded
// successfully.
func ParserToString(parser backend.Parser) (string, error) {
	var builder strings.Builder
	if err := FromParser(&builder, parser); err != nil {
		return "", err
	}
	return builder.String(), nil
}
