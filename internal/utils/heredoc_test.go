package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/utils"
)

var _ = Describe("HereDoc", func() {
	It("returns empty output for empty input", func() {
		Expect(utils.HereDoc("")).To(Equal(""))
	})

	It("drops the leading blank line", func() {
		Expect(utils.HereDoc("\nfoo")).To(Equal("foo"))
	})

	It("removes the common indentation shared by every line", func() {
		input := "\n    foo\n    bar\n"
		Expect(utils.HereDoc(input)).To(Equal("foo\nbar\n"))
	})

	It("keeps the relative indentation between lines", func() {
		input := "\n    foo\n        bar\n    baz\n"
		Expect(utils.HereDoc(input)).To(Equal("foo\n    bar\nbaz\n"))
	})

	It("ignores blank lines when computing the common indentation", func() {
		input := "\n    foo\n\n    bar\n"
		Expect(utils.HereDoc(input)).To(Equal("foo\n\nbar\n"))
	})

	It("reduces a blank line which only holds whitespace down to an empty line", func() {
		input := "\n    foo\n    \n    bar\n"
		Expect(utils.HereDoc(input)).To(Equal("foo\n\nbar\n"))
	})

	It("does not require a leading blank line", func() {
		Expect(utils.HereDoc("    foo\n    bar\n")).To(Equal("foo\nbar\n"))
	})

	It("leaves a single line without indentation unchanged", func() {
		Expect(utils.HereDoc("foo")).To(Equal("foo"))
	})

	It("falls back to no indentation when the lines do not share a common prefix", func() {
		input := "\n    foo\n\tbar\n"
		Expect(utils.HereDoc(input)).To(Equal("    foo\n\tbar\n"))
	})

	It("keeps a trailing line without a final newline as is", func() {
		input := "\n    foo\n    bar"
		Expect(utils.HereDoc(input)).To(Equal("foo\nbar"))
	})
})
