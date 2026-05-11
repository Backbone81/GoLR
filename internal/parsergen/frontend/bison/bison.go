package bison

import (
	"context"
	"fmt"
	"golr/internal/parsergen/frontend"
	bisonparser "golr/internal/parsergen/frontend/bison/parser"
	"golr/pkg/runtime"
	"io"
	"os"
	"runtime/trace"
	"strings"
)

// ToGrammar reads the context free grammar as GNU Bison grammar document from the given reader. Returns an error if the
// grammar document can not be parsed successfully.
func ToGrammar(reader io.Reader, filePath string) (frontend.Grammar, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Frontends: Bison: ToGrammar").End()

	data, err := io.ReadAll(reader)
	if err != nil {
		return frontend.Grammar{}, err
	}

	runeReader := runtime.NewUTF8RuneReader(data)
	scanner := bisonparser.TokenTransformer{
		Scanner: &bisonparser.WhitespaceSkipper{
			Scanner: &bisonparser.ContextScanner{
				Scanner: bisonparser.NewScanner(runeReader, filePath),
			},
		},
	}

	parser := bisonparser.NewParser()
	rootNode, err := parser.Parse(&scanner)
	if err != nil {
		return frontend.Grammar{}, err
	}

	walker := NewASTWalker()
	return walker.BuildGrammar(rootNode)
}

// GrammarFromFile reads the context free grammar as GNU Bison grammar document from the given file path. Returns an
// error if the file can not be read or the grammar document can not be parsed successfully.
func GrammarFromFile(filePath string) (frontend.Grammar, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return frontend.Grammar{}, fmt.Errorf("opening the Bison file %q: %w", filePath, err)
	}
	defer file.Close()

	return ToGrammar(file, filePath)
}

// GrammarFromString reads the context free grammar as GNU Bison grammar document from the given string. Returns an
// error if the grammar document can not be parsed successfully.
func GrammarFromString(bisonGrammar string) (frontend.Grammar, error) {
	return ToGrammar(strings.NewReader(bisonGrammar), "in-memory")
}
