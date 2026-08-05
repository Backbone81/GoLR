package golangtable

import intgolangtable "github.com/backbone81/golr/internal/parsergen/backend/golangtable"

type (
	Config = intgolangtable.Config
)

var (
	// FromParser writes the parser as Go source code to the given writer. Returns an error if the Go source code can
	// not be encoded successfully.
	FromParser = intgolangtable.FromParser

	// ParserToFile writes the parser as Go source code to the given file path. Returns an error if the file can not be
	// written or the Go source code can not be encoded successfully.
	ParserToFile = intgolangtable.ParserToFile

	// ParserToString returns the parser as Go source code. Returns an error if the Go source code can not be encoded
	// successfully.
	ParserToString = intgolangtable.ParserToString
)
