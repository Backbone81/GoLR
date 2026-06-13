package rust

import "github.com/backbone81/golr/internal/scannergen/backend/rust"

var (
	// FromDFA writes the DFA as Rust source code to the given writer. Returns an error if the Rust source code can not
	// be encoded successfully.
	FromDFA = rust.FromDFA

	// DFAToFile writes the DFA as Rust source code to the given file path. Returns an error if the file can not be
	// written or the Rust source code can not be encoded successfully.
	DFAToFile = rust.DFAToFile
)
