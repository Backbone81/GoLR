package dot

import "github.com/backbone81/golr/internal/scannergen/backend/dot"

var (
	// FromDFA writes the DFA as DOT document to the given writer. Returns an error if the DOT document can not be
	// encoded successfully.
	FromDFA = dot.FromDFA

	// DFAToFile writes the DFA as DOT document to the given file path. Returns an error if the file can not be
	// written or the DOT source code can not be encoded successfully.
	DFAToFile = dot.DFAToFile

	// DFAToString returns the parser as DOT document. Returns an error if the DOT document can not be encoded
	// successfully.
	DFAToString = dot.DFAToString
)
