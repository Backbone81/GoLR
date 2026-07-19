package yaml_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/backbone81/golr/internal/parsergen/backend/yaml"
	ielr1bisoncore "github.com/backbone81/golr/pkg/parsergen/core/ielr1/bison"
	bisonfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/bison"
	"github.com/backbone81/golr/testdata"
)

func BenchmarkFromParser(b *testing.B) {
	for _, wellKnownGrammar := range testdata.WellKnownGrammars {
		b.Run(wellKnownGrammar.Title, func(b *testing.B) {
			grammar, err := bisonfrontend.ToGrammar(
				bytes.NewBuffer(wellKnownGrammar.Content),
				wellKnownGrammar.FileName,
			)
			if err != nil {
				b.Fatal(err)
			}

			parser, _, err := ielr1bisoncore.GrammarToParser(grammar)
			if err != nil {
				b.Fatal(err)
			}

			for b.Loop() {
				if err := yaml.FromParser(io.Discard, parser); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
