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

	It("names string literals uniquely when their names come out the same", func() {
		grammar, err := bison.ToGrammar(strings.NewReader("%%\ns: \"_\" \" \";\n"), "test.y")
		Expect(err).ToNot(HaveOccurred())
		grammar, err = convert.BisonGrammar2GoLR(grammar)
		Expect(err).ToNot(HaveOccurred())

		var names []string
		for _, terminal := range grammar.Terminals {
			names = append(names, terminal.Name)
		}
		Expect(names).To(ContainElements("STRING__", "STRING___1"))
	})

	DescribeTable("translates the escape sequences of Bison literals to GoLR strings",
		func(source string, alias string) {
			grammar, err := bison.ToGrammar(strings.NewReader(source), "test.y")
			Expect(err).ToNot(HaveOccurred())
			grammar, err = convert.BisonGrammar2GoLR(grammar)
			Expect(err).ToNot(HaveOccurred())

			var aliases []string
			for _, terminal := range grammar.Terminals {
				aliases = append(aliases, terminal.Alias)
			}
			Expect(aliases).To(ContainElement(alias))
		},
		Entry("a backslash string", "%%\ns: \"\\\\\";\n", `"\\"`),
		Entry("a double quote string", "%%\ns: \"\\\"\";\n", `"\""`),
		Entry("a single quote character", "%%\ns: '\\'';\n", `"'"`),
		Entry("a hexadecimal escape", "%%\ns: \"\\x41\";\n", `"A"`),
		Entry("an octal escape", "%%\ns: '\\101';\n", `"A"`),
		Entry("a declared string", "%token X \"a\\\\b\"\n%%\ns: X;\n", `"a\\b"`),
	)

	It("falls back to @empty for a literal with a line feed", func() {
		grammar, err := bison.ToGrammar(strings.NewReader("%%\ns: '\\n';\n"), "test.y")
		Expect(err).ToNot(HaveOccurred())
		grammar, err = convert.BisonGrammar2GoLR(grammar)
		Expect(err).ToNot(HaveOccurred())
		for _, terminal := range grammar.Terminals {
			Expect(terminal.Alias).To(BeEmpty())
		}
	})
})
