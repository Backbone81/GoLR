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

	"golr/internal/scannergen/frontend"
)

// ToRules reads the scanner rules as YAML document from the given reader. Returns an error if the YAML
// document can not be decoded successfully.
func ToRules(reader io.Reader) ([]frontend.Rule, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Frontends: YAML: ToRules").End()

	var rules []frontend.Rule
	if err := yaml.NewDecoder(reader).Decode(&rules); err != nil {
		return nil, fmt.Errorf("decoding YAML to rules: %w", err)
	}
	return rules, nil
}

// FromRules writes the scanner rules as YAML document to the given writer. Returns an error if the YAML
// document can not be encoded successfully.
func FromRules(writer io.Writer, rules []frontend.Rule) (err error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Frontends: YAML: FromRules").End()

	encoder := yaml.NewEncoder(writer)
	defer func() {
		if closeErr := encoder.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing YAML encoder: %w", closeErr))
		}
	}()

	if err := encoder.Encode(rules); err != nil {
		return fmt.Errorf("encoding rules to YAML: %w", err)
	}
	return nil
}

// RulesFromFile reads the scanner rules as YAML document from the given file path. Returns an error if the
// file can not be read or the YAML document can not be decoded successfully.
func RulesFromFile(filePath string) (rules []frontend.Rule, err error) { //nolint:nonamedreturns // Required for defer
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening the YAML file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return ToRules(file)
}

// RulesToFile writes the scanner rules as YAML document to the given file path. Returns an error if the file
// can not be written or the YAML document can not be encoded successfully.
func RulesToFile(filePath string, rules []frontend.Rule) (err error) {
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

	return FromRules(file, rules)
}

// RulesFromString reads the scanner rules as YAML document from the given string. Returns an error if the
// YAML document can not be decoded successfully.
func RulesFromString(yamlDocument string) ([]frontend.Rule, error) {
	return ToRules(strings.NewReader(yamlDocument))
}

// RulesToString returns the scanner rules as YAML document. Returns an error if the YAML document can not be
// encoded successfully.
func RulesToString(rules []frontend.Rule) (string, error) {
	var builder strings.Builder
	if err := FromRules(&builder, rules); err != nil {
		return "", err
	}
	return builder.String(), nil
}
