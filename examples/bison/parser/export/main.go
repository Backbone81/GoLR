// This application exports the go specification as JSON and YAML. This makes the grammar available for other
// programming languages, and it makes it obvious in case the grammar changes by accident.
package main

import (
	"golr/examples/bison/spec"
	"golr/internal/scannergen/backend/golang"
	"golr/pkg/scannergen/backend/json"
	"golr/pkg/scannergen/backend/yaml"
	"golr/pkg/scannergen/core/subset"
)

func main() {
	rules := spec.GetScannerRules()
	dfa := subset.RulesToDFA(rules)
	if err := json.DFAToFile("examples/bison/parser/scanner.json", dfa); err != nil {
		panic(err)
	}
	if err := yaml.DFAToFile("examples/bison/parser/scanner.yaml", dfa); err != nil {
		panic(err)
	}
	if err := golang.DFAToFile("examples/bison/parser/scanner.go", dfa, golang.Config{PackageName: "parser"}); err != nil {
		panic(err)
	}
}
