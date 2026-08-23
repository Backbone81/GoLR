package cpp_test

import (
	"io"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend/cpp"
	ielr1golrcore "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
)

var _ = Describe("FromParser", func() {
	for wellKnownGrammar, parser := range ielr1golrcore.WellKnownParsers() {
		It("emits a parser for the "+wellKnownGrammar.Title+" grammar", func() {
			Expect(cpp.FromParser(io.Discard, parser, cpp.Config{
				Namespace:      "parser",
				ScannerInclude: cpp.DefaultScannerInclude,
			})).To(Succeed())
		})
	}
})

func BenchmarkFromParser(b *testing.B) {
	for wellKnownGrammar, parser := range ielr1golrcore.WellKnownParsers() {
		b.Run(wellKnownGrammar.Title, func(b *testing.B) {
			for b.Loop() {
				if err := cpp.FromParser(io.Discard, parser, cpp.Config{
					Namespace:      "parser",
					ScannerInclude: cpp.DefaultScannerInclude,
				}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
