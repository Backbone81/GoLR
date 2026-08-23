package c

import "github.com/backbone81/golr/internal/scannergen/backend/c"

type (
	Config = c.Config
)

const (
	// DefaultPrefix is the prefix the generated names carry when the caller names none.
	DefaultPrefix = c.DefaultPrefix
)

var (
	// FromDFA writes the DFA as C source code to the given writer. Returns an error if the C source code can not be
	// encoded successfully.
	FromDFA = c.FromDFA

	// DFAToFile writes the DFA as C source code to the given file path. Returns an error if the file can not be
	// written or the C source code can not be encoded successfully.
	DFAToFile = c.DFAToFile
)
