package json

import intjson "golr/internal/scannergen/backend/json"

var (
	// ToDFA reads the deterministic finite automaton as JSON document from the given reader. Returns an error if the
	// JSON document can not be decoded successfully.
	ToDFA = intjson.ToDFA

	// FromDFA writes the deterministic finite automaton as JSON document to the given writer. Returns an error if the
	// JSON document can not be encoded successfully.
	FromDFA = intjson.FromDFA

	// DFAFromFile reads the deterministic finite automaton as JSON document from the given file path. Returns an error
	// if the file can not be read or the JSON document can not be decoded successfully.
	DFAFromFile = intjson.DFAFromFile

	// DFAToFile writes the deterministic finite automaton as JSON document to the given file path. Returns an error if
	// the file can not be written or the JSON document can not be encoded successfully.
	DFAToFile = intjson.DFAToFile

	// DFAFromString reads the deterministic finite automaton as JSON document from the given string. Returns an error
	// if the JSON document can not be decoded successfully.
	DFAFromString = intjson.DFAFromString

	// DFAToString returns the deterministic finite automaton as JSON document. Returns an error if the JSON document
	// can not be encoded successfully.
	DFAToString = intjson.DFAToString
)
