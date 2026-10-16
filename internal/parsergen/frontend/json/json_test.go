package json_test

import (
	"bytes"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/frontend/json"
	bisonfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/bison"
	"github.com/backbone81/golr/testdata"
)

func BenchmarkToGrammar(b *testing.B) {
	for _, wellKnownGrammar := range testdata.WellKnownGrammars {
		b.Run(wellKnownGrammar.Title, func(b *testing.B) {
			grammar, err := bisonfrontend.ToGrammar(
				bytes.NewBuffer(wellKnownGrammar.Content()),
				wellKnownGrammar.FileName,
			)
			if err != nil {
				b.Fatal(err)
			}

			var buffer bytes.Buffer
			if err := json.FromGrammar(&buffer, grammar); err != nil {
				b.Fatal(err)
			}

			for b.Loop() {
				if _, err := json.ToGrammar(bytes.NewReader(buffer.Bytes())); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

var _ = Describe("JSON", func() {
	It("returns an error for a symbol index out of bounds", func() {
		_, err := json.GrammarFromString(
			`{"terminals":[{"name":"a"}],"nonterminals":[{"name":"s"}],"productions":[` +
				`{"nonterminalIdx":0,"symbolRefs":[{"nonterminal":false,"index":-1}]}],"startNonterminalIdx":0}`,
		)
		Expect(err).To(HaveOccurred())
	})
})
