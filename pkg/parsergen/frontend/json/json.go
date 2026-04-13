package json

import (
	intjson "golr/internal/parsergen/frontend/json"
)

// ToGrammar reads the context free grammar as JSON document from the given reader. Returns an error if the JSON
// document can not be decoded successfully.
var ToGrammar = intjson.ToGrammar

// GrammarFromFile reads the context free grammar as JSON document from the given file path. Returns an error if the
// file can not be read or the JSON document can not be decoded successfully.
var GrammarFromFile = intjson.GrammarFromFile

// GrammarFromString reads the context free grammar as JSON document from the given string. Returns an error if the
// JSON document can not be decoded successfully.
var GrammarFromString = intjson.GrammarFromString
