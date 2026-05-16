package json

import (
	intjson "github.com/backbone81/golr/internal/parsergen/backend/json"
)

// FromParser writes the parser as JSON document to the given writer. Returns an error if the JSON document can not be
// encoded successfully.
var FromParser = intjson.FromParser

// ParserToFile writes the parser as JSON document to the given file path. Returns an error if the file can not be
// written or the JSON document can not be encoded successfully.
var ParserToFile = intjson.ParserToFile

// ParserToString returns the parser as JSON document. Returns an error if the JSON document can not be encoded
// successfully.
var ParserToString = intjson.ParserToString
