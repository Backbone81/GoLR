package model_test

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/onsi/gomega/format"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/model"
	parsergenfrontend "github.com/backbone81/golr/internal/parsergen/frontend"
	golrfrontend "github.com/backbone81/golr/internal/parsergen/frontend/golr"
	"github.com/backbone81/golr/internal/utils"
)

// grammar is a valid file with every kind of definition, alias and reference.
var grammar = utils.HereDoc(`
	@scanner {
		DIGIT:  /[0-9]/ @fragment;
		WS:     /[ ]+/ @skip;
		PLUS:   "+";
		MINUS:  "-";
		NUMBER: /{DIGIT}+/;
	}
	
	@parser {
		@start: expr;
		@precedence {
			@left:       PLUS "-";
			@precedence: UMINUS;
		}
		expr
			: expr "+" expr @name(add)
			| expr MINUS expr
			| "-" expr @precedence(UMINUS) @name(neg)
			| NUMBER
			| @error
			;
		empty: @empty;
	}
`)

// scannerSection is a scanner section in front of the inputs of the parser section.
const scannerSection = `@scanner { A: "a"; } `

func TestSuite(t *testing.T) {
	format.MaxLength = 0
	RegisterFailHandler(Fail)
	RunSpecs(t, "Model Suite")
}

// span returns the span of the nth occurrence of the word in the text, counted from zero. A word which starts or ends
// with an identifier character matches only where it is not part of a longer identifier.
func span(text string, word string, nth int) model.Span {
	for offset := 0; ; offset++ {
		idx := strings.Index(text[offset:], word)
		if idx < 0 {
			return model.Span{Offset: -1}
		}
		offset += idx
		if (isIdentifierByte(word, 0) && isIdentifierByte(text, offset-1)) ||
			(isIdentifierByte(word, len(word)-1) && isIdentifierByte(text, offset+len(word))) {
			continue
		}
		if nth == 0 {
			return model.Span{Offset: offset, Length: len(word)}
		}
		nth--
	}
}

// isIdentifierByte reports whether the byte at the index exists and can be part of an identifier.
func isIdentifierByte(text string, idx int) bool {
	if idx < 0 || idx >= len(text) {
		return false
	}
	value := text[idx]
	return value == '_' || ('a' <= value && value <= 'z') || ('A' <= value && value <= 'Z') ||
		('0' <= value && value <= '9')
}

// names returns the names of the definitions.
func definitionNames(definitions []model.Definition) []string {
	result := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		result = append(result, definition.Name)
	}
	return result
}

// referenceNames returns the names of the references.
func referenceNames(references []model.Reference) []string {
	result := make([]string, 0, len(references))
	for _, reference := range references {
		result = append(result, reference.Name)
	}
	return result
}

// caret is the marker of the caret in the inputs of the completion specs.
const caret = "‸"

// withoutCaret returns the text without the caret marker and the offset of the marker.
func withoutCaret(text string) (string, int) {
	offset := strings.Index(text, caret)
	return strings.Replace(text, caret, "", 1), offset
}

// expectModelMatchesFrontend checks the model of a valid text against the grammar the golr frontend reads from it: the
// same terminals in the same order, the same nonterminals and the same aliases, and every reference resolves.
func expectModelMatchesFrontend(text []byte, filePath string) {
	_, grammar, err := golrfrontend.ToGrammar(bytes.NewReader(text), filePath)
	Expect(err).ToNot(HaveOccurred())
	m := model.Build(text)

	Expect(modelNames(m, model.KindTerminal)).To(Equal(grammarTerminals(grammar)), "terminals")
	Expect(modelNames(m, model.KindNonterminal)).To(ConsistOf(grammarNonterminals(grammar)), "nonterminals")
	Expect(modelAliases(m)).To(Equal(grammarAliases(grammar)), "aliases")
	Expect(unresolvedReferences(m)).To(BeEmpty(), "unresolved references")
}

// modelNames returns the names of the definitions of the kind in the order of the text, each name once.
func modelNames(m *model.Model, kind model.Kind) []string {
	var names []string
	for _, definition := range m.Definitions() {
		if definition.Kind == kind && !slices.Contains(names, definition.Name) {
			names = append(names, definition.Name)
		}
	}
	return names
}

// grammarTerminals returns the names of the terminals of the grammar without the error symbol, which the frontend adds
// on its first use.
func grammarTerminals(grammar parsergenfrontend.Grammar) []string {
	var names []string
	for _, terminal := range grammar.Terminals {
		if terminal.Name != parsergenfrontend.SymbolError.Name {
			names = append(names, terminal.Name)
		}
	}
	return names
}

// grammarNonterminals returns the names of the nonterminals of the grammar.
func grammarNonterminals(grammar parsergenfrontend.Grammar) []string {
	names := make([]string, 0, len(grammar.Nonterminals))
	for _, nonterminal := range grammar.Nonterminals {
		names = append(names, nonterminal.Name)
	}
	return names
}

// modelAliases returns the terminal of every alias of the model.
func modelAliases(m *model.Model) map[string]string {
	terminalByAlias := map[string]string{}
	for _, alias := range m.Aliases() {
		terminalByAlias[alias.Name] = alias.Terminal
	}
	return terminalByAlias
}

// grammarAliases returns the terminal of every alias of the grammar.
func grammarAliases(grammar parsergenfrontend.Grammar) map[string]string {
	terminalByAlias := map[string]string{}
	for _, terminal := range grammar.Terminals {
		if terminal.Alias != "" {
			terminalByAlias[terminal.Alias] = terminal.Name
		}
	}
	return terminalByAlias
}

// unresolvedReferences returns the names of the references which resolve to no declaration.
func unresolvedReferences(m *model.Model) []string {
	var names []string
	for _, reference := range m.References() {
		if len(m.DeclarationsNamed(reference.Name)) == 0 {
			names = append(names, reference.Name)
		}
	}
	return names
}
