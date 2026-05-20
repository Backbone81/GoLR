package golr

import (
	intgolr "github.com/backbone81/golr/internal/parsergen/frontend/golr"
)

// ToGrammar reads the context free grammar as GoLR grammar document from the given reader. Returns an error if the
// grammar document can not be parsed successfully.
var ToGrammar = intgolr.ToGrammar

// GrammarFromFile reads the context free grammar as GNU Bison grammar document from the given file path. Returns an
// error if the file can not be read or the grammar document can not be parsed successfully.
var GrammarFromFile = intgolr.GrammarFromFile

// GrammarFromString reads the context free grammar as GNU Bison grammar document from the given string. Returns an
// error if the grammar document can not be parsed successfully.
var GrammarFromString = intgolr.GrammarFromString
