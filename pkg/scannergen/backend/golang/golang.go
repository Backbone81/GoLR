package golang

import "golr/internal/scannergen/backend/golang"

type (
	Config = golang.Config
)

var (
	// FromDFA writes the DFA as Go source code to the given writer. Returns an error if the Go source code can not be
	// encoded successfully.
	FromDFA = golang.FromDFA

	// DFAToFile writes the DFA as Go source code to the given file path. Returns an error if the file can not be
	// written or the Go source code can not be encoded successfully.
	DFAToFile = golang.DFAToFile
)
