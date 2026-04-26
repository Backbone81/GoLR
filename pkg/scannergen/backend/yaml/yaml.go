package yaml

import intyaml "golr/internal/scannergen/backend/yaml"

var (
	// ToDFA reads the deterministic finite automaton as YAML document from the given reader. Returns an error if the YAML
	// document can not be decoded successfully.
	ToDFA = intyaml.ToDFA

	// FromDFA writes the deterministic finite automaton as YAML document to the given writer. Returns an error if the YAML
	// document can not be encoded successfully.
	FromDFA = intyaml.FromDFA

	// DFAFromFile reads the deterministic finite automaton as YAML document from the given file path. Returns an error if the
	// file can not be read or the YAML document can not be decoded successfully.
	DFAFromFile = intyaml.DFAFromFile

	// DFAToFile writes the deterministic finite automaton as YAML document to the given file path. Returns an error if the file
	// can not be written or the YAML document can not be encoded successfully.
	DFAToFile = intyaml.DFAToFile

	// DFAFromString reads the deterministic finite automaton as YAML document from the given string. Returns an error if the
	// YAML document can not be decoded successfully.
	DFAFromString = intyaml.DFAFromString

	// DFAToString returns the deterministic finite automaton as YAML document. Returns an error if the YAML document can not be
	// encoded successfully.
	DFAToString = intyaml.DFAToString
)
