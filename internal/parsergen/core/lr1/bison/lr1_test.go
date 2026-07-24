package bison_test

import (
	"bytes"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	lr1bisoncore "github.com/backbone81/golr/pkg/parsergen/core/lr1/bison"
	bisonfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/bison"
	"github.com/backbone81/golr/testdata"
)

// Disabled because LR(1) grammars either fail or take a long time to construct.
var _ = PDescribe("LR(1)", func() {
	Context("well known grammars", func() {
		for _, wellKnownGrammar := range testdata.WellKnownGrammars {
			It("should correctly build the "+wellKnownGrammar.Title+" parser", func() {
				grammar, err := bisonfrontend.ToGrammar(
					bytes.NewBuffer(wellKnownGrammar.Content),
					wellKnownGrammar.FileName,
				)
				Expect(err).ToNot(HaveOccurred())

				Expect(lr1bisoncore.GrammarToParser(grammar)).Error().ToNot(HaveOccurred())
			})
		}
	})
})

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
