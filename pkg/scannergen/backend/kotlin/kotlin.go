package kotlin

import "github.com/backbone81/golr/internal/scannergen/backend/kotlin"

// DefaultPackageName is the Kotlin package the generated scanner is declared in when the caller names none.
const DefaultPackageName = kotlin.DefaultPackageName

type (
	Config = kotlin.Config
)

var (
	// FromDFA writes the DFA as Kotlin source code to the given writer. Returns an error if the Kotlin source code can
	// not be encoded successfully.
	FromDFA = kotlin.FromDFA

	// DFAToFile writes the DFA as Kotlin source code to the given file path. Returns an error if the file can not be
	// written or the Kotlin source code can not be encoded successfully.
	DFAToFile = kotlin.DFAToFile
)
