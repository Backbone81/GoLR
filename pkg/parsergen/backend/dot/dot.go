package dot

import intdot "github.com/backbone81/golr/internal/parsergen/backend/dot"

var (
	// FromParser writes the parser as DOT document to the given writer. Returns an error if the DOT document can not be
	// encoded successfully.
	FromParser = intdot.FromParser

	// ParserToFile writes the parser as DOT document to the given file path. Returns an error if the file can not be
	// written or the DOT document can not be encoded successfully.
	ParserToFile = intdot.ParserToFile

	// ParserToString returns the parser as DOT document. Returns an error if the DOT document can not be encoded
	// successfully.
	ParserToString = intdot.ParserToString
)
