package javascript

import "github.com/backbone81/golr/internal/scannergen/backend/javascript"

var (
	// FromDFA writes the DFA as JavaScript source code to the given writer. Returns an error if the JavaScript source
	// code can not be encoded successfully.
	FromDFA = javascript.FromDFA

	// DFAToFile writes the DFA as JavaScript source code to the given file path. Returns an error if the file can not
	// be written or the JavaScript source code can not be encoded successfully.
	DFAToFile = javascript.DFAToFile
)
