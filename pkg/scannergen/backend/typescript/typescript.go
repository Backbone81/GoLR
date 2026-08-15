package typescript

import "github.com/backbone81/golr/internal/scannergen/backend/typescript"

var (
	// FromDFA writes the DFA as TypeScript source code to the given writer. Returns an error if the TypeScript source
	// code can not be encoded successfully.
	FromDFA = typescript.FromDFA

	// DFAToFile writes the DFA as TypeScript source code to the given file path. Returns an error if the file can not
	// be written or the TypeScript source code can not be encoded successfully.
	DFAToFile = typescript.DFAToFile
)
