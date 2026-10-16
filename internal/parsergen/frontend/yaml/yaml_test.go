package yaml_test

import (
	"bytes"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/frontend/yaml"
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

var _ = Describe("YAML", func() {
	It("returns an error for a symbol index out of bounds", func() {
		_, err := yaml.GrammarFromString(
			"terminals: [{name: a}]\nnonterminals: [{name: s}]\nproductions:\n" +
				"- {nonterminalIdx: 0, symbolRefs: [{nonterminal: false, index: -1}]}\nstartNonterminalIdx: 0\n",
		)
		Expect(err).To(HaveOccurred())
	})
})
