package golr

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/trace"
	"strings"

	parserfrontend "github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/parsergen/frontend/golr"
	golrparser "github.com/backbone81/golr/internal/parsergen/frontend/golr/parser"
	scannerfrontend "github.com/backbone81/golr/internal/scannergen/frontend"
)

// ToRules reads the scanner rules as GoLR grammar document from the given reader. Returns an error if the GoLR
// document can not be decoded successfully.
func ToRules(reader io.Reader, filePath string) ([]scannerfrontend.Rule, parserfrontend.Grammar, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Frontends: GoLR: ToRules").End()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, parserfrontend.Grammar{}, err
	}

	scanner := golrparser.WhitespaceSkipper{
		Scanner: golrparser.NewScanner(data, filePath),
	}

	parser := golrparser.NewParser()
	rootNode, err := parser.Parse(&scanner)
	if err != nil {
		return nil, parserfrontend.Grammar{}, err
	}

	walker := golr.NewASTWalker()
	rules, grammar, err := walker.BuildGrammar(rootNode)
	if err != nil {
		return nil, parserfrontend.Grammar{}, err
	}
	return rules, grammar, nil
}

// RulesFromFile reads the scanner rules as GoLR grammar document from the given file path. Returns an error if the
// file can not be read or the GoLR document can not be decoded successfully.
//
//nolint:nonamedreturns // Required for defer
func RulesFromFile(filePath string) (rules []scannerfrontend.Rule, grammar parserfrontend.Grammar, err error) {
	file, err := os.Open(filePath) //nolint:gosec // The caller is responsible for making sure the path is safe.
	if err != nil {
		return nil, parserfrontend.Grammar{}, fmt.Errorf("opening the GoLR file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return ToRules(file, filePath)
}

// RulesFromString reads the scanner rules as GoLR grammar document from the given string. Returns an error if the
// GoLR document can not be decoded successfully.
func RulesFromString(jsonDocument string) ([]scannerfrontend.Rule, parserfrontend.Grammar, error) {
	return ToRules(strings.NewReader(jsonDocument), "in-memory")
}
