package c

import "github.com/backbone81/golr/internal/parsergen/backend/c"

type (
	Config = c.Config
)

const (
	// DefaultPrefix is the prefix the generated names carry when the caller names none.
	DefaultPrefix = c.DefaultPrefix

	// DefaultScannerInclude is the header the generated parser includes the token type from when the caller names
	// none.
	DefaultScannerInclude = c.DefaultScannerInclude
)

var (
	// FromParser writes the parser as C source code to the given writer. Returns an error if the C source code can
	// not be encoded successfully.
	FromParser = c.FromParser

	// ParserToFile writes the parser as C source code to the given file path. Returns an error if the file can not
	// be written or the C source code can not be encoded successfully.
	ParserToFile = c.ParserToFile

	// ParserToString returns the parser as C source code. Returns an error if the C source code can not be encoded
	// successfully.
	ParserToString = c.ParserToString
)
