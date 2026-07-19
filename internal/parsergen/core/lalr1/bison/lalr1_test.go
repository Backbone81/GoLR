package bison_test

import (
	"bytes"
	"testing"

	lalr1bisoncore "github.com/backbone81/golr/pkg/parsergen/core/lalr1/bison"
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
				if _, _, err := lalr1bisoncore.GrammarToParser(grammar); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
