package java

import intjava "github.com/backbone81/golr/internal/parsergen/backend/java"

type (
	Config = intjava.Config
)

var (
	// FromParser writes the parser as Java source code to the given writer. Returns an error if the Java source code
	// can not be encoded successfully.
	FromParser = intjava.FromParser

	// ParserToFile writes the parser as Java source code to the given file path. Returns an error if the file can not
	// be written or the Java source code can not be encoded successfully.
	ParserToFile = intjava.ParserToFile

	// ParserToString returns the parser as Java source code. Returns an error if the Java source code can not be
	// encoded successfully.
	ParserToString = intjava.ParserToString
)
