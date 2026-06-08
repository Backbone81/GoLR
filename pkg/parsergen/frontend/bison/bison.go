package bison

import intbison "github.com/backbone81/golr/internal/parsergen/frontend/bison"

var (
	// ToGrammar reads the context free grammar as GNU Bison grammar document from the given reader. Returns an error
	// if the grammar document can not be parsed successfully.
	ToGrammar = intbison.ToGrammar

	// FromGrammar writes the context free grammar as GNU Bison grammar document to the given writer. Returns an error
	// if the grammar document can not be encoded successfully.
	FromGrammar = intbison.FromGrammar

	// GrammarFromFile reads the context free grammar as GNU Bison grammar document from the given file path. Returns
	// an error if the file can not be read or the grammar document can not be parsed successfully.
	GrammarFromFile = intbison.GrammarFromFile

	// GrammarToFile writes the context free grammar as GNU Bison grammar document to the given file path. Returns an
	// error if the file can not be written or the GNU Bison document can not be encoded successfully.
	GrammarToFile = intbison.GrammarToFile

	// GrammarFromString reads the context free grammar as GNU Bison grammar document from the given string. Returns an
	// error if the grammar document can not be parsed successfully.
	GrammarFromString = intbison.GrammarFromString

	// GrammarToString returns the context free grammar as GNU Bison grammar document. Returns an error if the GNU Bison
	// document can not be encoded successfully.
	GrammarToString = intbison.GrammarToString
)
