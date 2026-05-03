package yaml

import (
	"context"
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
func FromRules(writer io.Writer, rules []frontend.Rule) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Frontends: YAML: FromRules").End()

	encoder := yaml.NewEncoder(writer)
	defer encoder.Close()

	if err := encoder.Encode(rules); err != nil {
		return fmt.Errorf("encoding rules to YAML: %w", err)
	}
	return nil
}

// RulesFromFile reads the scanner rules as YAML document from the given file path. Returns an error if the
// file can not be read or the YAML document can not be decoded successfully.
func RulesFromFile(filePath string) ([]frontend.Rule, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return []frontend.Rule{}, fmt.Errorf("opening the YAML file %q: %w", filePath, err)
	}
	defer file.Close()

	return ToRules(file)
}

// RulesToFile writes the scanner rules as YAML document to the given file path. Returns an error if the file
// can not be written or the YAML document can not be encoded successfully.
func RulesToFile(filePath string, rules []frontend.Rule) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the YAML file %q: %w", filePath, err)
	}
	defer file.Close()

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
