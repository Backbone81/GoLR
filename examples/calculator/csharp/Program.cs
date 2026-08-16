using System;
using System.Globalization;
using System.Text;

using Calculator.Parser;

internal static class Program
{
    // We expect this to be used like "calculator '4 + 5 * 3'"
    private static int Main(string[] args)
    {
        if (args.Length != 1)
        {
            Console.Error.WriteLine("usage: calculator <expression>");
            return 1;
        }

        try
        {
            Console.WriteLine(Evaluate(args[0]).ToString(CultureInfo.InvariantCulture));
            return 0;
        }
        catch (InvalidOperationException error)
        {
            Console.Error.WriteLine(error.Message);
            return 1;
        }
    }

    // Evaluate parses the expression and returns the number it stands for.
    private static int Evaluate(string expression)
    {
        // The generated TokenSkipper will skip all whitespaces which the parser is not interested in.
        TokenSkipper scanner = new TokenSkipper(
            // The generated Scanner will convert the input into tokens. The filePath argument is used in error
            // messages. It wants the bytes of the input, not the string, so that offsets are byte offsets.
            new Scanner(Encoding.UTF8.GetBytes(expression), "expression"));

        // The expression is parsed by giving the generated parser the scanner to pull tokens from. We get the root
        // node of the parse tree and the errors the parse found.
        ParseResult result = new Parser().Parse(scanner);
        if (result.Errors.Count != 0)
        {
            throw new InvalidOperationException(result.Errors[0].Message);
        }
        if (result.Tree == null)
        {
            throw new InvalidOperationException("the expression could not be parsed");
        }

        // Traversing over the parse tree will calculate the result for us.
        return EvaluateNode(result.Tree);
    }

    // EvaluateNode recursively evaluates an expression node from the parse tree.
    // The number of children encodes which grammar production was matched:
    //   - 1 child:  INTEGER literal
    //   - 2 children: unary minus ("-" expression)
    //   - 3 children: binary operation (expression OP expression) or grouping ("(" expression ")")
    private static int EvaluateNode(ParseNode node)
    {
        // Each node has a symbol (the grammar symbol it represents), a lexeme (the raw bytes from the input, set for
        // terminal nodes), and children (sub-nodes).
        switch (node.Children.Count)
        {
            case 1:
                // expression: INTEGER
                return int.Parse(
                    Encoding.UTF8.GetString(node.Children[0].Lexeme.Span), CultureInfo.InvariantCulture);
            case 2:
                // expression: "-" expression
                return -EvaluateNode(node.Children[1]);
            case 3:
                return EvaluateThreeChildren(node);
            default:
                throw new InvalidOperationException("unexpected node structure");
        }
    }

    // EvaluateThreeChildren evaluates the two productions with three symbols on the right hand side.
    private static int EvaluateThreeChildren(ParseNode node)
    {
        // In "(" expression ")", the middle child is the nonterminal expression node.
        // In "expression OP expression", the middle child is a terminal operator token.
        // We use this to distinguish the two cases.
        ParseNode middle = node.Children[1];
        if (!middle.Symbol.TryGetTerminal(out Token operatorToken))
        {
            // expression: "(" expression ")"
            return EvaluateNode(middle);
        }

        int leftValue = EvaluateNode(node.Children[0]);
        int rightValue = EvaluateNode(node.Children[2]);

        switch (operatorToken)
        {
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
                if (rightValue == 0)
                {
                    throw new InvalidOperationException("division by zero");
                }
                return leftValue / rightValue;
            default:
                throw new InvalidOperationException("unexpected node structure");
        }
    }
}
