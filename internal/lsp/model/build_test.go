package model_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/model"
)

var _ = Describe("Build", func() {
	It("finds the definitions with their kind", func() {
		m := model.Build([]byte(grammar))
		Expect(m.Definitions()).To(Equal([]model.Definition{
			{Name: "DIGIT", Kind: model.KindFragment, Span: span(grammar, "DIGIT", 0)},
			{Name: "WS", Kind: model.KindTerminal, Span: span(grammar, "WS", 0)},
			{Name: "PLUS", Kind: model.KindTerminal, Span: span(grammar, "PLUS", 0)},
			{Name: "MINUS", Kind: model.KindTerminal, Span: span(grammar, "MINUS", 0)},
			{Name: "NUMBER", Kind: model.KindTerminal, Span: span(grammar, "NUMBER", 0)},
			{Name: "expr", Kind: model.KindNonterminal, Span: span(grammar, "expr", 1)},
			{Name: "empty", Kind: model.KindNonterminal, Span: span(grammar, "empty", 0)},
		}))
	})

	It("finds the aliases of the terminals", func() {
		m := model.Build([]byte(grammar))
		Expect(m.Aliases()).To(Equal([]model.Alias{
			{Name: `"+"`, Terminal: "PLUS", Span: span(grammar, `"+"`, 0)},
			{Name: `"-"`, Terminal: "MINUS", Span: span(grammar, `"-"`, 0)},
		}))
	})

	It("finds the references in @start, precedence lines, production bodies and @precedence(...)", func() {
		m := model.Build([]byte(grammar))
		Expect(m.References()).To(Equal([]model.Reference{
			{Name: "expr", Span: span(grammar, "expr", 0)},
			{Name: "PLUS", Span: span(grammar, "PLUS", 1)},
			{Name: `"-"`, Span: span(grammar, `"-"`, 1)},
			{Name: "UMINUS", Span: span(grammar, "UMINUS", 0)},
			{Name: "expr", Span: span(grammar, "expr", 2)},
			{Name: `"+"`, Span: span(grammar, `"+"`, 1)},
			{Name: "expr", Span: span(grammar, "expr", 3)},
			{Name: "expr", Span: span(grammar, "expr", 4)},
			{Name: "MINUS", Span: span(grammar, "MINUS", 1)},
			{Name: "expr", Span: span(grammar, "expr", 5)},
			{Name: `"-"`, Span: span(grammar, `"-"`, 2)},
			{Name: "expr", Span: span(grammar, "expr", 6)},
			{Name: "UMINUS", Span: span(grammar, "UMINUS", 1)},
			{Name: "NUMBER", Span: span(grammar, "NUMBER", 1)},
		}))
	})

	It("keeps the occurrences sorted by offset", func() {
		m := model.Build([]byte(grammar))
		occurrences := m.Occurrences()
		Expect(occurrences).To(HaveLen(len(m.Definitions()) + len(m.Aliases()) + len(m.References())))
		for idx := 1; idx < len(occurrences); idx++ {
			Expect(occurrences[idx-1].Span.End()).To(BeNumerically("<=", occurrences[idx].Span.Offset))
		}
	})

	It("gives an empty model for an empty text", func() {
		m := model.Build(nil)
		Expect(m.Definitions()).To(BeEmpty())
		Expect(m.Aliases()).To(BeEmpty())
		Expect(m.References()).To(BeEmpty())
		Expect(m.Occurrences()).To(BeEmpty())
		_, found := m.OccurrenceAt(0)
		Expect(found).To(BeFalse())
	})

	It("takes only the string right after the colon as the alias", func() {
		text := `@scanner { A: "a" "b"; B: /b/ "c"; }`
		Expect(model.Build([]byte(text)).Aliases()).To(Equal([]model.Alias{
			{Name: `"a"`, Terminal: "A", Span: span(text, `"a"`, 0)},
		}))
	})

	It("starts the next scanner rule at an identifier", func() {
		text := `@scanner { A: "a" B: "b"; }`
		m := model.Build([]byte(text))
		Expect(definitionNames(m.Definitions())).To(Equal([]string{"A", "B"}))
		Expect(m.Aliases()).To(Equal([]model.Alias{
			{Name: `"a"`, Terminal: "A", Span: span(text, `"a"`, 0)},
			{Name: `"b"`, Terminal: "B", Span: span(text, `"b"`, 0)},
		}))
	})

	It("starts the next parser rule at an identifier followed by a colon", func() {
		text := `@scanner { A: "a"; } @parser { s: A t: s A; }`
		m := model.Build([]byte(text))
		Expect(definitionNames(m.Definitions())).To(Equal([]string{"A", "s", "t"}))
		Expect(referenceNames(m.References())).To(Equal([]string{"A", "s", "A"}))
	})

	It("gives a fragment no alias", func() {
		text := `@scanner { A: "a" @fragment; }`
		m := model.Build([]byte(text))
		Expect(m.Definitions()).To(Equal([]model.Definition{{Name: "A", Kind: model.KindFragment, Span: span(text, "A", 0)}}))
		Expect(m.Aliases()).To(BeEmpty())
	})

	It("ends a scanner rule at the closing brace", func() {
		text := `@scanner { A: "a" } @parser { s: A; }`
		m := model.Build([]byte(text))
		Expect(definitionNames(m.Definitions())).To(Equal([]string{"A", "s"}))
		Expect(m.References()).To(Equal([]model.Reference{{Name: "A", Span: span(text, "A", 1)}}))
	})

	It("ends a section at the next section keyword", func() {
		text := `@scanner { A: "a"; @parser { s: A; }`
		m := model.Build([]byte(text))
		Expect(m.Definitions()).To(Equal([]model.Definition{
			{Name: "A", Kind: model.KindTerminal, Span: span(text, "A", 0)},
			{Name: "s", Kind: model.KindNonterminal, Span: span(text, "s", 0)},
		}))
		Expect(m.References()).To(Equal([]model.Reference{{Name: "A", Span: span(text, "A", 1)}}))
	})

	It("ends a precedence line at the closing brace", func() {
		text := `@scanner { A: "a"; } @parser { @precedence { @left: A } s: A; }`
		m := model.Build([]byte(text))
		Expect(definitionNames(m.Definitions())).To(Equal([]string{"A", "s"}))
		Expect(referenceNames(m.References())).To(Equal([]string{"A", "A"}))
	})

	It("keeps the references of a production with a syntax error", func() {
		text := `@scanner { A: "a"; B: "b"; } @parser { s: A @ B; }`
		Expect(referenceNames(model.Build([]byte(text)).References())).To(Equal([]string{"A", "B"}))
	})

	It("keeps the production being typed at the end of the text", func() {
		text := `@scanner { A: "a"; } @parser { s: A`
		m := model.Build([]byte(text))
		Expect(definitionNames(m.Definitions())).To(Equal([]string{"A", "s"}))
		Expect(m.References()).To(Equal([]model.Reference{{Name: "A", Span: span(text, "A", 1)}}))
	})

	It("keeps the next rule after an unterminated string, which takes in the semicolon", func() {
		text := "@scanner {\n    A: \"a;\n    B: \"b\";\n}"
		m := model.Build([]byte(text))
		Expect(definitionNames(m.Definitions())).To(Equal([]string{"A", "B"}))
		Expect(m.Aliases()).To(Equal([]model.Alias{{Name: `"b"`, Terminal: "B", Span: span(text, `"b"`, 0)}}))
	})

	It("does not take the name of @name, @error or @empty as symbols", func() {
		text := `@parser { s: @error @name(n) | @empty @name(m); }`
		m := model.Build([]byte(text))
		Expect(definitionNames(m.Definitions())).To(Equal([]string{"s"}))
		Expect(m.References()).To(BeEmpty())
	})

	It("skips comments", func() {
		text := "@scanner { /* B: \"b\"; */ A: \"a\"; // C: \"c\";\n } @parser { s: /* B */ A; }"
		m := model.Build([]byte(text))
		Expect(definitionNames(m.Definitions())).To(Equal([]string{"A", "s"}))
		Expect(referenceNames(m.References())).To(Equal([]string{"A"}))
	})
})
