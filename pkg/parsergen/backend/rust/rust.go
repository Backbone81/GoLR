package rust

import intrust "github.com/backbone81/golr/internal/parsergen/backend/rust"

// DefaultScannerModule is the module path the generated parser takes the token type from when the caller names none.
const DefaultScannerModule = intrust.DefaultScannerModule

type (
	Config = intrust.Config
)

var (
	// FromParser writes the parser as Rust source code to the given writer. Returns an error if the Rust source code
	// can not be encoded successfully.
	FromParser = intrust.FromParser

	// ParserToFile writes the parser as Rust source code to the given file path. Returns an error if the file can not
	// be written or the Rust source code can not be encoded successfully.
	ParserToFile = intrust.ParserToFile

	// ParserToString returns the parser as Rust source code. Returns an error if the Rust source code can not be
	// encoded successfully.
	ParserToString = intrust.ParserToString
)
