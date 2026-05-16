package yaml

import (
	intyaml "github.com/backbone81/golr/internal/parsergen/frontend/yaml"
)

// ToGrammar reads the context free grammar as YAML document from the given reader. Returns an error if the YAML
// document can not be decoded successfully.
var ToGrammar = intyaml.ToGrammar

// GrammarFromFile reads the context free grammar as YAML document from the given file path. Returns an error if the
// file can not be read or the YAML document can not be decoded successfully.
var GrammarFromFile = intyaml.GrammarFromFile

// GrammarFromString reads the context free grammar as YAML document from the given string. Returns an error if the
// YAML document can not be decoded successfully.
var GrammarFromString = intyaml.GrammarFromString
