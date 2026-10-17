package model_test

import (
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
})
