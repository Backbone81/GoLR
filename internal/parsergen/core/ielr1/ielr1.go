package ielr1

import (
	"context"
	"golr/internal/parsergen/backend"
	"golr/internal/parsergen/frontend"
	"runtime/trace"
)

// GrammarToParser calculates a parser from the context free grammar.
func GrammarToParser(augmentedGrammar frontend.Grammar) backend.Parser {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Cores: IELR1: GrammarToParser").End()

	builder := NewIELR1(augmentedGrammar)
	return builder.BuildParser()
}

type IELR1 struct {
}

func NewIELR1(augmentedGrammar frontend.Grammar) *IELR1 {
	return &IELR1{}
}

func (i *IELR1) BuildParser() backend.Parser {
	return backend.Parser{}
}
