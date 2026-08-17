mod parser;

use parser::parser::{ParseNode, ParseSymbol, Parser};
use parser::scanner::{Scanner, Token, TokenSkipper};
use std::process;

fn main() {
    // We expect this binary to be used like "calculator '4 + 5 * 3'"
    let args: Vec<String> = std::env::args().collect();
    if args.len() != 2 {
        eprintln!("usage: calculator <expression>");
        process::exit(1);
    }

    match evaluate(&args[1]) {
        Ok(value) => println!("{value}"),
        Err(message) => {
            eprintln!("{message}");
            process::exit(1);
        }
    }
}

// evaluate parses the expression and returns the number it stands for.
fn evaluate(expression: &str) -> Result<i32, String> {
    // The generated TokenSkipper will skip all whitespaces which the parser is not interested in.
    let mut scanner = TokenSkipper::new(
        // The generated Scanner will convert the input into tokens. The file_path argument is used in error messages.
        // It wants the bytes of the input, not the string, so that offsets are byte offsets.
        Scanner::new(expression.as_bytes(), "expression"),
    );

    // The expression is parsed by giving the generated parser the scanner to pull tokens from. We get the root node of
    // the parse tree and the errors the parse found.
    let result = Parser::new().parse(&mut scanner);
    if let Some(error) = result.errors.first() {
        return Err(error.to_string());
    }
    let Some(tree) = result.tree else {
        return Err("the expression could not be parsed".to_string());
    };

    // Traversing over the parse tree will calculate the result for us.
    evaluate_node(&tree)
}

// evaluate_node recursively evaluates an expression node from the parse tree.
// The number of children encodes which grammar production was matched:
//   - 1 child:    INTEGER literal
//   - 2 children: unary minus ("-" expression)
//   - 3 children: binary operation (expression OP expression) or grouping ("(" expression ")")
fn evaluate_node(node: &ParseNode) -> Result<i32, String> {
    // Each node has a symbol (the grammar symbol it represents), a lexeme (the raw bytes from the input, set for
    // terminal nodes), and children (sub-nodes).
    match node.children.len() {
        // expression: INTEGER
        1 => lexeme(&node.children[0])
            .parse()
            .map_err(|_| "not an integer".to_string()),
        // expression: "-" expression
        2 => Ok(-evaluate_node(&node.children[1])?),
        3 => evaluate_three_children(node),
        _ => Err("unexpected node structure".to_string()),
    }
}

// evaluate_three_children evaluates the two productions with three symbols on the right hand side.
fn evaluate_three_children(node: &ParseNode) -> Result<i32, String> {
    // In "(" expression ")", the middle child is the nonterminal expression node.
    // In "expression OP expression", the middle child is a terminal operator token.
    // We use this to distinguish the two cases.
    let middle = &node.children[1];
    let ParseSymbol::Terminal(operator) = middle.symbol else {
        // expression: "(" expression ")"
        return evaluate_node(middle);
    };

    let left_value = evaluate_node(&node.children[0])?;
    let right_value = evaluate_node(&node.children[2])?;

    match operator {
        // expression: expression "+" expression
        Token::TokenPlus => Ok(left_value + right_value),
        // expression: expression "-" expression
        Token::TokenMinus => Ok(left_value - right_value),
        // expression: expression "*" expression
        Token::TokenMultiply => Ok(left_value * right_value),
        // expression: expression "/" expression
        Token::TokenDivide => {
            if right_value == 0 {
                return Err("division by zero".to_string());
            }
            Ok(left_value / right_value)
        }
        _ => Err("unexpected node structure".to_string()),
    }
}

// lexeme returns the bytes a terminal node stands for as text.
fn lexeme(node: &ParseNode) -> String {
    String::from_utf8_lossy(node.lexeme).into_owned()
}
