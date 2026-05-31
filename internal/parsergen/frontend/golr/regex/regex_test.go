package regex_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/frontend/golr/regex"
	"github.com/backbone81/golr/internal/scannergen/frontend/dsl"
)

var _ = Describe("Regular Expressions", func() {
	It("should correctly parse /./", func() {
		node, err := regex.Parse([]byte(`/./`), nil)
		Expect(err).ToNot(HaveOccurred())

		Expect(node).To(Equal(dsl.Any()))
	})

	It("should correctly parse /a/", func() {
		node, err := regex.Parse([]byte(`/a/`), nil)
		Expect(err).ToNot(HaveOccurred())

		Expect(node).To(Equal(dsl.Literal("a")))
	})

	It("should correctly parse /ab/", func() {
		node, err := regex.Parse([]byte(`/ab/`), nil)
		Expect(err).ToNot(HaveOccurred())

		Expect(node).To(Equal(dsl.Literal("ab")))
	})

	It("should correctly parse /a|b/", func() {
		node, err := regex.Parse([]byte(`/a|b/`), nil)
		Expect(err).ToNot(HaveOccurred())

		Expect(node).To(Equal(dsl.Or(
			dsl.Literal("a"),
			dsl.Literal("b"),
		)))
	})

	It(`should correctly parse /a\*/`, func() {
		node, err := regex.Parse([]byte(`/a\*/`), nil)
		Expect(err).ToNot(HaveOccurred())

		Expect(node).To(Equal(dsl.Literal("a*")))
	})

	Context("Quantifiers", func() {
		It("should correctly parse /a*/", func() {
			node, err := regex.Parse([]byte(`/a*/`), nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.ZeroOrMore(
				dsl.Literal("a"),
			)))
		})

		It("should correctly parse /a+/", func() {
			node, err := regex.Parse([]byte(`/a+/`), nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.OneOrMore(
				dsl.Literal("a"),
			)))
		})

		It("should correctly parse /a?/", func() {
			node, err := regex.Parse([]byte(`/a?/`), nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.Optional(
				dsl.Literal("a"),
			)))
		})

		It("should correctly parse /a{2}/", func() {
			node, err := regex.Parse([]byte(`/a{2}/`), nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.Repetition(
				dsl.Literal("a"),
				2, 2,
			)))
		})

		It("should correctly parse /a{2,5}/", func() {
			node, err := regex.Parse([]byte(`/a{2,5}/`), nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.Repetition(
				dsl.Literal("a"),
				2, 5,
			)))
		})

		It("should correctly parse /a{2,}/", func() {
			node, err := regex.Parse([]byte(`/a{2,}/`), nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.Concat(
				dsl.Repetition(
					dsl.Literal("a"),
					2, 2,
				),
				dsl.ZeroOrMore(dsl.Literal("a")),
			)))
		})

		It("should correctly parse /a{,5}/", func() {
			node, err := regex.Parse([]byte(`/a{,5}/`), nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.Repetition(
				dsl.Literal("a"),
				0, 5,
			)))
		})
	})

	Context("Character Class", func() {
		It("should correctly parse /[ab-d]/", func() {
			node, err := regex.Parse([]byte(`/[ab-d]/`), nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.CharClass(
				dsl.CharRange('a', 'a'),
				dsl.CharRange('b', 'd'),
			)))
		})

		It("should correctly parse /[^ab-d]/", func() {
			node, err := regex.Parse([]byte(`/[^ab-d]/`), nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.NegCharClass(
				dsl.CharRange('a', 'a'),
				dsl.CharRange('b', 'd'),
			)))
		})

		It("should correctly parse /[-ab]/", func() {
			node, err := regex.Parse([]byte(`/[-ab]/`), nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.CharClass(
				dsl.CharRange('-', '-'),
				dsl.CharRange('a', 'a'),
				dsl.CharRange('b', 'b'),
			)))
		})

		It("should correctly parse /[^-ab]/", func() {
			node, err := regex.Parse([]byte(`/[^-ab]/`), nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.NegCharClass(
				dsl.CharRange('-', '-'),
				dsl.CharRange('a', 'a'),
				dsl.CharRange('b', 'b'),
			)))
		})

		It("should correctly parse /[ab-]/", func() {
			node, err := regex.Parse([]byte(`/[ab-]/`), nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.CharClass(
				dsl.CharRange('a', 'a'),
				dsl.CharRange('b', 'b'),
				dsl.CharRange('-', '-'),
			)))
		})

		It("should correctly parse /[^ab-]/", func() {
			node, err := regex.Parse([]byte(`/[^ab-]/`), nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.NegCharClass(
				dsl.CharRange('a', 'a'),
				dsl.CharRange('b', 'b'),
				dsl.CharRange('-', '-'),
			)))
		})

		It(`should correctly parse /[a\-b]/`, func() {
			node, err := regex.Parse([]byte(`/[a\-b]/`), nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.CharClass(
				dsl.CharRange('a', 'a'),
				dsl.CharRange('-', '-'),
				dsl.CharRange('b', 'b'),
			)))
		})

		It(`should correctly parse /[ab\]]/`, func() {
			node, err := regex.Parse([]byte(`/[ab\]]/`), nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.CharClass(
				dsl.CharRange('a', 'a'),
				dsl.CharRange('b', 'b'),
				dsl.CharRange(']', ']'),
			)))
		})
	})

	Context("POSIX Character Classes", func() {
		It("should correctly parse /[[:alpha:]]/", func() {
			_, err := regex.Parse([]byte(`/[[:alpha:]]/`), nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should correctly parse /[^[:alpha:]]/ as a negated class", func() {
			_, err := regex.Parse([]byte(`/[^[:alpha:]]/`), nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should correctly parse /[a[:digit:]]/ as a mixed class", func() {
			_, err := regex.Parse([]byte(`/[a[:digit:]]/`), nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return an error for an unknown POSIX class", func() {
			_, err := regex.Parse([]byte(`/[[:unknown:]]/`), nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown"))
		})
	})

	Context("Special Characters", func() {
		It(`should correctly parse /\n/`, func() {
			node, err := regex.Parse([]byte(`/\n/`), nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.Literal("\n")))
		})

		It(`should correctly parse /\r/`, func() {
			node, err := regex.Parse([]byte(`/\r/`), nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.Literal("\r")))
		})

		It(`should correctly parse /\t/`, func() {
			node, err := regex.Parse([]byte(`/\t/`), nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.Literal("\t")))
		})
	})

	Context("Grouping", func() {
		It(`should correctly parse /(abc)*/`, func() {
			node, err := regex.Parse([]byte(`/(abc)*/`), nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.ZeroOrMore(dsl.Literal("abc"))))
		})
	})

	Context("Fragments", func() {
		It("should correctly parse /{DIGIT}/ as a standalone atom", func() {
			fragments := map[string][]byte{
				"DIGIT": []byte(`/[0-9]/`),
			}
			node, err := regex.Parse([]byte(`/{DIGIT}/`), fragments)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.CharClass(
				dsl.CharRange('0', '9'),
			)))
		})

		It("should correctly parse /a{DIGIT}/ with the quantifier early exit", func() {
			fragments := map[string][]byte{
				"DIGIT": []byte(`/[0-9]/`),
			}
			node, err := regex.Parse([]byte(`/a{DIGIT}/`), fragments)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.Concat(
				dsl.Literal("a"),
				dsl.CharClass(dsl.CharRange('0', '9')),
			)))
		})

		It("should correctly parse /{DIGIT}+/", func() {
			fragments := map[string][]byte{
				"DIGIT": []byte(`/[0-9]/`),
			}
			node, err := regex.Parse([]byte(`/{DIGIT}+/`), fragments)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.OneOrMore(
				dsl.CharClass(dsl.CharRange('0', '9')),
			)))
		})

		It("should correctly parse /{DIGIT}|{ALPHA}/", func() {
			fragments := map[string][]byte{
				"DIGIT": []byte(`/[0-9]/`),
				"ALPHA": []byte(`/[a-zA-Z]/`),
			}
			node, err := regex.Parse([]byte(`/{DIGIT}|{ALPHA}/`), fragments)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.Or(
				dsl.CharClass(dsl.CharRange('0', '9')),
				dsl.CharClass(
					dsl.CharRange('a', 'z'),
					dsl.CharRange('A', 'Z'),
				),
			)))
		})

		It("should correctly parse a string-literal fragment /{KWORD}/", func() {
			fragments := map[string][]byte{
				"KWORD": []byte(`"keyword"`),
				"EMPTY": nil,
			}
			node, err := regex.Parse([]byte(`/{KWORD}/`), fragments)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.Literal("keyword")))
		})

		It("should correctly parse an empty fragment /{EMPTY}/", func() {
			fragments := map[string][]byte{
				"EMPTY": nil,
			}
			node, err := regex.Parse([]byte(`/{EMPTY}/`), fragments)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.CharClass()))
		})

		It("should correctly parse a nested fragment /{HEX}/", func() {
			fragments := map[string][]byte{
				"DIGIT": []byte(`/[0-9]/`),
				"HEX":   []byte(`/[a-f]{DIGIT}/`),
			}
			node, err := regex.Parse([]byte(`/{HEX}/`), fragments)
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.Concat(
				dsl.CharClass(dsl.CharRange('a', 'f')),
				dsl.CharClass(dsl.CharRange('0', '9')),
			)))
		})

		It("should return an error for an unknown fragment", func() {
			_, err := regex.Parse([]byte(`/{UNKNOWN}/`), nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("UNKNOWN"))
		})

		It("should return an error for a cyclic fragment reference", func() {
			fragments := map[string][]byte{
				"A": []byte(`/{B}/`),
				"B": []byte(`/{A}/`),
			}
			_, err := regex.Parse([]byte(`/{A}/`), fragments)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cyclic"))
		})
	})
})
