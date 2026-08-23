package golr

import (
	"bytes"
	"iter"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	bisonfrontend "github.com/backbone81/golr/internal/parsergen/frontend/bison"
	"github.com/backbone81/golr/testdata"
)

// WellKnownParsers provides the parser this core builds for every well-known grammar.
//
// Tests and benchmarks of everything downstream of a core need a real parse table to work on, and building one takes a
// frontend and a core call each time. Panics if a well-known grammar does not yield a parser, which is a programming
// error rather than a condition a caller could handle.
func WellKnownParsers() iter.Seq2[testdata.WellKnownGrammar, backend.Parser] {
	return func(yield func(testdata.WellKnownGrammar, backend.Parser) bool) {
		for _, wellKnownGrammar := range testdata.WellKnownGrammars {
			grammar, err := bisonfrontend.ToGrammar(
				bytes.NewBuffer(wellKnownGrammar.Content()),
				wellKnownGrammar.FileName,
			)
			if err != nil {
				panic(err)
			}

			parser, _, err := GrammarToParser(grammar, conflict.DefaultPolicy)
			if err != nil {
				panic(err)
			}

			if !yield(wellKnownGrammar, parser) {
				return
			}
		}
	}
}
