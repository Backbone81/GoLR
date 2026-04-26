// This application exports the go specification as JSON and YAML. This makes the grammar available for other
// programming languages, and it makes it obvious in case the grammar changes by accident.
package main

import (
	"golr/examples/golang/spec"
	"golr/pkg/scannergen/frontend/json"
	"golr/pkg/scannergen/frontend/yaml"
)

func main() {
	rules := spec.GetScannerRules()
	if err := json.RulesToFile("examples/golang/spec/scanner.json", rules); err != nil {
		panic(err)
	}
	if err := yaml.RulesToFile("examples/golang/spec/scanner.yaml", rules); err != nil {
		panic(err)
	}
}
