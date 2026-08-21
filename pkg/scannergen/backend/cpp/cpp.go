package cpp

import "github.com/backbone81/golr/internal/scannergen/backend/cpp"

type (
	Config = cpp.Config
)

var (
	// FromDFA writes the DFA as C++ source code to the given writer. Returns an error if the C++ source code can not
	// be encoded successfully.
	FromDFA = cpp.FromDFA

	// DFAToFile writes the DFA as C++ source code to the given file path. Returns an error if the file can not be
	// written or the C++ source code can not be encoded successfully.
	DFAToFile = cpp.DFAToFile
)
