package golr_test

import (
	"bytes"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	lalr1golrcore "github.com/backbone81/golr/pkg/parsergen/core/lalr1/golr"
	bisonfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/bison"
	"github.com/backbone81/golr/testdata"
)

var _ = Describe("LALR(1)", func() {
	Context("well known grammars", func() {
		for _, wellKnownGrammar := range testdata.WellKnownGrammars {
			It("should correctly build the "+wellKnownGrammar.Title+" parser", func() {
				grammar, err := bisonfrontend.ToGrammar(
					bytes.NewBuffer(wellKnownGrammar.Content),
					wellKnownGrammar.FileName,
				)
				Expect(err).ToNot(HaveOccurred())

				Expect(lalr1golrcore.GrammarToParser(grammar)).Error().ToNot(HaveOccurred())
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
				if _, _, err := lalr1golrcore.GrammarToParser(grammar); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
