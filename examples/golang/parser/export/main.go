// This application exports the go specification as JSON and YAML. This makes the grammar available for other
// programming languages, and it makes it obvious in case the grammar changes by accident.
package main

import (
	"github.com/backbone81/golr/examples/golang/spec"
	"github.com/backbone81/golr/pkg/scannergen/backend/golang"
	"github.com/backbone81/golr/pkg/scannergen/backend/json"
	"github.com/backbone81/golr/pkg/scannergen/backend/yaml"
	"github.com/backbone81/golr/pkg/scannergen/core/subset"
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
	if err := golang.DFAToFile(
		"examples/golang/parser/scanner.go",
		dfa,
		golang.Config{PackageName: "parser"},
	); err != nil {
		panic(err)
	}
}
