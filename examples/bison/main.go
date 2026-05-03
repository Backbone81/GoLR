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

	printTokens(filePath, data)
	printAbstractSyntaxTree(filePath, data)
}

func printTokens(filePath string, data []byte) {
	runeReader := runtime.NewUTF8RuneReader(data)
	scanner := parser.TokenTransformer{
		Scanner: &parser.WhitespaceSkipper{
			Scanner: parser.NewScanner(runeReader, filePath),
		},
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

func printAbstractSyntaxTree(filePath string, data []byte) {
	runeReader := runtime.NewUTF8RuneReader(data)
	scanner := parser.TokenTransformer{
		Scanner: &parser.WhitespaceSkipper{
			Scanner: parser.NewScanner(runeReader, filePath),
		},
	}

	parser := parser.NewParser()
	rootNode, err := parser.Parse(&scanner)
	if err != nil {
		panic(err)
	}
	nodeCount := printTree(rootNode, "", true)

	fmt.Println()
	fmt.Printf("%d nodes\n", nodeCount)
}

func printTree(node *parser.Node, prefix string, isLast bool) int {
	connector := "├─ "
	childPrefix := prefix + "│  "
	if isLast {
		connector = "└─ "
		childPrefix = prefix + "   "
	}
	fmt.Printf("%s%s%s %q\n", prefix, connector, node.Symbol, node.Lexeme)

	var nodeCounter int
	for i, child := range node.Children {
		nodeCounter++
		nodeCounter += printTree(child, childPrefix, i == len(node.Children)-1)
	}
	return nodeCounter
}
