package bison_test

import (
	"bytes"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/frontend"
	ielr1bisoncore "github.com/backbone81/golr/pkg/parsergen/core/ielr1/bison"
	bisonfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/bison"
	"github.com/backbone81/golr/testdata"
)

var _ = Describe("IELR(1)", func() {
	Context("well known grammars", func() {
		for _, wellKnownGrammar := range testdata.WellKnownGrammars {
			It("should correctly build the "+wellKnownGrammar.Title+" parser", func() {
				grammar, err := bisonfrontend.ToGrammar(
					bytes.NewBuffer(wellKnownGrammar.Content),
					wellKnownGrammar.FileName,
				)
				Expect(err).ToNot(HaveOccurred())

				parser, _, err := ielr1bisoncore.GrammarToParser(grammar)
				Expect(err).ToNot(HaveOccurred())

				// The parser carries the numbering of the augmented grammar and not the one GNU Bison reports.
				// The two differ where GNU Bison predefines a symbol GoLR does not have, and where it moves the
				// nonterminals it considers useless to the end. This is what the GoLR cores build as well, so a
				// caller can switch cores without every symbol and production index moving underneath them.
				Expect(parser.Grammar).To(Equal(frontend.AugmentGrammar(grammar)))
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
				_, _, err := ielr1bisoncore.GrammarToParser(grammar)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
