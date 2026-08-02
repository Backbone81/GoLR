package golangtable

import "github.com/backbone81/golr/internal/scannergen/backend/golangtable"

type (
	Config = golangtable.Config
)

var (
	// FromDFA writes the DFA as Go source code to the given writer. Returns an error if the Go source code can not be
	// encoded successfully.
	FromDFA = golangtable.FromDFA

	// DFAToFile writes the DFA as Go source code to the given file path. Returns an error if the file can not be
	// written or the Go source code can not be encoded successfully.
	DFAToFile = golangtable.DFAToFile
)
