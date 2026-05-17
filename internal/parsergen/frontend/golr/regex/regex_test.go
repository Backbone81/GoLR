package regex_test

import (
	"math"

	"github.com/backbone81/golr/internal/parsergen/frontend/golr/regex"
	"github.com/backbone81/golr/internal/scannergen/frontend/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Regular Expressions", func() {
	It("should correctly parse /./", func() {
		node, err := regex.Parse([]byte(`/./`))
		Expect(err).ToNot(HaveOccurred())

		Expect(node).To(Equal(dsl.Any()))
	})

	It("should correctly parse /a/", func() {
		node, err := regex.Parse([]byte(`/a/`))
		Expect(err).ToNot(HaveOccurred())

		Expect(node).To(Equal(dsl.Literal("a")))
	})

	It("should correctly parse /ab/", func() {
		node, err := regex.Parse([]byte(`/ab/`))
		Expect(err).ToNot(HaveOccurred())

		Expect(node).To(Equal(dsl.Literal("ab")))
	})

	It("should correctly parse /a|b/", func() {
		node, err := regex.Parse([]byte(`/a|b/`))
		Expect(err).ToNot(HaveOccurred())

		Expect(node).To(Equal(dsl.Or(
			dsl.Literal("a"),
			dsl.Literal("b"),
		)))
	})

	It(`should correctly parse /a\*/`, func() {
		node, err := regex.Parse([]byte(`/a\*/`))
		Expect(err).ToNot(HaveOccurred())

		Expect(node).To(Equal(dsl.Literal("a*")))
	})

	Context("Quantifiers", func() {
		It("should correctly parse /a*/", func() {
			node, err := regex.Parse([]byte(`/a*/`))
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.ZeroOrMore(
				dsl.Literal("a"),
			)))
		})

		It("should correctly parse /a+/", func() {
			node, err := regex.Parse([]byte(`/a+/`))
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.OneOrMore(
				dsl.Literal("a"),
			)))
		})

		It("should correctly parse /a?/", func() {
			node, err := regex.Parse([]byte(`/a?/`))
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.Optional(
				dsl.Literal("a"),
			)))
		})

		It("should correctly parse /a{2}/", func() {
			node, err := regex.Parse([]byte(`/a{2}/`))
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.Repetition(
				dsl.Literal("a"),
				2, 2,
			)))
		})

		It("should correctly parse /a{2,5}/", func() {
			node, err := regex.Parse([]byte(`/a{2,5}/`))
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.Repetition(
				dsl.Literal("a"),
				2, 5,
			)))
		})

		It("should correctly parse /a{2,}/", func() {
			node, err := regex.Parse([]byte(`/a{2,}/`))
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.Repetition(
				dsl.Literal("a"),
				2, math.MaxInt,
			)))
		})

		It("should correctly parse /a{,5}/", func() {
			node, err := regex.Parse([]byte(`/a{,5}/`))
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.Repetition(
				dsl.Literal("a"),
				0, 5,
			)))
		})
	})

	Context("Character Class", func() {
		It("should correctly parse /[ab-d]/", func() {
			node, err := regex.Parse([]byte(`/[ab-d]/`))
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.CharClass(
				dsl.CharRange('a', 'a'),
				dsl.CharRange('b', 'd'),
			)))
		})

		It("should correctly parse /[^ab-d]/", func() {
			node, err := regex.Parse([]byte(`/[^ab-d]/`))
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.NegCharClass(
				dsl.CharRange('a', 'a'),
				dsl.CharRange('b', 'd'),
			)))
		})

		It("should correctly parse /[-ab]/", func() {
			node, err := regex.Parse([]byte(`/[-ab]/`))
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.CharClass(
				dsl.CharRange('-', '-'),
				dsl.CharRange('a', 'a'),
				dsl.CharRange('b', 'b'),
			)))
		})

		It("should correctly parse /[^-ab]/", func() {
			node, err := regex.Parse([]byte(`/[^-ab]/`))
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.NegCharClass(
				dsl.CharRange('-', '-'),
				dsl.CharRange('a', 'a'),
				dsl.CharRange('b', 'b'),
			)))
		})

		It("should correctly parse /[ab-]/", func() {
			node, err := regex.Parse([]byte(`/[ab-]/`))
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.CharClass(
				dsl.CharRange('a', 'a'),
				dsl.CharRange('b', 'b'),
				dsl.CharRange('-', '-'),
			)))
		})

		It("should correctly parse /[^ab-]/", func() {
			node, err := regex.Parse([]byte(`/[^ab-]/`))
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.NegCharClass(
				dsl.CharRange('a', 'a'),
				dsl.CharRange('b', 'b'),
				dsl.CharRange('-', '-'),
			)))
		})

		It(`should correctly parse /[a\-b]/`, func() {
			node, err := regex.Parse([]byte(`/[a\-b]/`))
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.NegCharClass(
				dsl.CharRange('a', 'a'),
				dsl.CharRange('-', '-'),
				dsl.CharRange('b', 'b'),
			)))
		})

		It(`should correctly parse /[ab\]]/`, func() {
			node, err := regex.Parse([]byte(`/[a\-b]/`))
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.NegCharClass(
				dsl.CharRange('a', 'a'),
				dsl.CharRange('b', 'b'),
				dsl.CharRange(']', ']'),
			)))
		})
	})

	Context("Special Characters", func() {
		It(`should correctly parse /\n/`, func() {
			node, err := regex.Parse([]byte(`/\n/`))
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.Literal("\n")))
		})

		It(`should correctly parse /\r/`, func() {
			node, err := regex.Parse([]byte(`/\r/`))
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.Literal("\r")))
		})

		It(`should correctly parse /\t/`, func() {
			node, err := regex.Parse([]byte(`/\t/`))
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.Literal("\t")))
		})
	})

	Context("Grouping", func() {
		It(`should correctly parse /(abc)*/`, func() {
			node, err := regex.Parse([]byte(`/(abc)*/`))
			Expect(err).ToNot(HaveOccurred())

			Expect(node).To(Equal(dsl.ZeroOrMore(dsl.Literal("abc"))))
		})
	})
})
