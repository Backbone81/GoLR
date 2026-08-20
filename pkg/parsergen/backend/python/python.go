package python

import intpython "github.com/backbone81/golr/internal/parsergen/backend/python"

// DefaultScannerModule is the module the generated parser imports the token constants from when the caller names none.
const DefaultScannerModule = intpython.DefaultScannerModule

type (
	Config = intpython.Config
)

var (
	// FromParser writes the parser as Python source code to the given writer. Returns an error if the Python source
	// code can not be encoded successfully.
	FromParser = intpython.FromParser

	// ParserToFile writes the parser as Python source code to the given file path. Returns an error if the file can
	// not be written or the Python source code can not be encoded successfully.
	ParserToFile = intpython.ParserToFile

	// ParserToString returns the parser as Python source code. Returns an error if the Python source code can not be
	// encoded successfully.
	ParserToString = intpython.ParserToString
)
