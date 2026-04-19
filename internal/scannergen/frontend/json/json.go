package json

import (
	"context"
	"encoding/json"
	"fmt"
	"golr/internal/scannergen/frontend"
	"io"
	"os"
	"runtime/trace"
	"strings"
)

// ToRules reads the scanner rules as JSON document from the given reader. Returns an error if the JSON
// document can not be decoded successfully.
func ToRules(reader io.Reader) ([]frontend.Rule, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Frontends: JSON: ToRules").End()

	var result []frontend.Rule
	if err := json.NewDecoder(reader).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding JSON to rules: %w", err)
	}
	return result, nil
}

// FromRules writes the scanner rules as JSON document to the given writer. Returns an error if the JSON
// document can not be encoded successfully.
func FromRules(writer io.Writer, rules []frontend.Rule) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Frontends: JSON: FromRules").End()

	if err := json.NewEncoder(writer).Encode(rules); err != nil {
		return fmt.Errorf("encoding rules to JSON: %w", err)
	}
	return nil
}

// RulesFromFile reads the scanner rules as JSON document from the given file path. Returns an error if the
// file can not be read or the JSON document can not be decoded successfully.
func RulesFromFile(filePath string) ([]frontend.Rule, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return []frontend.Rule{}, fmt.Errorf("opening the JSON file %q: %w", filePath, err)
	}
	defer file.Close()

	return ToRules(file)
}

// RulesToFile writes the scanner rules as JSON document to the given file path. Returns an error if the file
// can not be written or the JSON document can not be encoded successfully.
func RulesToFile(filePath string, rules []frontend.Rule) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the JSON file %q: %w", filePath, err)
	}
	defer file.Close()

	return FromRules(file, rules)
}

// RulesFromString reads the scanner rules as JSON document from the given string. Returns an error if the
// JSON document can not be decoded successfully.
func RulesFromString(jsonDocument string) ([]frontend.Rule, error) {
	return ToRules(strings.NewReader(jsonDocument))
}

// RulesToString returns the scanner rules as JSON document. Returns an error if the JSON document can not be
// encoded successfully.
func RulesToString(rules []frontend.Rule) (string, error) {
	var builder strings.Builder
	if err := FromRules(&builder, rules); err != nil {
		return "", err
	}
	return builder.String(), nil
}
