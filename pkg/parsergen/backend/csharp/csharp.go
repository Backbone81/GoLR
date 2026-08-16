package csharp

import intcsharp "github.com/backbone81/golr/internal/parsergen/backend/csharp"

type (
	Config = intcsharp.Config
)

var (
	// FromParser writes the parser as C# source code to the given writer. Returns an error if the C# source code can
	// not be encoded successfully.
	FromParser = intcsharp.FromParser

	// ParserToFile writes the parser as C# source code to the given file path. Returns an error if the file can not
	// be written or the C# source code can not be encoded successfully.
	ParserToFile = intcsharp.ParserToFile

	// ParserToString returns the parser as C# source code. Returns an error if the C# source code can not be encoded
	// successfully.
	ParserToString = intcsharp.ParserToString
)
