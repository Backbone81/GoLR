import java.nio.charset.StandardCharsets;
import java.util.List;

import parser.Parser;
import parser.Scanner;

public class Main {
    // We expect this binary to be used like "calculator '4 + 5 * 3'"
    public static void main(String[] args) {
        if (args.length != 1) {
            System.err.println("usage: calculator <expression>");
            System.exit(1);
        }

        try {
            System.out.println(evaluate(args[0]));
        } catch (IllegalStateException error) {
            System.err.println(error.getMessage());
            System.exit(1);
        }
    }

    // evaluate parses the expression and returns the number it stands for.
    private static int evaluate(String expression) {
        // The generated TokenSkipper will skip all whitespaces which the parser is not interested in.
        Scanner.TokenSkipper scanner = new Scanner.TokenSkipper(
                // The generated Scanner will convert the input into tokens. The filePath argument is used in error
                // messages. It wants the bytes of the input, not the string, so that offsets are byte offsets.
                new Scanner(expression.getBytes(StandardCharsets.UTF_8), "expression"));

        // The expression is parsed by giving the generated parser the scanner to pull tokens from. We get the root
        // node of the parse tree and the errors the parse found.
        Parser.ParseResult result = new Parser().parse(scanner);
        if (!result.errors().isEmpty()) {
            throw new IllegalStateException(result.errors().get(0).message());
        }
        if (result.tree() == null) {
            throw new IllegalStateException("the expression could not be parsed");
        }

        // Traversing over the parse tree will calculate the result for us.
        return evaluateNode(result.tree());
    }

    // evaluateNode recursively evaluates an expression node from the parse tree.
    // The number of children encodes which grammar production was matched:
    //   - 1 child:  INTEGER literal
    //   - 2 children: unary minus ("-" expression)
    //   - 3 children: binary operation (expression OP expression) or grouping ("(" expression ")")
    private static int evaluateNode(Parser.ParseNode node) {
        // Each node has a symbol (the grammar symbol it represents), a lexeme (the raw bytes from the input, set for
        // terminal nodes), and children (sub-nodes).
        List<Parser.ParseNode> children = node.children();
        return switch (children.size()) {
            // expression: INTEGER
            case 1 -> Integer.parseInt(lexeme(children.get(0)));
            // expression: "-" expression
            case 2 -> -evaluateNode(children.get(1));
            case 3 -> evaluateThreeChildren(node);
            default -> throw new IllegalStateException("unexpected node structure");
        };
    }

    // evaluateThreeChildren evaluates the two productions with three symbols on the right hand side.
    private static int evaluateThreeChildren(Parser.ParseNode node) {
        // In "(" expression ")", the middle child is the nonterminal expression node.
        // In "expression OP expression", the middle child is a terminal operator token.
        // We use this to distinguish the two cases.
        Parser.ParseNode middle = node.children().get(1);
        if (!(middle.symbol() instanceof Parser.TerminalSymbol operator)) {
            // expression: "(" expression ")"
            return evaluateNode(middle);
        }

        int leftValue = evaluateNode(node.children().get(0));
        int rightValue = evaluateNode(node.children().get(2));

        return switch (operator.token()) {
            // expression: expression "+" expression
            case TOKEN_PLUS -> leftValue + rightValue;
            // expression: expression "-" expression
            case TOKEN_MINUS -> leftValue - rightValue;
            // expression: expression "*" expression
            case TOKEN_MULTIPLY -> leftValue * rightValue;
            // expression: expression "/" expression
            case TOKEN_DIVIDE -> {
                if (rightValue == 0) {
                    throw new IllegalStateException("division by zero");
                }
                yield leftValue / rightValue;
            }
            default -> throw new IllegalStateException("unexpected node structure");
        };
    }

    // lexeme returns the bytes a terminal node stands for as text.
    private static String lexeme(Parser.ParseNode node) {
        return StandardCharsets.UTF_8.decode(node.lexeme()).toString();
    }
}
