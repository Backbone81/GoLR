import { Scanner, Token, TokenSkipper } from "./parser/scanner.js";
import { Parser, ParseSymbol } from "./parser/parser.js";

// evaluate parses the expression and returns the number it stands for.
export function evaluate(expression) {
    // The generated TokenSkipper will skip all whitespaces which the parser is not interested in.
    const scanner = new TokenSkipper(
        // The generated Scanner will convert the input into tokens. The filePath argument is used in error messages.
        // It wants the bytes of the input, not the string, so that offsets are byte offsets.
        new Scanner(new TextEncoder().encode(expression), "expression"),
    );

    // The expression is parsed by giving the generated parser the scanner to pull tokens from. We get the root node of
    // the parse tree and the errors the parse found.
    const { tree, errors } = new Parser().parse(scanner);
    if (errors.length !== 0) {
        throw new Error(errors[0].message);
    }

    // Traversing over the parse tree will calculate the result for us.
    return evaluateNode(tree);
}

// evaluateNode recursively evaluates an expression node from the parse tree.
// The number of children encodes which grammar production was matched:
//   - 1 child:  INTEGER literal
//   - 2 children: unary minus ("-" expression)
//   - 3 children: binary operation (expression OP expression) or grouping ("(" expression ")")
function evaluateNode(node) {
    // Each node has a symbol (the grammar symbol it represents), a lexeme (the raw bytes from the input, set for
    // terminal nodes), and children (sub-nodes).
    switch (node.children.length) {
        case 1:
            // expression: INTEGER
            return Number(new TextDecoder().decode(node.children[0].lexeme));
        case 2:
            // expression: "-" expression
            return -evaluateNode(node.children[1]);
        case 3:
            return evaluateThreeChildren(node);
    }
    throw new Error("unexpected node structure");
}

// evaluateThreeChildren evaluates the two productions with three symbols on the right hand side.
function evaluateThreeChildren(node) {
    // In "(" expression ")", the middle child is the nonterminal expression node.
    // In "expression OP expression", the middle child is a terminal operator token.
    // We use this to distinguish the two cases.
    const operator = ParseSymbol.terminal(node.children[1].symbol);
    if (operator === null) {
        // expression: "(" expression ")"
        return evaluateNode(node.children[1]);
    }

    const leftValue = evaluateNode(node.children[0]);
    const rightValue = evaluateNode(node.children[2]);

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
            // The calculator is integer based, and JavaScript division is not, so the result is truncated towards
            // zero the way it is in the other languages of this example.
            return Math.trunc(leftValue / rightValue);
    }
    throw new Error("unexpected node structure");
}

// We expect this to be used like "calculator '4 + 5 * 3'"
const args = process.argv.slice(2);
if (args.length !== 1) {
    process.stderr.write("usage: calculator <expression>\n");
    process.exit(1);
}

try {
    process.stdout.write(`${evaluate(args[0])}\n`);
} catch (error) {
    process.stderr.write(`${error.message}\n`);
    process.exit(1);
}
