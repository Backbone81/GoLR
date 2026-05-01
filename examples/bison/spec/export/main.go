// This application exports the go specification as JSON and YAML. This makes the grammar available for other
// programming languages, and it makes it obvious in case the grammar changes by accident.
package main

import (
	"golr/internal/parsergen/backend/golang"
	"golr/internal/parsergen/core/ielr1"
	frontend2 "golr/internal/parsergen/frontend"
)

func main() {
	parser, err := ielr1.GrammarToParser(frontend2.Grammar{})
	if err != nil {
		panic(err)
	}
	if err := golang.ParserToFile("examples/bison/parser/parser.go", parser, golang.Config{PackageName: "parser"}); err != nil {
		panic(err)
	}
}
