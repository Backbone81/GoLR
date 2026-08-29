package kotlin

import intkotlin "github.com/backbone81/golr/internal/parsergen/backend/kotlin"

// DefaultPackageName is the Kotlin package the generated parser is declared in when the caller names none.
const DefaultPackageName = intkotlin.DefaultPackageName

type (
	Config = intkotlin.Config
)

var (
	// FromParser writes the parser as Kotlin source code to the given writer. Returns an error if the Kotlin source
	// code can not be encoded successfully.
	FromParser = intkotlin.FromParser

	// ParserToFile writes the parser as Kotlin source code to the given file path. Returns an error if the file can
	// not be written or the Kotlin source code can not be encoded successfully.
	ParserToFile = intkotlin.ParserToFile

	// ParserToString returns the parser as Kotlin source code. Returns an error if the Kotlin source code can not be
	// encoded successfully.
	ParserToString = intkotlin.ParserToString
)
