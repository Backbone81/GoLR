package main

import (
	"errors"
	"fmt"
	"golr/examples/bison/parser"
	"golr/pkg/runtime"
	"io"
	"os"
)

func main() {
	filePath := "examples/bison/spec/bison-3.8.2.y"
	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	runeReader := runtime.NewUTF8RuneReader(data)
	scanner := parser.WhitespaceSkipper{
		Scanner: parser.NewScanner(runeReader, filePath),
	}

	var tokenCounter int
	for scanner.Next() {
		tokenCounter++
		fmt.Printf("%d:%d:%s %q\n", scanner.Line(), scanner.Column(), scanner.Token(), scanner.Lexeme())
	}
	if scanner.Err() != nil && !errors.Is(scanner.Err(), io.EOF) {
		panic(fmt.Sprintf("%s:%d:%d: %s\n", filePath, scanner.Line(), scanner.Column(), scanner.Err()))
	}

	fmt.Println()
	fmt.Printf("%d tokens\n", tokenCounter)
}
