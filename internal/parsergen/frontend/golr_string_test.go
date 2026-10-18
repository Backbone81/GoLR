package frontend_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/frontend"
)

var _ = Describe("GoLR string literals", func() {
	DescribeTable("quote and unquote a value",
		func(value string, literal string) {
			Expect(frontend.QuoteGoLRString(value)).To(Equal(literal))
			Expect(frontend.UnquoteGoLRString(literal)).To(Equal(value))
		},
		Entry("plain", "+", `"+"`),
		Entry("a double quote", `"`, `"\""`),
		Entry("a backslash", `\`, `"\\"`),
		Entry("a backslash before a double quote", `\"`, `"\\\""`),
		Entry("a tab as is", "\t", "\"\t\""),
	)

	It("refuses to quote a line feed", func() {
		Expect(frontend.QuoteGoLRString("\n")).Error().To(HaveOccurred())
	})
})
