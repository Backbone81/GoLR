package java

import "github.com/backbone81/golr/internal/scannergen/backend/java"

type (
	Config = java.Config
)

var (
	// FromDFA writes the DFA as Java source code to the given writer. Returns an error if the Java source code can not
	// be encoded successfully.
	FromDFA = java.FromDFA

	// DFAToFile writes the DFA as Java source code to the given file path. Returns an error if the file can not be
	// written or the Java source code can not be encoded successfully.
	DFAToFile = java.DFAToFile
)
