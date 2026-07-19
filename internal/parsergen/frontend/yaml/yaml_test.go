package yaml_test

import (
	"bytes"
	"testing"

	"github.com/backbone81/golr/internal/parsergen/frontend/yaml"
	bisonfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/bison"
	"github.com/backbone81/golr/testdata"
)

func BenchmarkToGrammar(b *testing.B) {
	for _, wellKnownGrammar := range testdata.WellKnownGrammars {
		b.Run(wellKnownGrammar.Title, func(b *testing.B) {
			grammar, err := bisonfrontend.ToGrammar(
				bytes.NewBuffer(wellKnownGrammar.Content),
				wellKnownGrammar.FileName,
			)
			if err != nil {
				b.Fatal(err)
			}

			var buffer bytes.Buffer
			if err := yaml.FromGrammar(&buffer, grammar); err != nil {
				b.Fatal(err)
			}

			for b.Loop() {
				if _, err := yaml.ToGrammar(bytes.NewReader(buffer.Bytes())); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
