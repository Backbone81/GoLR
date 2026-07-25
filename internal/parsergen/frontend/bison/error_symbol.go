package bison

import (
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// ErrorTokenName is the name GNU Bison knows the error symbol under. GNU Bison predefines the symbol, so a grammar
// neither declares it as a token nor is allowed to declare a token of that name, and its XML report gives the symbol
// back under the same name.
//
// GoLR spells the symbol `$error` instead, with the leading dollar sign which marks a reserved symbol and which no
// terminal a frontend can read is allowed to carry. Every crossing of the boundary to GNU Bison therefore translates
// the name, which is what ToBisonSymbolName and FromBisonSymbolName are for. The boundary is crossed in both
// directions: this frontend reads and writes GNU Bison grammar files, and the Bison-backed parser cores write out the
// grammar and read the resulting tables back from the XML report.
const ErrorTokenName = "error"

// ToBisonSymbolName translates a GoLR symbol name into the name GNU Bison knows it under.
func ToBisonSymbolName(name string) string {
	if name == frontend.SymbolError.Name {
		return ErrorTokenName
	}
	return name
}

// FromBisonSymbolName translates a symbol name GNU Bison reported into the name GoLR knows it under.
func FromBisonSymbolName(name string) string {
	if name == ErrorTokenName {
		return frontend.SymbolError.Name
	}
	return name
}
