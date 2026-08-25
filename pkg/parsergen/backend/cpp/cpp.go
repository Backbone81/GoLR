package cpp

import intcpp "github.com/backbone81/golr/internal/parsergen/backend/cpp"

// DefaultNamespace is the C++ namespace the generated parser is declared in when the caller names none.
const DefaultNamespace = intcpp.DefaultNamespace

// DefaultScannerInclude is the header the generated parser includes the token type from when the caller names none.
const DefaultScannerInclude = intcpp.DefaultScannerInclude

type (
	Config = intcpp.Config
)

var (
	// FromParser writes the parser as C++ source code to the given writer. Returns an error if the C++ source code
	// can not be encoded successfully.
	FromParser = intcpp.FromParser

	// ParserToFile writes the parser as C++ source code to the given file path. Returns an error if the file can not
	// be written or the C++ source code can not be encoded successfully.
	ParserToFile = intcpp.ParserToFile

	// ParserToString returns the parser as C++ source code. Returns an error if the C++ source code can not be
	// encoded successfully.
	ParserToString = intcpp.ParserToString
)
