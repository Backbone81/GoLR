#include <iostream>
#include <stdexcept>
#include <string>
#include <variant>

#include "parser/parser.hpp"

namespace {

using calculator::parser::ParseNode;
using calculator::parser::ParseResult;
using calculator::parser::Parser;
using calculator::parser::Scanner;
using calculator::parser::Token;
using calculator::parser::TokenSkipper;

long evaluate_node(const ParseNode& node);

// Evaluates the two productions with three symbols on the right hand side.
long evaluate_three_children(const ParseNode& node) {
    // In "(" expression ")", the middle child is the nonterminal expression node.
    // In "expression OP expression", the middle child is a terminal operator token.
    // We use this to distinguish the two cases.
    const ParseNode& middle = node.children[1];
    const Token* operation = std::get_if<Token>(&middle.symbol);
    if (operation == nullptr) {
        // expression: "(" expression ")"
        return evaluate_node(middle);
    }

    const long left_value = evaluate_node(node.children[0]);
    const long right_value = evaluate_node(node.children[2]);

    switch (*operation) {
    case Token::TokenPlus:
        // expression: expression "+" expression
        return left_value + right_value;
    case Token::TokenMinus:
        // expression: expression "-" expression
        return left_value - right_value;
    case Token::TokenMultiply:
        // expression: expression "*" expression
        return left_value * right_value;
    case Token::TokenDivide:
        // expression: expression "/" expression
        if (right_value == 0) {
            throw std::runtime_error("division by zero");
        }
        return left_value / right_value;
    default:
        throw std::runtime_error("unexpected node structure");
    }
}

// Recursively evaluates an expression node from the parse tree.
//
// The number of children encodes which grammar production was matched:
//   - 1 child:    INTEGER literal
//   - 2 children: unary minus ("-" expression)
//   - 3 children: binary operation (expression OP expression) or grouping ("(" expression ")")
long evaluate_node(const ParseNode& node) {
    // Each node has a symbol (the grammar symbol it represents), a lexeme (the raw bytes from the input, set for
    // terminal nodes), and children (sub-nodes).
    switch (node.children.size()) {
    case 1:
        // expression: INTEGER
        return std::stol(std::string(node.children[0].lexeme));
    case 2:
        // expression: "-" expression
        return -evaluate_node(node.children[1]);
    case 3:
        return evaluate_three_children(node);
    default:
        throw std::runtime_error("unexpected node structure");
    }
}

// Parses the expression and returns the number it stands for.
long evaluate(const std::string& expression) {
    // The generated Scanner will convert the input into tokens. The file path argument is used in error messages. It
    // wants the bytes of the input, so that offsets are byte offsets.
    Scanner scanner(expression, "expression");

    // The generated TokenSkipper will skip all whitespaces which the parser is not interested in.
    TokenSkipper skipper(scanner);

    // The expression is parsed by giving the generated parser the scanner to pull tokens from. We get the root node of
    // the parse tree and the errors the parse found.
    Parser parser;
    const ParseResult result = parser.parse(skipper);
    if (!result.errors.empty()) {
        throw std::runtime_error(result.errors.front().message());
    }
    if (!result.tree.has_value()) {
        throw std::runtime_error("the expression could not be parsed");
    }

    // Traversing over the parse tree will calculate the result for us.
    return evaluate_node(*result.tree);
}

}  // namespace

int main(int argc, char** argv) {
    // We expect this to be used like "calculator '4 + 5 * 3'"
    if (argc != 2) {
        std::cerr << "usage: calculator <expression>\n";
        return 1;
    }

    try {
        std::cout << evaluate(argv[1]) << "\n";
    } catch (const std::exception& error) {
        std::cerr << error.what() << "\n";
        return 1;
    }
    return 0;
}
