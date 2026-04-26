// This application exports the go specification as JSON and YAML. This makes the grammar available for other
// programming languages, and it makes it obvious in case the grammar changes by accident.
package main

import (
	"golr/examples/golang/spec"
	"golr/pkg/scannergen/backend/json"
	"golr/pkg/scannergen/backend/yaml"
	"golr/pkg/scannergen/core/subset"
)

func main() {
	rules := spec.GetScannerRules()
	dfa := subset.RulesToDFA(rules)
	if err := json.DFAToFile("examples/golang/parser/scanner.json", dfa); err != nil {
		panic(err)
	}
	if err := yaml.DFAToFile("examples/golang/parser/scanner.yaml", dfa); err != nil {
		panic(err)
	}
}
