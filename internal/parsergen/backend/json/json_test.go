package json_test

import (
	"io"
	"testing"

	"golr/internal/parsergen/backend/json"
	"golr/internal/parsergen/core/ielr1"
	"golr/internal/parsergen/frontend/bison"
)

func BenchmarkFromParser(b *testing.B) {
	b.Run("GNU Bison 3.8.2", func(b *testing.B) {
		grammar, err := bison.GrammarFromFile("../../../../testdata/bison-3.8.2.y")
		if err != nil {
			b.Fatal(err)
		}

		parser, err := ielr1.GrammarToParser(grammar)
		if err != nil {
			b.Fatal(err)
		}

		for b.Loop() {
			if err := json.FromParser(io.Discard, parser); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("GNU GCC 2.95.3 C", func(b *testing.B) {
		grammar, err := bison.GrammarFromFile("../../../../testdata/gcc-2.95.3-c.y")
		if err != nil {
			b.Fatal(err)
		}

		parser, err := ielr1.GrammarToParser(grammar)
		if err != nil {
			b.Fatal(err)
		}

		for b.Loop() {
			if err := json.FromParser(io.Discard, parser); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("GNU GCC 2.95.3 Objective C", func(b *testing.B) {
		grammar, err := bison.GrammarFromFile("../../../../testdata/gcc-2.95.3-objc.y")
		if err != nil {
			b.Fatal(err)
		}

		parser, err := ielr1.GrammarToParser(grammar)
		if err != nil {
			b.Fatal(err)
		}

		for b.Loop() {
			if err := json.FromParser(io.Discard, parser); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("GNU GCC 3.3.6 C++", func(b *testing.B) {
		grammar, err := bison.GrammarFromFile("../../../../testdata/gcc-3.3.6-cpp.y")
		if err != nil {
			b.Fatal(err)
		}

		parser, err := ielr1.GrammarToParser(grammar)
		if err != nil {
			b.Fatal(err)
		}

		for b.Loop() {
			if err := json.FromParser(io.Discard, parser); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("GNU GCC 4.2.4 Java", func(b *testing.B) {
		grammar, err := bison.GrammarFromFile("../../../../testdata/gcc-4.2.4-java.y")
		if err != nil {
			b.Fatal(err)
		}

		parser, err := ielr1.GrammarToParser(grammar)
		if err != nil {
			b.Fatal(err)
		}

		for b.Loop() {
			if err := json.FromParser(io.Discard, parser); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Go 1.5.4", func(b *testing.B) {
		grammar, err := bison.GrammarFromFile("../../../../testdata/go-1.5.4.y")
		if err != nil {
			b.Fatal(err)
		}

		parser, err := ielr1.GrammarToParser(grammar)
		if err != nil {
			b.Fatal(err)
		}

		for b.Loop() {
			if err := json.FromParser(io.Discard, parser); err != nil {
				b.Fatal(err)
			}
		}
	})
}
