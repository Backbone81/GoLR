package golr

import intgolr "github.com/backbone81/golr/internal/scannergen/frontend/golr"

var (
	// ToRules reads the scanner rules as GoLR grammar document from the given reader. Returns an error if the GoLR
	// document can not be decoded successfully.
	ToRules = intgolr.ToRules

	// RulesFromFile reads the scanner rules as GoLR grammar document from the given file path. Returns an error if the
	// file can not be read or the GoLR document can not be decoded successfully.
	RulesFromFile = intgolr.RulesFromFile

	// RulesFromString reads the scanner rules as GoLR grammar document from the given string. Returns an error if the
	// GoLR document can not be decoded successfully.
	RulesFromString = intgolr.RulesFromString
)
