/* The C runner for the backend test corpus. It reads an input file, runs the generated scanner and the generated
   parser over it, and writes the canonical scanner trace, parser trace and tree trace the harness diffs against.

   This file has no dependencies beyond the standard library, and it must not grow any. It runs in the image with no
   network, which is what proves that generated GoLR code needs nothing but the bare language.

   The generated scanner and parser are headers which carry their implementation behind a macro, so defining both here
   compiles everything into this one translation unit and one command covers all of it. */

#define PARSER_SCANNER_IMPLEMENTATION
#define PARSER_PARSER_IMPLEMENTATION

#include "parser.h"
#include "scanner.h"

#include <stdio.h>
#include <stdlib.h>

static const char *const SCANNER_TRACE_FILE_NAME = "scanner.actual";
static const char *const PARSER_TRACE_FILE_NAME = "parser.actual";
static const char *const TREE_TRACE_FILE_NAME = "tree.actual";

/* The bytes a trace line carries as they are. Everything outside of it is escaped. */
#define PRINTABLE_LOW 0x20
#define PRINTABLE_HIGH 0x7e

/* Writes the bytes of a lexeme escaped. The caller writes the quotes around the result.

   The lexeme is escaped byte by byte, never character by character, so a multi byte UTF-8 sequence becomes one \xHH
   escape per byte. Decoding it first would report character offsets and disagree with every other backend. */
static void write_escaped_lexeme(FILE *out, ParserStringView lexeme) {
    size_t idx;
    for (idx = 0; idx < lexeme.length; idx++) {
        unsigned char value = (unsigned char)lexeme.data[idx];
        switch (value) {
        case '\\':
            fputs("\\\\", out);
            break;
        case '"':
            fputs("\\\"", out);
            break;
        case '\n':
            fputs("\\n", out);
            break;
        case '\r':
            fputs("\\r", out);
            break;
        case '\t':
            fputs("\\t", out);
            break;
        default:
            if (PRINTABLE_LOW <= value && value <= PRINTABLE_HIGH) {
                fputc((int)value, out);
                break;
            }
            /* The trace asks for a zero padded pair of lower case hex digits. */
            fprintf(out, "\\x%02x", value);
            break;
        }
    }
}

/* Scans the whole input and writes one line per event: the position the token or the failed match starts at, a
   keyword, and for a token its rule and lexeme, for a failed match the bytes it could not match. */
static void write_scanner_trace(FILE *out, const char *source, size_t source_length, const char *input_path) {
    /* The plain scanner and not the token skipper: a skipped rule matched like any other, and the position of the
       tokens around it is only checkable when it is in the trace. */
    ParserScanner scanner;
    ParserPosition position;
    char location[48];

    parser_scanner_init(&scanner, source, source_length, input_path);

    while (parser_scanner_next(&scanner)) {
        ParserStringView lexeme = parser_scanner_lexeme(&scanner);

        position = parser_scanner_position(&scanner, parser_scanner_byte_offset(&scanner));
        snprintf(location, sizeof(location), "%zu:%zu", position.line, position.column);

        if (parser_scanner_token(&scanner) == PARSER_TOKEN_INVALID_TOKEN) {
            fprintf(out, "%-7s %-7s \"", location, "ERROR");
            write_escaped_lexeme(out, lexeme);
            fputs("\"\n", out);
            continue;
        }

        fprintf(out, "%-7s %-7s %s \"", location, "TOKEN", parser_token_to_string(parser_scanner_token(&scanner)));
        write_escaped_lexeme(out, lexeme);
        fputs("\"\n", out);
    }

    /* The position after the scanner ran out of input, which is one past the last byte only when it consumed
       everything. It is the offset every off by one in a line table lands on. */
    position = parser_scanner_position(&scanner, parser_scanner_byte_offset(&scanner));
    snprintf(location, sizeof(location), "%zu:%zu", position.line, position.column);
    fprintf(out, "%-7s %s\n", location, "EOF");

    parser_scanner_free(&scanner);
}

/* Writes one trace line to the file behind the context pointer. This is the parser's trace hook. */
static void write_trace_line(void *context, const char *line) {
    fprintf((FILE *)context, "%s\n", line);
}

/* Parses the whole input and writes the line the parser's trace hook emits for every action. The hook reports the error
   recovery steps too, which the returned tree does not, so the tree and the errors are ignored. */
static void write_parser_trace(FILE *out, const char *source, size_t source_length, const char *input_path) {
    /* The token skipper here, because a skipped rule never reaches the parser. */
    ParserScanner scanner;
    ParserTokenSkipper skipper;
    ParserTokenSource source_of_tokens;
    ParserParser parser;
    ParserParseResult result;

    parser_scanner_init(&scanner, source, source_length, input_path);
    parser_token_skipper_init(&skipper, parser_scanner_as_token_source(&scanner));
    source_of_tokens = parser_token_skipper_as_token_source(&skipper);

    parser_parser_init(&parser);
    parser.trace = write_trace_line;
    parser.trace_context = out;

    result = parser_parser_parse(&parser, &source_of_tokens);

    parser_parse_result_free(&result);
    parser_parser_free(&parser);
    parser_scanner_free(&scanner);
}

/* Names a terminal for a trace line, giving the three tokens the grammar cannot spell a dollar name. */
static const char *terminal_trace_name(ParserToken terminal) {
    switch (terminal) {
    case PARSER_TOKEN_END_TOKEN:
        return "$end";
    case PARSER_TOKEN_ERROR_TOKEN:
        return "$error";
    case PARSER_TOKEN_INVALID_TOKEN:
        return "$invalid";
    default:
        return parser_token_to_string(terminal);
    }
}

/* The bare grammar name of a symbol, which for a nonterminal is what parser_nonterminal_to_string returns and for a
   terminal is the name the traces spell it with. */
static const char *symbol_trace_name(const ParserParseSymbol *symbol) {
    ParserToken terminal;
    ParserNonterminal nonterminal;

    if (parser_parse_symbol_terminal(symbol, &terminal)) {
        return terminal_trace_name(terminal);
    }
    parser_parse_symbol_nonterminal(symbol, &nonterminal);
    return parser_nonterminal_to_string(nonterminal);
}

/* Writes the line of the given node and the lines of everything below it, which is the pre-order the tree trace is
   read in: a node, then what it was built from. The payload carries the indentation and the position and span columns
   do not, so they stay in the same place however deep a node sits. */
static void write_tree_node(FILE *out, ParserScanner *scanner, const ParserParseNode *node, size_t depth) {
    ParserPosition position = parser_scanner_position(scanner, node->byte_offset);
    ParserToken terminal;
    ParserNonterminal nonterminal;
    char location[48];
    char span[48];
    size_t idx;

    snprintf(location, sizeof(location), "%zu:%zu", position.line, position.column);
    snprintf(span, sizeof(span), "%zu+%zu", node->byte_offset, node->byte_length);
    fprintf(out, "%-7s %-7s ", location, span);
    for (idx = 0; idx < depth; idx++) {
        fputs("  ", out);
    }

    if (!parser_parse_symbol_terminal(&node->symbol, &terminal)) {
        parser_parse_symbol_nonterminal(&node->symbol, &nonterminal);
        /* The node is named the way the REDUCE line of a parser trace names the production it was reduced from. */
        fprintf(out, "%s =>", parser_nonterminal_to_string(nonterminal));
        if (node->child_count == 0) {
            fputs(" ε", out);
        }
        for (idx = 0; idx < node->child_count; idx++) {
            fprintf(out, " %s", symbol_trace_name(&node->children[idx].symbol));
        }
        fputc('\n', out);
    } else if (terminal == PARSER_TOKEN_ERROR_TOKEN) {
        /* The error node stands for no token of its own, so its span is all it carries. */
        fprintf(out, "%s\n", terminal_trace_name(terminal));
    } else {
        /* The text is read off the source through the span and never carried along from the token, which is what makes
           the trace state that the span is right. */
        fprintf(out, "%s \"", terminal_trace_name(terminal));
        write_escaped_lexeme(out, parser_scanner_text(scanner, node->byte_offset, node->byte_length));
        fputs("\"\n", out);
    }

    for (idx = 0; idx < node->child_count; idx++) {
        write_tree_node(out, scanner, &node->children[idx], depth + 1);
    }
}

/* Parses the whole input and writes one line per node of the tree the parse built, in pre-order. A parse which was
   given up builds no tree and writes nothing, which is the empty trace the harness expects for it. */
static void write_tree_trace(FILE *out, const char *source, size_t source_length, const char *input_path) {
    /* The scanner stays at hand after the parse, because a node carries the span of the source it covers and not the
       source itself, so the trace resolves every node through the position and the text of the scanner. */
    ParserScanner scanner;
    ParserTokenSkipper skipper;
    ParserTokenSource source_of_tokens;
    ParserParser parser;
    ParserParseResult result;

    parser_scanner_init(&scanner, source, source_length, input_path);
    parser_token_skipper_init(&skipper, parser_scanner_as_token_source(&scanner));
    source_of_tokens = parser_token_skipper_as_token_source(&skipper);

    parser_parser_init(&parser);
    result = parser_parser_parse(&parser, &source_of_tokens);

    if (result.tree != NULL) {
        write_tree_node(out, &scanner, result.tree, 0);
    }

    parser_parse_result_free(&result);
    parser_parser_free(&parser);
    parser_scanner_free(&scanner);
}

/* Opens a trace file, or reports why it could not be opened. */
static FILE *open_trace(const char *file_name) {
    FILE *out = fopen(file_name, "wb");
    if (out == NULL) {
        fprintf(stderr, "writing %s failed\n", file_name);
    }
    return out;
}

int main(int argc, char **argv) {
    const char *input_path;
    FILE *in;
    FILE *out;
    long size;
    size_t source_length;
    char *source;

    if (argc != 2) {
        fprintf(stderr, "the runner takes the input file as its only argument\n");
        return 1;
    }
    input_path = argv[1];

    /* The bytes and not decoded text, because the generated scanner wants the bytes. Handing it text would be the
       classic mistake this harness exists to catch, which is why the stream is binary. */
    in = fopen(input_path, "rb");
    if (in == NULL) {
        fprintf(stderr, "reading %s failed\n", input_path);
        return 1;
    }
    if (fseek(in, 0, SEEK_END) != 0 || (size = ftell(in)) < 0 || fseek(in, 0, SEEK_SET) != 0) {
        fprintf(stderr, "reading %s failed\n", input_path);
        fclose(in);
        return 1;
    }
    source_length = (size_t)size;

    /* One byte more than the input, so that an empty input still gets an allocation which is not null. */
    source = (char *)malloc(source_length + 1);
    if (source == NULL) {
        fprintf(stderr, "reading %s failed: out of memory\n", input_path);
        fclose(in);
        return 1;
    }
    if (source_length > 0 && fread(source, 1, source_length, in) != source_length) {
        fprintf(stderr, "reading %s failed\n", input_path);
        fclose(in);
        free(source);
        return 1;
    }
    fclose(in);

    /* Each trace is written on its own, so a scanner which breaks still lets the parser trace be written. A case
       failing all three traces has to stay distinguishable from one failing only the last. */
    out = open_trace(SCANNER_TRACE_FILE_NAME);
    if (out != NULL) {
        write_scanner_trace(out, source, source_length, input_path);
        fclose(out);
    }

    out = open_trace(PARSER_TRACE_FILE_NAME);
    if (out != NULL) {
        write_parser_trace(out, source, source_length, input_path);
        fclose(out);
    }

    out = open_trace(TREE_TRACE_FILE_NAME);
    if (out != NULL) {
        write_tree_trace(out, source, source_length, input_path);
        fclose(out);
    }

    free(source);
    return 0;
}
