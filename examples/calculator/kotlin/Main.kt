import java.nio.charset.StandardCharsets
import kotlin.system.exitProcess
import parser.ParseNode
import parser.Parser
import parser.Scanner
import parser.TerminalSymbol
import parser.Token
import parser.TokenSkipper

// We expect this binary to be used like "calculator '4 + 5 * 3'"
fun main(args: Array<String>) {
    if (args.size != 1) {
        System.err.println("usage: calculator <expression>")
        exitProcess(1)
    }

    try {
        println(evaluate(args[0]))
    } catch (error: IllegalStateException) {
        System.err.println(error.message)
        exitProcess(1)
    }
}

// evaluate parses the expression and returns the number it stands for.
private fun evaluate(expression: String): Int {
    // The generated TokenSkipper will skip all whitespaces which the parser is not interested in.
    val scanner = TokenSkipper(
        // The generated Scanner will convert the input into tokens. The filePath argument is used in error messages.
        // It wants the bytes of the input, not the string, so that offsets are byte offsets.
        Scanner(expression.toByteArray(), "expression"),
    )

    // The expression is parsed by giving the generated parser the scanner to pull tokens from. We get the root node of
    // the parse tree and the errors the parse found.
    val result = Parser().parse(scanner)
    result.errors.firstOrNull()?.let { parseError -> error(parseError.message) }

    // The tree is null when the parse could not be carried to its end, which the type system makes us handle here.
    val tree = result.tree ?: error("the expression could not be parsed")

    // Traversing over the parse tree will calculate the result for us.
    return evaluateNode(tree)
}

// evaluateNode recursively evaluates an expression node from the parse tree.
// The number of children encodes which grammar production was matched:
//   - 1 child:  INTEGER literal
//   - 2 children: unary minus ("-" expression)
//   - 3 children: binary operation (expression OP expression) or grouping ("(" expression ")")
private fun evaluateNode(node: ParseNode): Int =
    // Each node has a symbol (the grammar symbol it represents), a lexeme (the raw bytes from the input, set for
    // terminal nodes), and children (sub-nodes).
    when (node.children.size) {
        // expression: INTEGER
        1 -> lexeme(node.children[0]).toInt()
        // expression: "-" expression
        2 -> -evaluateNode(node.children[1])
        3 -> evaluateThreeChildren(node)
        else -> error("unexpected node structure")
    }

// evaluateThreeChildren evaluates the two productions with three symbols on the right hand side.
private fun evaluateThreeChildren(node: ParseNode): Int {
    // In "(" expression ")", the middle child is the nonterminal expression node.
    // In "expression OP expression", the middle child is a terminal operator token.
    // We use this to distinguish the two cases.
    val middle = node.children[1]
    // expression: "(" expression ")"
    val operator = middle.symbol as? TerminalSymbol ?: return evaluateNode(middle)

    val leftValue = evaluateNode(node.children[0])
    val rightValue = evaluateNode(node.children[2])

    return when (operator.token) {
        // expression: expression "+" expression
        Token.TOKEN_PLUS -> leftValue + rightValue
        // expression: expression "-" expression
        Token.TOKEN_MINUS -> leftValue - rightValue
        // expression: expression "*" expression
        Token.TOKEN_MULTIPLY -> leftValue * rightValue
        // expression: expression "/" expression
        Token.TOKEN_DIVIDE -> {
            if (rightValue == 0) {
                error("division by zero")
            }
            leftValue / rightValue
        }
        else -> error("unexpected node structure")
    }
}

// lexeme returns the bytes a terminal node stands for as text.
private fun lexeme(node: ParseNode): String = StandardCharsets.UTF_8.decode(node.lexeme).toString()
