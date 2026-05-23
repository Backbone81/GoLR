package golr

import (
	intgolr "github.com/backbone81/golr/internal/parsergen/frontend/golr"
)

// ToGrammar reads the context free grammar as GoLR grammar document from the given reader. Returns an error if the
// grammar document can not be parsed successfully.
var ToGrammar = intgolr.ToGrammar

// FromGrammar writes the context free grammar as GoLR grammar document to the given writer. Returns an error if
// the grammar document can not be encoded successfully.
var FromGrammar = intgolr.FromGrammar

// GrammarFromFile reads the context free grammar as GNU Bison grammar document from the given file path. Returns an
// error if the file can not be read or the grammar document can not be parsed successfully.
var GrammarFromFile = intgolr.GrammarFromFile

// GrammarToFile writes the context free grammar as GoLR grammar document to the given file path. Returns an error
// if the file can not be written or the GoLR document can not be encoded successfully.
var GrammarToFile = intgolr.GrammarToFile

// GrammarFromString reads the context free grammar as GNU Bison grammar document from the given string. Returns an
// error if the grammar document can not be parsed successfully.
var GrammarFromString = intgolr.GrammarFromString

// GrammarToString returns the context free grammar as GoLR grammar document. Returns an error if the GoLR
// document can not be encoded successfully.
var GrammarToString = intgolr.GrammarToString
