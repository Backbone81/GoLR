/* The C runner for the backend test corpus. It reads an input file, runs the generated scanner and the generated
   parser over it, and writes the canonical scanner trace and parser trace the harness diffs against.

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

/* Scans the whole input and writes one line per event. */
static void write_scanner_trace(FILE *out, const char *source, size_t source_length, const char *input_path) {
    /* The plain scanner and not the token skipper: a skipped rule matched like any other, and the offsets of the
       tokens around it are only checkable when it is in the trace. */
    ParserScanner scanner;
    parser_scanner_init(&scanner, source, source_length, input_path);

    while (parser_scanner_next(&scanner)) {
        /* For a failed match this is the start of the attempt, not the byte which could not be consumed. */
        size_t start = parser_scanner_byte_offset(&scanner);
        ParserStringView lexeme;

        if (parser_scanner_token(&scanner) == PARSER_TOKEN_INVALID_TOKEN) {
            fprintf(out, "ERROR %zu\n", start);
            continue;
        }

        /* The end comes from the length of the lexeme rather than from an offset the scanner reports, so that a lexeme
           which disagrees with the offsets cannot pass unnoticed. */
        lexeme = parser_scanner_lexeme(&scanner);
        fprintf(out, "TOKEN %s %zu %zu \"", parser_token_to_string(parser_scanner_token(&scanner)), start,
                start + lexeme.length);
        write_escaped_lexeme(out, lexeme);
        fputs("\"\n", out);
    }

    /* The offset the scanner reports after it ran out of input, not the length of the source. The two agree only when
       the scanner consumed everything. */
    fprintf(out, "EOF %zu\n", parser_scanner_byte_offset(&scanner));
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
       failing both traces has to stay distinguishable from one failing only the scanner. */
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

    free(source);
    return 0;
}
