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

	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// ToGrammar reads the context free grammar as JSON document from the given reader. Returns an error if the JSON
// document can not be decoded successfully.
func ToGrammar(reader io.Reader) (frontend.Grammar, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Frontends: JSON: ToGrammar").End()

	var result frontend.Grammar
	if err := json.NewDecoder(reader).Decode(&result); err != nil {
		return frontend.Grammar{}, fmt.Errorf("decoding JSON to grammar: %w", err)
	}
	return result, nil
}

// FromGrammar writes the context free grammar as JSON document to the given writer. Returns an error if the JSON
// document can not be encoded successfully.
func FromGrammar(writer io.Writer, grammar frontend.Grammar) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Frontends: JSON: FromGrammar").End()

	if err := json.NewEncoder(writer).Encode(grammar); err != nil {
		return fmt.Errorf("encoding grammar to JSON: %w", err)
	}
	return nil
}

// GrammarFromFile reads the context free grammar as JSON document from the given file path. Returns an error if the
// file can not be read or the JSON document can not be decoded successfully.
//
//nolint:nonamedreturns // Required for defer
func GrammarFromFile(filePath string) (grammar frontend.Grammar, err error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Open(filePath)
	if err != nil {
		return frontend.Grammar{}, fmt.Errorf("opening the JSON file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return ToGrammar(file)
}

// GrammarToFile writes the context free grammar as JSON document to the given file path. Returns an error if the file
// can not be written or the JSON document can not be encoded successfully.
func GrammarToFile(filePath string, grammar frontend.Grammar) (err error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the JSON file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return FromGrammar(file, grammar)
}

// GrammarFromString reads the context free grammar as JSON document from the given string. Returns an error if the
// JSON document can not be decoded successfully.
func GrammarFromString(jsonDocument string) (frontend.Grammar, error) {
	return ToGrammar(strings.NewReader(jsonDocument))
}

// GrammarToString returns the context free grammar as JSON document. Returns an error if the JSON document can not be
// encoded successfully.
func GrammarToString(grammar frontend.Grammar) (string, error) {
	var builder strings.Builder
	if err := FromGrammar(&builder, grammar); err != nil {
		return "", err
	}
	return builder.String(), nil
}
