package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"golr/examples/bison/parser"
	"golr/pkg/runtime"
)

func main() {
	if len(os.Args) < 2 {
		panic("provide the file path to a GNU Bison grammar file as parameter")
	}
	filePath := os.Args[1]

	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
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
			Scanner: &parser.ContextScanner{
				Scanner: parser.NewScanner(runeReader, filePath),
			},
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
			Scanner: &parser.ContextScanner{
				Scanner: parser.NewScanner(runeReader, filePath),
			},
		},
	}

	parser := parser.NewParser()
	rootNode, err := parser.Parse(&scanner)
	if err != nil {
		panic(err)
	}
	nodeCount := printTree(rootNode, "", true, 0)

	fmt.Println()
	fmt.Printf("%d nodes\n", nodeCount)
}

func printTree(node *parser.Node, prefix string, isLast bool, depth int) int {
	var connector, childPrefix string
	if depth > 0 {
		if isLast {
			connector = "└─ "
			childPrefix = prefix + "   "
		} else {
			connector = "├─ "
			childPrefix = prefix + "│  "
		}
	}

	if terminal, ok := node.Symbol.Terminal(); ok {
		fmt.Printf("%s%s%s %q\n", prefix, connector, terminal, node.Lexeme)
	} else {
		nonterminal, _ := node.Symbol.Nonterminal()
		fmt.Printf("%s%s%s\n", prefix, connector, nonterminal)
	}

	var nodeCounter int
	for i, child := range node.Children {
		nodeCounter++
		nodeCounter += printTree(child, childPrefix, i == len(node.Children)-1, depth+1)
	}
	return nodeCounter
}
