package yaml

import (
	"context"
	"fmt"
	"golr/internal/parsergen/frontend"

	"github.com/goccy/go-yaml"

	"io"
	"os"
	"runtime/trace"
	"strings"
)

// ToGrammar reads the context free grammar as YAML document from the given reader. Returns an error if the YAML
// document can not be decoded successfully.
func ToGrammar(reader io.Reader) (frontend.Grammar, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Frontends: YAML: ToGrammar").End()

	var result frontend.Grammar
	if err := yaml.NewDecoder(reader).Decode(&result); err != nil {
		return frontend.Grammar{}, fmt.Errorf("decoding YAML to grammar: %w", err)
	}
	return result, nil
}

// FromGrammar writes the context free grammar as YAML document to the given writer. Returns an error if the YAML
// document can not be encoded successfully.
func FromGrammar(writer io.Writer, grammar frontend.Grammar) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Frontends: YAML: FromGrammar").End()

	encoder := yaml.NewEncoder(writer)
	defer encoder.Close()

	if err := encoder.Encode(grammar); err != nil {
		return fmt.Errorf("encoding grammar to YAML: %w", err)
	}
	return nil
}

// GrammarFromFile reads the context free grammar as YAML document from the given file path. Returns an error if the
// file can not be read or the YAML document can not be decoded successfully.
func GrammarFromFile(filePath string) (frontend.Grammar, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return frontend.Grammar{}, fmt.Errorf("opening the YAML file %q: %w", filePath, err)
	}
	defer file.Close()

	return ToGrammar(file)
}

// GrammarToFile writes the context free grammar as YAML document to the given file path. Returns an error if the file
// can not be written or the YAML document can not be encoded successfully.
func GrammarToFile(filePath string, grammar frontend.Grammar) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the YAML file %q: %w", filePath, err)
	}
	defer file.Close()

	return FromGrammar(file, grammar)
}

// GrammarFromString reads the context free grammar as YAML document from the given string. Returns an error if the
// YAML document can not be decoded successfully.
func GrammarFromString(yamlDocument string) (frontend.Grammar, error) {
	return ToGrammar(strings.NewReader(yamlDocument))
}

// GrammarToString returns the context free grammar as YAML document. Returns an error if the YAML document can not be
// encoded successfully.
func GrammarToString(grammar frontend.Grammar) (string, error) {
	var builder strings.Builder
	if err := FromGrammar(&builder, grammar); err != nil {
		return "", err
	}
	return builder.String(), nil
}
