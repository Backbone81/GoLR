import { Scanner, Token, TokenSkipper } from "./parser/scanner.js";
import { Parser, ParseSymbol } from "./parser/parser.js";
import type { ParseNode } from "./parser/parser.js";

// evaluate parses the expression and returns the number it stands for.
export function evaluate(expression: string): number {
    // The generated TokenSkipper will skip all whitespaces which the parser is not interested in.
    const scanner = new TokenSkipper(
        // The generated Scanner will convert the input into tokens. The filePath argument is used in error messages.
        // It wants the bytes of the input, not the string, so that offsets are byte offsets.
        new Scanner(new TextEncoder().encode(expression), "expression"),
    );

    // The expression is parsed by giving the generated parser the scanner to pull tokens from. We get the root node of
    // the parse tree and the errors the parse found.
    const { tree, errors } = new Parser().parse(scanner);
    const [firstError] = errors;
    if (firstError !== undefined) {
        throw new Error(firstError.message);
    }
    if (tree === null) {
        throw new Error("the expression could not be parsed");
    }

    // Traversing over the parse tree will calculate the result for us.
    return evaluateNode(tree);
}

// childAt returns one child of a node.
function childAt(node: ParseNode, index: number): ParseNode {
    const child = node.children[index];
    if (child === undefined) {
        throw new Error("unexpected node structure");
    }
    return child;
}

// evaluateNode recursively evaluates an expression node from the parse tree.
// The number of children encodes which grammar production was matched:
//   - 1 child:  INTEGER literal
//   - 2 children: unary minus ("-" expression)
//   - 3 children: binary operation (expression OP expression) or grouping ("(" expression ")")
function evaluateNode(node: ParseNode): number {
    // Each node has a symbol (the grammar symbol it represents), a lexeme (the raw bytes from the input, set for
    // terminal nodes), and children (sub-nodes).
    switch (node.children.length) {
        case 1: {
            // expression: INTEGER
            const lexeme = childAt(node, 0).lexeme;
            if (lexeme === null) {
                throw new Error("unexpected node structure");
            }
            return Number(new TextDecoder().decode(lexeme));
        }
        case 2:
            // expression: "-" expression
            return -evaluateNode(childAt(node, 1));
        case 3:
            return evaluateThreeChildren(node);
    }
    throw new Error("unexpected node structure");
}

// evaluateThreeChildren evaluates the two productions with three symbols on the right hand side.
function evaluateThreeChildren(node: ParseNode): number {
    // In "(" expression ")", the middle child is the nonterminal expression node.
    // In "expression OP expression", the middle child is a terminal operator token.
    // We use this to distinguish the two cases.
    const middle = childAt(node, 1);
    const operator = ParseSymbol.terminal(middle.symbol);
    if (operator === null) {
        // expression: "(" expression ")"
        return evaluateNode(middle);
    }

    const leftValue = evaluateNode(childAt(node, 0));
    const rightValue = evaluateNode(childAt(node, 2));

    switch (operator) {
        case Token.TokenPlus:
            // expression: expression "+" expression
            return leftValue + rightValue;
        case Token.TokenMinus:
            // expression: expression "-" expression
            return leftValue - rightValue;
        case Token.TokenMultiply:
            // expression: expression "*" expression
            return leftValue * rightValue;
        case Token.TokenDivide:
            // expression: expression "/" expression
            if (rightValue === 0) {
                throw new Error("division by zero");
            }
            // The calculator is integer based, and TypeScript division is not, so the result is truncated towards
            // zero the way it is in the other languages of this example.
            return Math.trunc(leftValue / rightValue);
    }
    throw new Error("unexpected node structure");
}

// We expect this to be used like "calculator '4 + 5 * 3'"
const args = process.argv.slice(2);
const [expression] = args;
if (args.length !== 1 || expression === undefined) {
    process.stderr.write("usage: calculator <expression>\n");
    process.exit(1);
}

try {
    process.stdout.write(`${evaluate(expression)}\n`);
} catch (error) {
    process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
    process.exit(1);
}
