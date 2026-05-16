// This application exports the GNU Bison specification as Go code. This makes the grammar available for the frontend
// to actually process grammar files.
package main

import (
	golangparsergen "github.com/backbone81/golr/internal/parsergen/backend/golang"
	"github.com/backbone81/golr/internal/parsergen/core/ielr1"
	"github.com/backbone81/golr/internal/parsergen/frontend/bison"
	"github.com/backbone81/golr/internal/parsergen/frontend/bison/spec"
	golangscannergen "github.com/backbone81/golr/internal/scannergen/backend/golang"
	"github.com/backbone81/golr/pkg/scannergen/core/subset"
)

func main() {
	rules := spec.GetScannerRules()
	dfa := subset.RulesToDFA(rules)
	if err := golangscannergen.DFAToFile(
		"internal/parsergen/frontend/bison/parser/scanner.go",
		dfa,
		golangscannergen.Config{
			PackageName: "parser",
		},
	); err != nil {
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
	if err := golangparsergen.ParserToFile(
		"internal/parsergen/frontend/bison/parser/parser.go",
		parser,
		golangparsergen.Config{
			PackageName: "parser",
		},
	); err != nil {
		panic(err)
	}
}
