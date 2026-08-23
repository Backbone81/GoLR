package yaml_test

import (
	"io"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend/yaml"
	ielr1golrcore "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
)

var _ = Describe("FromParser", func() {
	for wellKnownGrammar, parser := range ielr1golrcore.WellKnownParsers() {
		It("emits a parser for the "+wellKnownGrammar.Title+" grammar", func() {
			Expect(yaml.FromParser(io.Discard, parser)).To(Succeed())
		})
	}
})

func BenchmarkFromParser(b *testing.B) {
	for wellKnownGrammar, parser := range ielr1golrcore.WellKnownParsers() {
		b.Run(wellKnownGrammar.Title, func(b *testing.B) {
			for b.Loop() {
				if err := yaml.FromParser(io.Discard, parser); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
