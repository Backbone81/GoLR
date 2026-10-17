package model_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/backbone81/golr/internal/convert"
	"github.com/backbone81/golr/internal/lsp/model"
	bisonfrontend "github.com/backbone81/golr/internal/parsergen/frontend/bison"
	golrfrontend "github.com/backbone81/golr/internal/parsergen/frontend/golr"
	"github.com/backbone81/golr/testdata"
)

// golangGrammarFilePath is the largest hand written grammar file of the repository.
const golangGrammarFilePath = "../../../examples/golang/spec/golang.golr"

// wellKnownGrammarText returns the well-known grammar converted to a golr file, the way `golr convert` does it.
func wellKnownGrammarText(wellKnownGrammar testdata.WellKnownGrammar) ([]byte, error) {
	grammar, err := bisonfrontend.ToGrammar(bytes.NewReader(wellKnownGrammar.Content()), wellKnownGrammar.FileName)
	if err != nil {
		return nil, err
	}
	grammar, err = convert.BisonGrammar2GoLR(grammar)
	if err != nil {
		return nil, err
	}
	var text bytes.Buffer
	if err := golrfrontend.FromGrammar(&text, nil, grammar); err != nil {
		return nil, err
	}
	return text.Bytes(), nil
}

// BenchmarkBuild builds the model of the golang example and of the well-known grammars converted to golr files. A
// well-known grammar which the converter cannot convert is skipped.
func BenchmarkBuild(b *testing.B) {
	b.Run("GoLR golang example", func(b *testing.B) {
		text, err := os.ReadFile(golangGrammarFilePath)
		if err != nil {
			b.Fatal(err)
		}

		for b.Loop() {
			model.Build(text)
		}
	})

	for _, wellKnownGrammar := range testdata.WellKnownGrammars {
		b.Run(wellKnownGrammar.Title, func(b *testing.B) {
			text, err := wellKnownGrammarText(wellKnownGrammar)
			if err != nil {
				b.Skip("conversion to golr failed: ", err)
			}

			for b.Loop() {
				model.Build(text)
			}
		})
	}
}
