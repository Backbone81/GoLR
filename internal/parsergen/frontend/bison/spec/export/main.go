// This application exports the go specification as JSON and YAML. This makes the grammar available for other
// programming languages, and it makes it obvious in case the grammar changes by accident.
package main

import (
	"golr/examples/bison/spec"
	golangparsergen "golr/internal/parsergen/backend/golang"
	yamlparsergen "golr/internal/parsergen/backend/yaml"
	"golr/internal/parsergen/core/ielr1"
	frontend2 "golr/internal/parsergen/frontend"
	golangscannergen "golr/internal/scannergen/backend/golang"
	yamlscannergen "golr/pkg/scannergen/backend/yaml"
	"golr/pkg/scannergen/core/subset"
)

func main() {
	rules := spec.GetScannerRules()
	dfa := subset.RulesToDFA(rules)
	if err := yamlscannergen.DFAToFile("internal/parsergen/frontend/bison/parser/scanner.yaml", dfa); err != nil {
		panic(err)
	}
	if err := golangscannergen.DFAToFile("internal/parsergen/frontend/bison/parser/scanner.go", dfa, golangscannergen.Config{PackageName: "parser"}); err != nil {
		panic(err)
	}

	parser, err := ielr1.GrammarToParser(frontend2.Grammar{})
	if err != nil {
		panic(err)
	}
	if err := yamlparsergen.ParserToFile("internal/parsergen/frontend/bison/parser/parser.yaml", parser); err != nil {
		panic(err)
	}
	if err := golangparsergen.ParserToFile("internal/parsergen/frontend/bison/parser/parser.go", parser, golangparsergen.Config{PackageName: "parser"}); err != nil {
		panic(err)
	}
}
