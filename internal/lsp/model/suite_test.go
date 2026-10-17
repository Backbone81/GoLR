package model_test

import (
	"strings"
	"testing"

	"github.com/onsi/gomega/format"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/model"
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
