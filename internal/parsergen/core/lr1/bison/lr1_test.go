package bison_test

import (
	"bytes"
	"testing"

	lr1bisoncore "github.com/backbone81/golr/pkg/parsergen/core/lr1/bison"
	bisonfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/bison"
	"github.com/backbone81/golr/testdata"
)

func BenchmarkGrammarToParser(b *testing.B) {
	for _, wellKnownGrammar := range testdata.WellKnownGrammars {
		b.Run(wellKnownGrammar.Title, func(b *testing.B) {
			grammar, err := bisonfrontend.ToGrammar(
				bytes.NewBuffer(wellKnownGrammar.Content),
				wellKnownGrammar.FileName,
			)
			if err != nil {
				b.Fatal(err)
			}

			for b.Loop() {
				_, _, err := lr1bisoncore.GrammarToParser(grammar)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
