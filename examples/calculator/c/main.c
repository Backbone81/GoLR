/* The generated scanner and parser are headers which carry their implementation behind a macro. Defining both here
   emits the tables and the function bodies into this translation unit, which is the one place they belong. */
#define CALCULATOR_SCANNER_IMPLEMENTATION
#define CALCULATOR_PARSER_IMPLEMENTATION

#include "parser/parser.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static bool evaluate_node(const CalculatorParseNode *node, long *value, const char **error);

/* Evaluates the two productions with three symbols on the right hand side. */
static bool evaluate_three_children(const CalculatorParseNode *node, long *value, const char **error) {
    /* In "(" expression ")", the middle child is the nonterminal expression node.
       In "expression OP expression", the middle child is a terminal operator token.
       We use this to distinguish the two cases. */
    const CalculatorParseNode *middle = &node->children[1];
    CalculatorToken operation;
    long left_value;
    long right_value;

    if (!calculator_parse_symbol_terminal(&middle->symbol, &operation)) {
        /* expression: "(" expression ")" */
        return evaluate_node(middle, value, error);
    }

    if (!evaluate_node(&node->children[0], &left_value, error)) {
        return false;
    }
    if (!evaluate_node(&node->children[2], &right_value, error)) {
        return false;
    }

    switch (operation) {
    case CALCULATOR_TOKEN_PLUS:
        /* expression: expression "+" expression */
        *value = left_value + right_value;
        return true;
    case CALCULATOR_TOKEN_MINUS:
        /* expression: expression "-" expression */
        *value = left_value - right_value;
        return true;
    case CALCULATOR_TOKEN_MULTIPLY:
        /* expression: expression "*" expression */
        *value = left_value * right_value;
        return true;
    case CALCULATOR_TOKEN_DIVIDE:
        /* expression: expression "/" expression */
        if (right_value == 0) {
            *error = "division by zero";
            return false;
        }
        *value = left_value / right_value;
        return true;
    default:
        *error = "unexpected node structure";
        return false;
    }
}

/* Recursively evaluates an expression node from the parse tree.

   The number of children encodes which grammar production was matched:
     - 1 child:    INTEGER literal
     - 2 children: unary minus ("-" expression)
     - 3 children: binary operation (expression OP expression) or grouping ("(" expression ")") */
static bool evaluate_node(const CalculatorParseNode *node, long *value, const char **error) {
    /* Each node has a symbol (the grammar symbol it represents), a lexeme (the raw bytes from the input, set for
       terminal nodes), and children (sub-nodes). */
    char digits[32];
    long operand;

    switch (node->child_count) {
    case 1:
        /* expression: INTEGER. The lexeme is a range of the input and is not terminated by a zero, so it is copied
           before it is read as a number. */
        snprintf(digits, sizeof(digits), "%.*s", (int)node->children[0].lexeme.length, node->children[0].lexeme.data);
        *value = strtol(digits, NULL, 10);
        return true;
    case 2:
        /* expression: "-" expression */
        if (!evaluate_node(&node->children[1], &operand, error)) {
            return false;
        }
        *value = -operand;
        return true;
    case 3:
        return evaluate_three_children(node, value, error);
    default:
        *error = "unexpected node structure";
        return false;
    }
}

/* Parses the expression and writes the number it stands for. */
static bool evaluate(const char *expression, size_t length, long *value, char *error_buffer, size_t error_size) {
    /* The generated scanner will convert the input into tokens. The file path argument is used in error messages. It
       wants the bytes of the input, so that offsets are byte offsets. */
    CalculatorScanner scanner;
    /* The generated token skipper will skip all whitespaces which the parser is not interested in. */
    CalculatorTokenSkipper skipper;
    CalculatorTokenSkipperScanner source;
    CalculatorParser parser;
    CalculatorParseResult result;
    const char *error = NULL;
    bool evaluated;

    calculator_scanner_init(&scanner, expression, length, "expression");
    calculator_token_skipper_init(&skipper, &scanner);
    source = calculator_token_skipper_as_token_skipper_scanner(&skipper);

    /* The expression is parsed by giving the generated parser the scanner to pull tokens from. We get the root node of
       the parse tree and the errors the parse found. */
    calculator_parser_init(&parser);
    result = calculator_parser_parse(&parser, &source);

    if (result.error_count > 0) {
        calculator_parse_error_message(&result.errors[0], error_buffer, error_size);
        calculator_parse_result_free(&result);
        calculator_parser_free(&parser);
        return false;
    }
    if (result.tree == NULL) {
        snprintf(error_buffer, error_size, "the expression could not be parsed");
        calculator_parse_result_free(&result);
        calculator_parser_free(&parser);
        return false;
    }

    /* Traversing over the parse tree will calculate the result for us. */
    evaluated = evaluate_node(result.tree, value, &error);
    if (!evaluated) {
        snprintf(error_buffer, error_size, "%s", error);
    }

    /* The tree belongs to the result, so it is read before the result is released. */
    calculator_parse_result_free(&result);
    calculator_parser_free(&parser);
    return evaluated;
}

int main(int argc, char **argv) {
    char error[256];
    long value;

    /* We expect this to be used like "calculator '4 + 5 * 3'" */
    if (argc != 2) {
        fprintf(stderr, "usage: calculator <expression>\n");
        return 1;
    }

    if (!evaluate(argv[1], strlen(argv[1]), &value, error, sizeof(error))) {
        fprintf(stderr, "%s\n", error);
        return 1;
    }
    printf("%ld\n", value);
    return 0;
}
