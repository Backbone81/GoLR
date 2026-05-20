package golr

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/trace"
	"strings"

	"github.com/backbone81/golr/internal/parsergen/frontend"
	golrparser "github.com/backbone81/golr/internal/parsergen/frontend/golr/parser"
	"github.com/backbone81/golr/pkg/runtime"
)

// ToGrammar reads the context free grammar as GoLR grammar document from the given reader. Returns an error if the
// grammar document can not be parsed successfully.
func ToGrammar(reader io.Reader, filePath string) (frontend.Grammar, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Frontends: GoLR: ToGrammar").End()

	data, err := io.ReadAll(reader)
	if err != nil {
		return frontend.Grammar{}, err
	}

	runeReader := runtime.NewUTF8RuneReader(data)
	scanner := golrparser.WhitespaceSkipper{
		Scanner: golrparser.NewScanner(runeReader, filePath),
	}

	parser := golrparser.NewParser()
	rootNode, err := parser.Parse(&scanner)
	if err != nil {
		return frontend.Grammar{}, err
	}

	walker := NewASTWalker()
	_, grammar, err := walker.BuildGrammar(rootNode)
	if err != nil {
		return frontend.Grammar{}, err
	}
	return grammar, nil
}

// GrammarFromFile reads the context free grammar as GoLR grammar document from the given file path. Returns an
// error if the file can not be read or the grammar document can not be parsed successfully.
//
//nolint:nonamedreturns // Required for defer
func GrammarFromFile(filePath string) (grammar frontend.Grammar, err error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Open(filePath)
	if err != nil {
		return frontend.Grammar{}, fmt.Errorf("opening the GoLR file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return ToGrammar(file, filePath)
}

// GrammarFromString reads the context free grammar as GoLR grammar document from the given string. Returns an
// error if the grammar document can not be parsed successfully.
func GrammarFromString(golrGrammar string) (frontend.Grammar, error) {
	return ToGrammar(strings.NewReader(golrGrammar), "in-memory")
}
