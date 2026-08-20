import sys

from parser.parser import ParseNode, Parser, TerminalSymbol
from parser.scanner import Scanner, Token, TokenSkipper


def evaluate(expression: str) -> int:
    """Parses the expression and returns the number it stands for."""
    # The generated TokenSkipper will skip all whitespaces which the parser is not interested in.
    scanner = TokenSkipper(
        # The generated Scanner will convert the input into tokens. The file_path argument is used in error messages.
        # It wants the bytes of the input, not the string, so that offsets are byte offsets.
        Scanner(expression.encode(), "expression"),
    )

    # The expression is parsed by giving the generated parser the scanner to pull tokens from. We get the root node of
    # the parse tree and the errors the parse found.
    result = Parser().parse(scanner)
    if result.errors:
        raise result.errors[0]
    if result.tree is None:
        raise ValueError("the expression could not be parsed")

    # Traversing over the parse tree will calculate the result for us.
    return evaluate_node(result.tree)


def evaluate_node(node: ParseNode) -> int:
    """Recursively evaluates an expression node from the parse tree.

    The number of children encodes which grammar production was matched:
      - 1 child:    INTEGER literal
      - 2 children: unary minus ("-" expression)
      - 3 children: binary operation (expression OP expression) or grouping ("(" expression ")")
    """
    # Each node has a symbol (the grammar symbol it represents), a lexeme (the raw bytes from the input, set for
    # terminal nodes), and children (sub-nodes).
    match len(node.children):
        case 1:
            # expression: INTEGER
            return int(node.children[0].lexeme)
        case 2:
            # expression: "-" expression
            return -evaluate_node(node.children[1])
        case 3:
            return evaluate_three_children(node)
    raise ValueError("unexpected node structure")


def evaluate_three_children(node: ParseNode) -> int:
    """Evaluates the two productions with three symbols on the right hand side."""
    # In "(" expression ")", the middle child is the nonterminal expression node.
    # In "expression OP expression", the middle child is a terminal operator token.
    # We use this to distinguish the two cases.
    middle = node.children[1]
    if not isinstance(middle.symbol, TerminalSymbol):
        # expression: "(" expression ")"
        return evaluate_node(middle)

    left_value = evaluate_node(node.children[0])
    right_value = evaluate_node(node.children[2])

    match middle.symbol.token:
        case Token.TOKEN_PLUS:
            # expression: expression "+" expression
            return left_value + right_value
        case Token.TOKEN_MINUS:
            # expression: expression "-" expression
            return left_value - right_value
        case Token.TOKEN_MULTIPLY:
            # expression: expression "*" expression
            return left_value * right_value
        case Token.TOKEN_DIVIDE:
            # expression: expression "/" expression
            if right_value == 0:
                raise ZeroDivisionError("division by zero")
            # Python's // rounds towards negative infinity, so the sign is taken out of the division to truncate
            # towards zero the way the other languages of this example do.
            quotient = abs(left_value) // abs(right_value)
            if (left_value < 0) != (right_value < 0):
                return -quotient
            return quotient
    raise ValueError("unexpected node structure")


def main() -> None:
    # We expect this to be used like "calculator '4 + 5 * 3'"
    if len(sys.argv) != 2:
        print("usage: calculator <expression>", file=sys.stderr)
        sys.exit(1)

    try:
        print(evaluate(sys.argv[1]))
    except Exception as error:
        print(error, file=sys.stderr)
        sys.exit(1)


main()
