// This application exports the GNU Bison specification as JSON, YAML and Go. This makes the grammar available for other
// programming languages, and it makes it obvious in case the grammar changes by accident.
package main

import (
	"github.com/backbone81/golr/examples/bison/spec"
	golangparsergen "github.com/backbone81/golr/pkg/parsergen/backend/golang"
	jsonparsergen "github.com/backbone81/golr/pkg/parsergen/backend/json"
	yamlparsergen "github.com/backbone81/golr/pkg/parsergen/backend/yaml"
	"github.com/backbone81/golr/pkg/parsergen/core/ielr1"
	"github.com/backbone81/golr/pkg/parsergen/frontend/bison"
	golangscannergen "github.com/backbone81/golr/pkg/scannergen/backend/golang"
	jsonscannergen "github.com/backbone81/golr/pkg/scannergen/backend/json"
	yamlscannergen "github.com/backbone81/golr/pkg/scannergen/backend/yaml"
	"github.com/backbone81/golr/pkg/scannergen/core/subset"
)

func main() {
	rules := spec.GetScannerRules()
	dfa := subset.RulesToDFA(rules)
	if err := jsonscannergen.DFAToFile("examples/bison/parser/scanner.json", dfa); err != nil {
		panic(err)
	}
	if err := yamlscannergen.DFAToFile("examples/bison/parser/scanner.yaml", dfa); err != nil {
		panic(err)
	}
	if err := golangscannergen.DFAToFile(
		"examples/bison/parser/scanner.go",
		dfa,
		golangscannergen.Config{
			PackageName: "parser",
		},
	); err != nil {
		panic(err)
	}

	grammar, err := bison.GrammarFromFile("examples/bison/spec/bison-3.8.2.y")
	if err != nil {
		panic(err)
	}

	parser, err := ielr1.GrammarToParser(grammar)
	if err != nil {
		panic(err)
	}
	if err := jsonparsergen.ParserToFile("examples/bison/parser/parser.json", parser); err != nil {
		panic(err)
	}
	if err := yamlparsergen.ParserToFile("examples/bison/parser/parser.yaml", parser); err != nil {
		panic(err)
	}
	if err := golangparsergen.ParserToFile(
		"examples/bison/parser/parser.go",
		parser,
		golangparsergen.Config{
			PackageName: "parser",
		},
	); err != nil {
		panic(err)
	}
}
