// This application exports the GoLR specification as Go code. This makes the grammar available for the frontend
// to actually process grammar files.
package main

import (
	"github.com/backbone81/golr/internal/parsergen/frontend/golr/spec"
	golangscannergen "github.com/backbone81/golr/internal/scannergen/backend/golang"
	"github.com/backbone81/golr/pkg/scannergen/core/subset"
)

func main() {
	rules := spec.GetScannerRules()
	dfa := subset.RulesToDFA(rules)
	if err := golangscannergen.DFAToFile(
		"internal/parsergen/frontend/golr/parser/scanner.go",
		dfa,
		golangscannergen.Config{
			PackageName: "parser",
		},
	); err != nil {
		panic(err)
	}
}
