package javascript

import intjavascript "github.com/backbone81/golr/internal/parsergen/backend/javascript"

// DefaultScannerModule is the module specifier the generated parser imports the token constants from when the caller
// names none.
const DefaultScannerModule = intjavascript.DefaultScannerModule

type (
	Config = intjavascript.Config
)

var (
	// FromParser writes the parser as JavaScript source code to the given writer. Returns an error if the JavaScript
	// source code can not be encoded successfully.
	FromParser = intjavascript.FromParser

	// ParserToFile writes the parser as JavaScript source code to the given file path. Returns an error if the file
	// can not be written or the JavaScript source code can not be encoded successfully.
	ParserToFile = intjavascript.ParserToFile

	// ParserToString returns the parser as JavaScript source code. Returns an error if the JavaScript source code can
	// not be encoded successfully.
	ParserToString = intjavascript.ParserToString
)
