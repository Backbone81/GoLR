package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/backbone81/golr/examples/calculator/parser"
	"github.com/backbone81/golr/pkg/runtime"
)

func main() {
	// We expect this binary to be used like "calculator '4 + 5 * 3'"
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: calculator <expression>")
		os.Exit(1)
	}

	result, err := Evaluate(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(result)
}

func Evaluate(expression string) (int, error) {
	// The UTF8 rune reader is responsible for decoding UTF8 encoded runes when they are encoded in more than one byte.
	runeReader := runtime.NewUTF8RuneReader([]byte(expression))

	// The generated TokenSkipper will skip all whitespaces which the parser is not interested in.
	scanner := &parser.TokenSkipper{
		// The generated Scanner will convert the input into tokens. The filePath argument is used in error messages.
		Scanner: parser.NewScanner(runeReader, "expression"),
	}

	// The expression is parsed by giving the generated parser the scanner to pull tokens from. We get the root node
	// of the abstract syntax tree as a result.
	rootNode, err := parser.NewParser().Parse(scanner)
	if err != nil {
		return 0, err
	}

	// Traversing over the abstract syntax tree will calculate the result for us.
	result, err := evaluateNode(rootNode)
	if err != nil {
		return 0, err
	}
	return result, nil
}

// evaluateNode recursively evaluates an expression node from the abstract syntax tree.
// The number of children encodes which grammar production was matched:
//   - 1 child:  INTEGER literal
//   - 2 children: unary minus ("-" expression)
//   - 3 children: binary operation (expression OP expression) or grouping ("(" expression ")")
func evaluateNode(node *parser.Node) (int, error) {
	// Each Node has a Symbol (the grammar symbol it represents), a Lexeme (the raw
	// bytes from input, set for terminal nodes), and Children (sub-nodes).
	switch len(node.Children) {
	case 1:
		// expression: INTEGER
		return strconv.Atoi(string(node.Children[0].Lexeme))
	case 2:
		// expression: "-" expression
		value, err := evaluateNode(node.Children[1])
		if err != nil {
			return 0, err
		}
		return -value, nil
	case 3:
		// In "(" expression ")", the middle child is the nonterminal expression node.
		// In "expression OP expression", the middle child is a terminal operator token.
		// We use this to distinguish the two cases.
		token, isTerminal := node.Children[1].Symbol.Terminal()
		if !isTerminal {
			// expression: "(" expression ")"
			return evaluateNode(node.Children[1])
		}

		leftValue, err := evaluateNode(node.Children[0])
		if err != nil {
			return 0, err
		}

		rightValue, err := evaluateNode(node.Children[2])
		if err != nil {
			return 0, err
		}

		//nolint:exhaustive // No need to be exhaustive here.
		switch token {
		case parser.TokenPlus:
			// expression: expression "+" expression
			return leftValue + rightValue, nil
		case parser.TokenMinus:
			// expression: expression "-" expression
			return leftValue - rightValue, nil
		case parser.TokenMultiply:
			// expression: expression "*" expression
			return leftValue * rightValue, nil
		case parser.TokenDivide:
			// expression: expression "/" expression
			if rightValue == 0 {
				return 0, errors.New("division by zero")
			}
			return leftValue / rightValue, nil
		}
	}
	return 0, errors.New("unexpected node structure")
}
