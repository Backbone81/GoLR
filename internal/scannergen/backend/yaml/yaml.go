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

	"golr/internal/scannergen/backend"
)

// ToDFA reads the deterministic finite automaton as YAML document from the given reader. Returns an error if the YAML
// document can not be decoded successfully.
func ToDFA(reader io.Reader) (backend.DFA, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Backends: YAML: ToDFA").End()

	var result backend.DFA
	if err := yaml.NewDecoder(reader).Decode(&result); err != nil {
		return backend.DFA{}, fmt.Errorf("decoding YAML to DFA: %w", err)
	}
	return result, nil
}

// FromDFA writes the deterministic finite automaton as YAML document to the given writer. Returns an error if the YAML
// document can not be encoded successfully.
func FromDFA(writer io.Writer, dfa backend.DFA) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Frontends: YAML: FromDFA").End()

	if err := yaml.NewEncoder(writer).Encode(&dfa); err != nil {
		return err
	}
	return nil
}

// DFAFromFile reads the deterministic finite automaton as YAML document from the given file path. Returns an error if
// the file can not be read or the YAML document can not be decoded successfully.
func DFAFromFile(filePath string) (dfa backend.DFA, err error) { //nolint:nonamedreturns // Required for defer
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Open(filePath)
	if err != nil {
		return backend.DFA{}, fmt.Errorf("opening the YAML file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return ToDFA(file)
}

// DFAToFile writes the deterministic finite automaton as YAML document to the given file path. Returns an error if the
// file can not be written or the YAML document can not be encoded successfully.
func DFAToFile(filePath string, inputDFA backend.DFA) (err error) {
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

	return FromDFA(file, inputDFA)
}

// DFAFromString reads the deterministic finite automaton as YAML document from the given string. Returns an error if
// the YAML document can not be decoded successfully.
func DFAFromString(yamlDocument string) (backend.DFA, error) {
	return ToDFA(strings.NewReader(yamlDocument))
}

// DFAToString returns the deterministic finite automaton as YAML document. Returns an error if the YAML document can
// not be encoded successfully.
func DFAToString(inputDFA backend.DFA) (string, error) {
	var builder strings.Builder
	if err := FromDFA(&builder, inputDFA); err != nil {
		return "", err
	}
	return builder.String(), nil
}
