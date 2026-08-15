package typescript

import inttypescript "github.com/backbone81/golr/internal/parsergen/backend/typescript"

// DefaultScannerModule is the module specifier the generated parser imports the token constants from when the caller
// names none.
const DefaultScannerModule = inttypescript.DefaultScannerModule

type (
	Config = inttypescript.Config
)

var (
	// FromParser writes the parser as TypeScript source code to the given writer. Returns an error if the TypeScript
	// source code can not be encoded successfully.
	FromParser = inttypescript.FromParser

	// ParserToFile writes the parser as TypeScript source code to the given file path. Returns an error if the file
	// can not be written or the TypeScript source code can not be encoded successfully.
	ParserToFile = inttypescript.ParserToFile

	// ParserToString returns the parser as TypeScript source code. Returns an error if the TypeScript source code can
	// not be encoded successfully.
	ParserToString = inttypescript.ParserToString
)
