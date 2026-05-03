package json

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime/trace"
	"strings"

	"golr/internal/scannergen/backend"
)

// ToDFA reads the deterministic finite automaton as JSON document from the given reader. Returns an error if the JSON
// document can not be decoded successfully.
func ToDFA(reader io.Reader) (backend.DFA, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Backends: JSON: ToDFA").End()

	var result backend.DFA
	if err := json.NewDecoder(reader).Decode(&result); err != nil {
		return backend.DFA{}, fmt.Errorf("decoding JSON to DFA: %w", err)
	}
	return result, nil
}

// FromDFA writes the deterministic finite automaton as JSON document to the given writer. Returns an error if the JSON
// document can not be encoded successfully.
func FromDFA(writer io.Writer, dfa backend.DFA) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Frontends: JSON: FromDFA").End()

	if err := json.NewEncoder(writer).Encode(&dfa); err != nil {
		return fmt.Errorf("encoding JSON from DFA: %w", err)
	}
	return nil
}

// DFAFromFile reads the deterministic finite automaton as JSON document from the given file path. Returns an error if the
// file can not be read or the JSON document can not be decoded successfully.
func DFAFromFile(filePath string) (backend.DFA, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return backend.DFA{}, fmt.Errorf("opening the JSON file %q: %w", filePath, err)
	}
	defer file.Close()

	return ToDFA(file)
}

// DFAToFile writes the deterministic finite automaton as JSON document to the given file path. Returns an error if the file
// can not be written or the JSON document can not be encoded successfully.
func DFAToFile(filePath string, inputDFA backend.DFA) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the JSON file %q: %w", filePath, err)
	}
	defer file.Close()

	return FromDFA(file, inputDFA)
}

// DFAFromString reads the deterministic finite automaton as JSON document from the given string. Returns an error if the
// JSON document can not be decoded successfully.
func DFAFromString(jsonDocument string) (backend.DFA, error) {
	return ToDFA(strings.NewReader(jsonDocument))
}

// DFAToString returns the deterministic finite automaton as JSON document. Returns an error if the JSON document can not be
// encoded successfully.
func DFAToString(inputDFA backend.DFA) (string, error) {
	var builder strings.Builder
	if err := FromDFA(&builder, inputDFA); err != nil {
		return "", err
	}
	return builder.String(), nil
}
