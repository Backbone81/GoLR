package bison

import intbison "golr/internal/parsergen/frontend/bison"

var (
	// ToGrammar reads the context free grammar as GNU Bison grammar document from the given reader. Returns an error
	// if the grammar document can not be parsed successfully.
	ToGrammar = intbison.ToGrammar

	// GrammarFromFile reads the context free grammar as GNU Bison grammar document from the given file path. Returns
	// an error if the file can not be read or the grammar document can not be parsed successfully.
	GrammarFromFile = intbison.GrammarFromFile

	// GrammarFromString reads the context free grammar as GNU Bison grammar document from the given string. Returns an
	// error if the grammar document can not be parsed successfully.
	GrammarFromString = intbison.GrammarFromString
)
