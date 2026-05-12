// This application exports the go specification as JSON and YAML. This makes the grammar available for other
// programming languages, and it makes it obvious in case the grammar changes by accident.
package main

import (
	golangparsergen "golr/internal/parsergen/backend/golang"
	"golr/internal/parsergen/core/ielr1"
	"golr/internal/parsergen/frontend/bison"
	"golr/internal/parsergen/frontend/bison/spec"
	golangscannergen "golr/internal/scannergen/backend/golang"
	"golr/pkg/scannergen/core/subset"
)

func main() {
	rules := spec.GetScannerRules()
	dfa := subset.RulesToDFA(rules)
	if err := golangscannergen.DFAToFile("internal/parsergen/frontend/bison/parser/scanner.go", dfa, golangscannergen.Config{PackageName: "parser"}); err != nil {
		panic(err)
	}

	grammar, err := bison.GrammarFromFile("internal/parsergen/frontend/bison/spec/bison-3.8.2.y")
	if err != nil {
		panic(err)
	}

	parser, err := ielr1.GrammarToParser(grammar)
	if err != nil {
		panic(err)
	}
	if err := golangparsergen.ParserToFile("internal/parsergen/frontend/bison/parser/parser.go", parser, golangparsergen.Config{PackageName: "parser"}); err != nil {
		panic(err)
	}
}
