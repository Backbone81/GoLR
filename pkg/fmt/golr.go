package fmt

import intfmt "github.com/backbone81/golr/internal/fmt"

// GoLR parses the GoLR grammar from the given reader and writes the formatted version to the given writer.
var GoLR = intfmt.GoLR

// GoLRFile parses the GoLR grammar from the given input file path and writes the formatted version to the given output
// file path. Input and output file path can be the same. A temporary file is used to ensure that any parsing errors
// do not lead to an empty input file.
var GoLRFile = intfmt.GoLRFile

// GoLRString parses the GoLR grammar from the given string and returns the formatted version.
var GoLRString = intfmt.GoLRString
