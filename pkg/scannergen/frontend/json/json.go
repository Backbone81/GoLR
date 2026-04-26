package json

import intjson "golr/internal/scannergen/frontend/json"

var (
	// ToRules reads the scanner rules as JSON document from the given reader. Returns an error if the JSON
	// document can not be decoded successfully.
	ToRules = intjson.ToRules

	// FromRules writes the scanner rules as JSON document to the given writer. Returns an error if the JSON
	// document can not be encoded successfully.
	FromRules = intjson.FromRules

	// RulesFromFile reads the scanner rules as JSON document from the given file path. Returns an error if the
	// file can not be read or the JSON document can not be decoded successfully.
	RulesFromFile = intjson.RulesFromFile

	// RulesToFile writes the scanner rules as JSON document to the given file path. Returns an error if the file
	// can not be written or the JSON document can not be encoded successfully.
	RulesToFile = intjson.RulesToFile

	// RulesFromString reads the scanner rules as JSON document from the given string. Returns an error if the
	// JSON document can not be decoded successfully.
	RulesFromString = intjson.RulesFromString

	// RulesToString returns the scanner rules as JSON document. Returns an error if the JSON document can not be
	// encoded successfully.
	RulesToString = intjson.RulesToString
)
