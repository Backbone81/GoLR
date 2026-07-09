package bison_test

import (
	"testing"

	ielr1bison "github.com/backbone81/golr/internal/parsergen/core/ielr1/bison"
	"github.com/backbone81/golr/internal/parsergen/frontend/bison"
)

type BenchmarkInput struct {
	Title string
	Path  string
}

var BenchmarkInputs = []BenchmarkInput{
	{
		Title: "GNU Bison 3.8.2",
		Path:  "../../../../testdata/bison-3.8.2.y",
	},
	{
		Title: "GCC 2.95.3 C",
		Path:  "../../../../testdata/gcc-2.95.3-c.y",
	},
	{
		Title: "GCC 2.95.3 Objective C",
		Path:  "../../../../testdata/gcc-2.95.3-objc.y",
	},
	{
		Title: "GCC 3.3.6 C++",
		Path:  "../../../../testdata/gcc-3.3.6-cpp.y",
	},
	{
		Title: "GCC 4.2.4 Java",
		Path:  "../../../../testdata/gcc-4.2.4-java.y",
	},
	{
		Title: "Go 1.5.4",
		Path:  "../../../../testdata/go-1.5.4.y",
	},
	{
		Title: "PHP 8.6.7",
		Path:  "../../../../testdata/php-8.6.7.y",
	},
	{
		Title: "PostgreSQL 18.4",
		Path:  "../../../../testdata/postgres-18.4.y",
	},
}

func BenchmarkGrammarToParser(b *testing.B) {
	for _, input := range BenchmarkInputs {
		b.Run(input.Title, func(b *testing.B) {
			grammar, err := bison.GrammarFromFile(input.Path)
			if err != nil {
				b.Fatal(err)
			}

			for b.Loop() {
				if _, err := ielr1bison.GrammarToParser(grammar); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
