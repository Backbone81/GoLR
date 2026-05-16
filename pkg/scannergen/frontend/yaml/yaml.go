package yaml

import intyaml "github.com/backbone81/golr/internal/scannergen/frontend/yaml"

var (
	// ToRules reads the scanner rules as YAML document from the given reader. Returns an error if the YAML
	// document can not be decoded successfully.
	ToRules = intyaml.ToRules

	// FromRules writes the scanner rules as YAML document to the given writer. Returns an error if the YAML
	// document can not be encoded successfully.
	FromRules = intyaml.FromRules

	// RulesFromFile reads the scanner rules as YAML document from the given file path. Returns an error if the
	// file can not be read or the YAML document can not be decoded successfully.
	RulesFromFile = intyaml.RulesFromFile

	// RulesToFile writes the scanner rules as YAML document to the given file path. Returns an error if the file
	// can not be written or the YAML document can not be encoded successfully.
	RulesToFile = intyaml.RulesToFile

	// RulesFromString reads the scanner rules as YAML document from the given string. Returns an error if the
	// YAML document can not be decoded successfully.
	RulesFromString = intyaml.RulesFromString

	// RulesToString returns the scanner rules as YAML document. Returns an error if the YAML document can not be
	// encoded successfully.
	RulesToString = intyaml.RulesToString
)
