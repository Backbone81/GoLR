package convert_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/convert"
	"github.com/backbone81/golr/internal/parsergen/frontend/bison"
)

var _ = Describe("BisonGrammar2GoLR", func() {
	DescribeTable("refuses terminals which share an alias",
		func(source string) {
			grammar, err := bison.ToGrammar(strings.NewReader(source), "test.y")
			Expect(err).ToNot(HaveOccurred())
			_, err = convert.BisonGrammar2GoLR(grammar)
			Expect(err).To(MatchError(ContainSubstring("share the alias")))
		},
		Entry("declared with the same string", "%token X \"a\" Y \"a\"\n%%\ns: X Y;\n"),
		Entry("a character literal and a string", "%token PLUS \"+\"\n%%\ns: PLUS '+';\n"),
	)

	It("keeps distinct aliases", func() {
		grammar, err := bison.ToGrammar(strings.NewReader("%token X \"a\" Y \"b\"\n%%\ns: X Y '+';\n"), "test.y")
		Expect(err).ToNot(HaveOccurred())
		_, err = convert.BisonGrammar2GoLR(grammar)
		Expect(err).ToNot(HaveOccurred())
	})
})
