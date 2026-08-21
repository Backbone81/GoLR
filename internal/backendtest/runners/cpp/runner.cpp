// The C++ runner for the backend test corpus. It reads an input file, runs the generated scanner and the generated
// parser over it, and writes the canonical scanner trace and parser trace the harness diffs against.
//
// This file has no dependencies beyond the standard library, and it must not grow any. It runs in the image with no
// network, which is what proves that generated GoLR code needs nothing but the bare language.
//
// The generated scanner and parser are headers, so compiling this file compiles them too and one command covers both.

#include <cstdio>
#include <exception>
#include <fstream>
#include <iterator>
#include <optional>
#include <string>
#include <string_view>
#include <vector>

#include "parser.hpp"
#include "scanner.hpp"

namespace {

const char* const SCANNER_TRACE_FILE_NAME = "scanner.actual";
const char* const PARSER_TRACE_FILE_NAME = "parser.actual";

// The bytes a trace line carries as they are. Everything outside of it is escaped.
constexpr unsigned char PRINTABLE_LOW = 0x20;
constexpr unsigned char PRINTABLE_HIGH = 0x7e;

// escape_lexeme escapes the bytes of a lexeme. The caller writes the quotes around the result.
//
// The lexeme is escaped byte by byte, never character by character, so a multi byte UTF-8 sequence becomes one \xHH
// escape per byte. Decoding it first would report character offsets and disagree with every other backend.
std::string escape_lexeme(std::string_view lexeme) {
    std::string result;
    for (const char raw : lexeme) {
        const unsigned char value = static_cast<unsigned char>(raw);
        switch (value) {
        case '\\':
            result += "\\\\";
            break;
        case '"':
            result += "\\\"";
            break;
        case '\n':
            result += "\\n";
            break;
        case '\r':
            result += "\\r";
            break;
        case '\t':
            result += "\\t";
            break;
        default:
            if (PRINTABLE_LOW <= value && value <= PRINTABLE_HIGH) {
                result += raw;
                break;
            }
            // The trace asks for a zero padded pair of lower case hex digits.
            char escape[5];
            std::snprintf(escape, sizeof(escape), "\\x%02x", value);
            result += escape;
            break;
        }
    }
    return result;
}

// append_scanner_trace scans the whole input and appends one line per event.
void append_scanner_trace(std::vector<std::string>& lines, std::string_view source, const std::string& input_path) {
    // The plain Scanner and not the TokenSkipper: a skipped rule matched like any other, and the offsets of the tokens
    // around it are only checkable when it is in the trace.
    parser::Scanner scanner(source, input_path);

    while (scanner.next()) {
        // For a failed match this is the start of the attempt, not the byte which could not be consumed.
        const std::size_t start = scanner.byte_offset();

        if (scanner.token() == parser::Token::InvalidToken) {
            lines.push_back("ERROR " + std::to_string(start));
            continue;
        }

        // The end comes from the length of the lexeme rather than from an offset the scanner reports, so that a lexeme
        // which disagrees with the offsets cannot pass unnoticed.
        const std::string_view lexeme = scanner.lexeme();
        const std::size_t end = start + lexeme.size();
        lines.push_back("TOKEN " + std::string(parser::to_string(scanner.token())) + " " + std::to_string(start) +
                        " " + std::to_string(end) + " \"" + escape_lexeme(lexeme) + "\"");
    }

    // The offset the scanner reports after it ran out of input, not the length of the source. The two agree only when
    // the scanner consumed everything.
    lines.push_back("EOF " + std::to_string(scanner.byte_offset()));
}

// append_node_trace appends the post-order walk of the subtree, which is the shift and reduce sequence of the parse.
void append_node_trace(std::vector<std::string>& lines, const parser::ParseNode& node) {
    for (const parser::ParseNode& child : node.children) {
        append_node_trace(lines, child);
    }

    if (const std::optional<parser::Token> terminal = node.symbol.terminal()) {
        if (*terminal == parser::Token::ErrorToken) {
            // The leaf the recovery pushed where it resumed. It stands for the dropped input and names no token.
            lines.push_back("RESYNC");
            return;
        }
        lines.push_back("SHIFT " + std::string(parser::to_string(*terminal)));
        return;
    }

    lines.push_back("REDUCE " + std::string(parser::to_string(*node.symbol.nonterminal())) + " " +
                    std::to_string(node.children.size()));
}

// append_parser_trace parses the whole input and appends its errors, the walk of the tree and the accept event.
//
// The errors come first because a shift carries no offset to interleave them by. A parse which was given up returns no
// tree, so its trace is the errors alone and the missing accept event is what says the two outcomes apart.
void append_parser_trace(std::vector<std::string>& lines, std::string_view source, const std::string& input_path) {
    // The TokenSkipper here, because a skipped rule never reaches the parser.
    parser::Scanner scanner(source, input_path);
    parser::TokenSkipper skipper(scanner);

    parser::Parser instance;
    const parser::ParseResult result = instance.parse(skipper);

    for (const parser::ParseError& error : result.errors) {
        lines.push_back("ERROR " + std::to_string(error.byte_offset));
    }

    if (!result.tree.has_value()) {
        return;
    }
    append_node_trace(lines, *result.tree);
    lines.push_back("ACCEPT");
}

// write_trace produces one trace and writes it to its file. Whatever was produced before an exception is written all
// the same, so a runner which breaks half way still says how far it got, and the other trace is still produced.
template <typename Produce>
void write_trace(const char* file_name, Produce produce) {
    std::vector<std::string> lines;

    try {
        produce(lines);
    } catch (const std::exception& error) {
        std::fprintf(stderr, "producing %s failed: %s\n", file_name, error.what());
    } catch (...) {
        std::fprintf(stderr, "producing %s failed\n", file_name);
    }

    // Every line is terminated, and an empty trace is an empty file rather than a bare newline. The stream is binary
    // and the line ending is spelled out, because the trace is LF whatever the platform would use.
    std::ofstream out(file_name, std::ios::binary);
    for (const std::string& line : lines) {
        out << line << '\n';
    }
    out.close();
    if (!out) {
        std::fprintf(stderr, "writing %s failed\n", file_name);
    }
}

}  // namespace

int main(int argc, char** argv) {
    if (argc != 2) {
        std::fprintf(stderr, "the runner takes the input file as its only argument\n");
        return 1;
    }
    const std::string input_path = argv[1];

    // The bytes and not a decoded string, because the generated scanner wants the bytes. Handing it text would be the
    // classic mistake this harness exists to catch, which is why the stream is binary.
    std::ifstream in(input_path, std::ios::binary);
    if (!in) {
        std::fprintf(stderr, "reading %s failed\n", input_path.c_str());
        return 1;
    }
    const std::string source((std::istreambuf_iterator<char>(in)), std::istreambuf_iterator<char>());

    write_trace(SCANNER_TRACE_FILE_NAME, [&](std::vector<std::string>& lines) {
        append_scanner_trace(lines, source, input_path);
    });
    write_trace(PARSER_TRACE_FILE_NAME, [&](std::vector<std::string>& lines) {
        append_parser_trace(lines, source, input_path);
    });
    return 0;
}
