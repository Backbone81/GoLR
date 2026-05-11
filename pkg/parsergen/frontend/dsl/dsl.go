package dsl

import (
	intdsl "golr/internal/parsergen/frontend/dsl"
)

// Grammar describes the context free grammar.
type Grammar = intdsl.Grammar

// NewGrammar creates a new grammar to add terminals, nonterminals and productions to.
var NewGrammar = intdsl.NewGrammar
