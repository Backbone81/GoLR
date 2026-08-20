package python

import "github.com/backbone81/golr/internal/scannergen/backend/python"

var (
	// FromDFA writes the DFA as Python source code to the given writer. Returns an error if the Python source code
	// can not be encoded successfully.
	FromDFA = python.FromDFA

	// DFAToFile writes the DFA as Python source code to the given file path. Returns an error if the file can not be
	// written or the Python source code can not be encoded successfully.
	DFAToFile = python.DFAToFile
)
