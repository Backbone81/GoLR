package model_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/model"
)

var _ = Describe("Model", func() {
	It("finds the occurrence at an offset", func() {
		m := model.Build([]byte(grammar))
		number := span(grammar, "NUMBER", 1)

		occurrence, found := m.OccurrenceAt(number.Offset)
		Expect(found).To(BeTrue())
		Expect(occurrence).To(Equal(model.Occurrence{Name: "NUMBER", Role: model.RoleReference, Span: number}))

		occurrence, found = m.OccurrenceAt(number.Offset + 2)
		Expect(found).To(BeTrue())
		Expect(occurrence.Span).To(Equal(number))

		occurrence, found = m.OccurrenceAt(number.End())
		Expect(found).To(BeTrue())
		Expect(occurrence.Span).To(Equal(number))

		_, found = m.OccurrenceAt(number.Offset - 1)
		Expect(found).To(BeFalse())
	})

	It("finds the definitions and aliases at their offset", func() {
		m := model.Build([]byte(grammar))
		occurrence, found := m.OccurrenceAt(span(grammar, "PLUS", 0).Offset)
		Expect(found).To(BeTrue())
		Expect(occurrence.Role).To(Equal(model.RoleDefinition))

		occurrence, found = m.OccurrenceAt(span(grammar, `"+"`, 0).Offset)
		Expect(found).To(BeTrue())
		Expect(occurrence.Role).To(Equal(model.RoleAlias))
	})

	It("prefers the occurrence starting at an offset over the one ending there", func() {
		text := `@parser { s: A"b"; }`
		occurrence, found := model.Build([]byte(text)).OccurrenceAt(span(text, `"b"`, 0).Offset)
		Expect(found).To(BeTrue())
		Expect(occurrence.Name).To(Equal(`"b"`))
	})

	It("finds nothing past the last occurrence", func() {
		_, found := model.Build([]byte(grammar)).OccurrenceAt(len(grammar))
		Expect(found).To(BeFalse())
	})

	It("finds the definitions of a name", func() {
		m := model.Build([]byte(grammar))
		Expect(m.DefinitionsNamed("expr")).To(Equal([]model.Definition{
			{Name: "expr", Kind: model.KindNonterminal, Span: span(grammar, "expr", 1)},
		}))
		Expect(m.DefinitionsNamed("UMINUS")).To(BeEmpty())
	})

	It("resolves a string to its alias and a name to its definition", func() {
		m := model.Build([]byte(grammar))
		Expect(m.DeclarationsNamed(`"+"`)).To(Equal([]model.Occurrence{
			{Name: `"+"`, Role: model.RoleAlias, Span: span(grammar, `"+"`, 0)},
		}))
		Expect(m.DeclarationsNamed("PLUS")).To(Equal([]model.Occurrence{
			{Name: "PLUS", Role: model.RoleDefinition, Span: span(grammar, "PLUS", 0)},
		}))
		Expect(m.DeclarationsNamed("UMINUS")).To(BeEmpty())
		Expect(m.DeclarationsNamed(`"*"`)).To(BeEmpty())
	})

	It("resolves a name declared twice to both declarations", func() {
		text := `@scanner { A: "a"; A: "a"; } @parser { s: A "a"; }`
		m := model.Build([]byte(text))
		Expect(m.DeclarationsNamed("A")).To(HaveLen(2))
		Expect(m.DeclarationsNamed(`"a"`)).To(HaveLen(2))
	})

	It("finds the references of a name", func() {
		m := model.Build([]byte(grammar))
		Expect(m.ReferencesNamed("UMINUS")).To(Equal([]model.Reference{
			{Name: "UMINUS", Span: span(grammar, "UMINUS", 0)},
			{Name: "UMINUS", Span: span(grammar, "UMINUS", 1)},
		}))
		Expect(m.ReferencesNamed("PLUS")).To(Equal([]model.Reference{{Name: "PLUS", Span: span(grammar, "PLUS", 1)}}))
		Expect(m.ReferencesNamed("empty")).To(BeEmpty())
	})

	It("finds the usages of a terminal by its name and its alias in the order of the text", func() {
		m := model.Build([]byte(grammar))
		Expect(m.UsagesNamed("MINUS")).To(Equal([]model.Reference{
			{Name: `"-"`, Span: span(grammar, `"-"`, 1)},
			{Name: "MINUS", Span: span(grammar, "MINUS", 1)},
			{Name: `"-"`, Span: span(grammar, `"-"`, 2)},
		}))
	})

	It("finds the usages of an alias by its string only", func() {
		m := model.Build([]byte(grammar))
		Expect(m.UsagesNamed(`"+"`)).To(Equal([]model.Reference{{Name: `"+"`, Span: span(grammar, `"+"`, 1)}}))
	})

	It("counts the usages by an alias once when a terminal is declared twice with it", func() {
		text := `@scanner { A: "a"; A: "a"; } @parser { s: A "a"; }`
		Expect(model.Build([]byte(text)).UsagesNamed("A")).To(Equal([]model.Reference{
			{Name: "A", Span: span(text, "A", 2)},
			{Name: `"a"`, Span: span(text, `"a"`, 2)},
		}))
	})

	It("finds the usages of a terminal declared twice with different strings", func() {
		text := `@scanner { A: "a"; A: "b"; } @parser { s: "b" A "a"; }`
		Expect(referenceNames(model.Build([]byte(text)).UsagesNamed("A"))).To(Equal([]string{`"b"`, "A", `"a"`}))
	})

	Context("CompletionAt", func() {
		topLevel := []string{"@scanner", "@parser"}
		scannerRuleBody := []string{"@empty", "@skip", "@fragment"}
		parserStatementStart := []string{"@start", "@precedence"}
		precedenceLineStart := []string{"@left", "@right", "@none", "@precedence"}
		precedenceLineBody := []string{"@error"}
		productionBody := []string{"@empty", "@error", "@precedence", "@name"}

		DescribeTable("offers the keywords and symbols of the place in front of the word",
			func(text string, keywords []string, symbols bool) {
				text, offset := withoutCaret(text)
				completion := model.Build([]byte(text)).CompletionAt(offset)
				Expect(completion.Keywords).To(Equal(keywords))
				Expect(completion.Symbols).To(Equal(symbols))
			},
			Entry("empty text", "‸", topLevel, false),
			Entry("keyword at the top level", "@sc‸", topLevel, false),
			Entry("after a section keyword", "@scanner ‸", nil, false),
			Entry("name of a scanner rule", "@scanner { ‸", nil, false),
			Entry("name of a scanner rule being typed", "@scanner { AB‸", nil, false),
			Entry("before the colon of a scanner rule", "@scanner { A ‸", nil, false),
			Entry("scanner rule body", "@scanner { A: ‸", scannerRuleBody, false),
			Entry("annotation of a scanner rule", "@scanner { A: /a/ @‸", scannerRuleBody, false),
			Entry("after a scanner rule", `@scanner { A: "a"; ‸`, nil, false),
			Entry("after the scanner section", `@scanner { A: "a"; } ‸`, topLevel, false),
			Entry("parser statement start", scannerSection+"@parser { ‸", parserStatementStart, false),
			Entry("start body", scannerSection+"@parser { @start: ‸", nil, true),
			Entry("start name being typed", scannerSection+"@parser { @start: ex‸", nil, true),
			Entry("after the start name", scannerSection+"@parser { @start: s ‸", nil, false),
			Entry("after the start declaration", scannerSection+"@parser { @start: s; ‸", parserStatementStart, false),
			Entry("precedence line start", scannerSection+"@parser { @precedence { ‸", precedenceLineStart, false),
			Entry("precedence line body", scannerSection+"@parser { @precedence { @left: ‸", precedenceLineBody, true),
			Entry("after a symbol of a precedence line", scannerSection+"@parser { @precedence { @left: A ‸",
				precedenceLineBody, true),
			Entry("after a precedence line", scannerSection+"@parser { @precedence { @left: A; ‸",
				precedenceLineStart, false),
			Entry("after the precedence block", scannerSection+"@parser { @precedence { @left: A; } ‸",
				parserStatementStart, false),
			Entry("name of a parser rule being typed", scannerSection+"@parser { ex‸", parserStatementStart, false),
			Entry("before the colon of a parser rule", scannerSection+"@parser { s ‸", nil, false),
			Entry("production body", scannerSection+"@parser { s: ‸", productionBody, true),
			Entry("symbol being typed", scannerSection+"@parser { s: A B‸", productionBody, true),
			Entry("after an alternative", scannerSection+"@parser { s: A | ‸", productionBody, true),
			Entry("keyword in a production body", scannerSection+"@parser { s: A @‸", productionBody, false),
			Entry("half typed keyword in a production body", scannerSection+"@parser { s: A @na‸", productionBody, false),
			Entry("after @precedence", scannerSection+"@parser { s: A @precedence ‸", nil, false),
			Entry("precedence argument", scannerSection+"@parser { s: A @precedence(‸", nil, true),
			Entry("precedence argument being typed", scannerSection+"@parser { s: A @precedence(U‸", nil, true),
			Entry("after the precedence argument", scannerSection+"@parser { s: A @precedence(U ‸", nil, false),
			Entry("after the precedence annotation", scannerSection+"@parser { s: A @precedence(U) ‸",
				productionBody, true),
			Entry("inside @name(...)", scannerSection+"@parser { s: A @name(‸", nil, false),
			Entry("after the name annotation", scannerSection+"@parser { s: A @name(n) ‸", productionBody, true),
			Entry("after a parser rule", scannerSection+"@parser { s: A; ‸", parserStatementStart, false),
			Entry("body of a rule after a missing semicolon", scannerSection+"@parser { s: A t: ‸", productionBody, true),
			Entry("parser section after a missing closing brace", `@scanner { A: "a"; @parser { ‸`,
				parserStatementStart, false),
			Entry("in front of a string", scannerSection+`@parser { s: A ‸"a"; }`, productionBody, true),
			Entry("inside a string", scannerSection+`@parser { s: A "a‸"; }`, nil, false),
			Entry("after a string", scannerSection+`@parser { s: A "a"‸; }`, productionBody, true),
			Entry("inside an unterminated string", scannerSection+`@parser { s: A "ab‸`, nil, false),
			Entry("at the end of the line of an unterminated string", scannerSection+"@parser { s: A \"ab‸\n }",
				nil, false),
			Entry("on the line after an unterminated string", scannerSection+"@parser { s: A \"ab\n‸ }",
				productionBody, true),
			Entry("inside a regular expression", "@scanner { A: /a‸/; }", nil, false),
			Entry("inside an unterminated regular expression", "@scanner { A: /ab‸\n }", nil, false),
			Entry("inside a line comment", scannerSection+"@parser { s: A // x‸\n }", nil, false),
			Entry("on the line after a line comment", scannerSection+"@parser { s: A // x\n‸ }", productionBody, true),
			Entry("inside a line comment at the end of the text", scannerSection+"@parser { s: A // x‸", nil, false),
			Entry("inside a block comment", scannerSection+"@parser { s: A /* x ‸ */ }", nil, false),
			Entry("after a block comment", scannerSection+"@parser { s: A /* x */ ‸ }", productionBody, true),
			Entry("inside an unterminated block comment", scannerSection+"@parser { s: A /* x\n‸", nil, false),
			Entry("inside a comment at the top level", "/* x ‸ */", nil, false),
			Entry("after a comment at the top level", "/* x */ ‸", topLevel, false),
		)

		It("starts the word at the identifier before the caret", func() {
			text, offset := withoutCaret(scannerSection + "@parser { s: A expr‸ }")
			Expect(model.Build([]byte(text)).CompletionAt(offset).Offset).To(Equal(strings.Index(text, "expr")))
		})

		It("starts the word at the leading @", func() {
			text, offset := withoutCaret(scannerSection + "@parser { s: A @na‸ }")
			Expect(model.Build([]byte(text)).CompletionAt(offset).Offset).To(Equal(strings.Index(text, "@na")))
		})

		It("starts the word at the caret when no identifier is in front of it", func() {
			text, offset := withoutCaret(scannerSection + "@parser { s: A ‸ }")
			Expect(model.Build([]byte(text)).CompletionAt(offset).Offset).To(Equal(offset))
		})

		It("counts only ASCII characters as part of an identifier", func() {
			text, offset := withoutCaret(scannerSection + "@parser { s: Aä‸ }")
			Expect(model.Build([]byte(text)).CompletionAt(offset).Offset).To(Equal(offset))
		})

		It("clamps an offset outside the text", func() {
			m := model.Build([]byte(scannerSection))
			Expect(m.CompletionAt(-1).Keywords).To(Equal(topLevel))
			Expect(m.CompletionAt(1000).Offset).To(Equal(len(scannerSection)))
		})
	})
})
