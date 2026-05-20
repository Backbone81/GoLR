package golr

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/trace"
	"strings"

	"github.com/backbone81/golr/internal/parsergen/frontend/golr"
	golrparser "github.com/backbone81/golr/internal/parsergen/frontend/golr/parser"
	"github.com/backbone81/golr/internal/scannergen/frontend"
	"github.com/backbone81/golr/pkg/runtime"
)

// ToRules reads the scanner rules as GoLR grammar document from the given reader. Returns an error if the GoLR
// document can not be decoded successfully.
func ToRules(reader io.Reader, filePath string) ([]frontend.Rule, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Scannergen: Frontends: GoLR: ToRules").End()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	runeReader := runtime.NewUTF8RuneReader(data)
	scanner := golrparser.WhitespaceSkipper{
		Scanner: golrparser.NewScanner(runeReader, filePath),
	}

	parser := golrparser.NewParser()
	rootNode, err := parser.Parse(&scanner)
	if err != nil {
		return nil, err
	}

	walker := golr.NewASTWalker()
	rules, _, err := walker.BuildGrammar(rootNode)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

// RulesFromFile reads the scanner rules as GoLR grammar document from the given file path. Returns an error if the
// file can not be read or the GoLR document can not be decoded successfully.
func RulesFromFile(filePath string) (rules []frontend.Rule, err error) { //nolint:nonamedreturns // Required for defer
	file, err := os.Open(filePath) //nolint:gosec // The caller is responsible for making sure the path is safe.
	if err != nil {
		return nil, fmt.Errorf("opening the GoLR file %q: %w", filePath, err)
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
func RulesFromString(jsonDocument string) ([]frontend.Rule, error) {
	return ToRules(strings.NewReader(jsonDocument), "in-memory")
}
