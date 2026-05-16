package yaml

import (
	intyaml "github.com/backbone81/golr/internal/parsergen/backend/yaml"
)

// FromParser writes the parser as YAML document to the given writer. Returns an error if the YAML document can not be
// encoded successfully.
var FromParser = intyaml.FromParser

// ParserToFile writes the parser as YAML document to the given file path. Returns an error if the file can not be
// written or the YAML document can not be encoded successfully.
var ParserToFile = intyaml.ParserToFile

// ParserToString returns the parser as YAML document. Returns an error if the YAML document can not be encoded
// successfully.
var ParserToString = intyaml.ParserToString
