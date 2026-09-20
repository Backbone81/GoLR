// The C++ runner for the backend test corpus. It reads an input file, runs the generated scanner and the generated
// parser over it, and writes the canonical scanner trace, parser trace and tree trace the harness diffs against.
//
// This file has no dependencies beyond the standard library, and it must not grow any. It runs in the image with no
// network, which is what proves that generated GoLR code needs nothing but the bare language.
//
// The generated scanner and parser are headers, so compiling this file compiles them too and one command covers both.

#include <cstdio>
#include <cstdlib>
#include <exception>
#include <fstream>
#include <iterator>
#include <string>
#include <string_view>
#include <variant>
#include <vector>

#include "parser.hpp"
#include "scanner.hpp"

namespace {

const char* const SCANNER_TRACE_FILE_NAME = "scanner.actual";
const char* const PARSER_TRACE_FILE_NAME = "parser.actual";
const char* const TREE_TRACE_FILE_NAME = "tree.actual";

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

// pad_field left justifies a field to width 7, like the parser trace lines.
std::string pad_field(std::string field) {
    if (field.size() < 7) {
        field.resize(7, ' ');
    }
    return field;
}

// check_position holds the offset based position() and text() against the line(), column() and lexeme() of the token
// the scanner currently sits on. The two ways of asking exist side by side until the release which drops line() and
// column(), and the corpus is where they have to agree: every case of it is far more input than a hand written test
// covers.
void check_position(const parser::Scanner& scanner) {
    const parser::Position position = scanner.position(scanner.byte_offset());
    if (position.line != scanner.line() || position.column != scanner.column() ||
        position.file_path != scanner.file_path()) {
        std::fprintf(stderr, "position(%zu) is %.*s %zu:%zu, but the scanner reports %.*s %zu:%zu\n",
                     scanner.byte_offset(), static_cast<int>(position.file_path.size()), position.file_path.data(),
                     position.line, position.column, static_cast<int>(scanner.file_path().size()),
                     scanner.file_path().data(), scanner.line(), scanner.column());
        std::exit(1);
    }

    const std::string_view lexeme = scanner.lexeme();
    if (scanner.text(scanner.byte_offset(), lexeme.size()) != lexeme) {
        std::fprintf(stderr, "text(%zu, %zu) differs from the lexeme\n", scanner.byte_offset(), lexeme.size());
        std::exit(1);
    }
}

// append_scanner_trace scans the whole input and appends one line per event: the position the token or the failed
// match starts at, a keyword, and for a token its rule and lexeme, for a failed match the bytes it could not match.
void append_scanner_trace(std::vector<std::string>& lines, std::string_view source, const std::string& input_path) {
    // The plain Scanner and not the TokenSkipper: a skipped rule matched like any other, and the position of the
    // tokens around it is only checkable when it is in the trace.
    parser::Scanner scanner(source, input_path);

    while (scanner.next()) {
        check_position(scanner);

        const std::string location =
            pad_field(std::to_string(scanner.line()) + ":" + std::to_string(scanner.column()));
        const std::string lexeme = escape_lexeme(scanner.lexeme());

        if (scanner.token() == parser::Token::InvalidToken) {
            lines.push_back(location + " " + pad_field("ERROR") + " \"" + lexeme + "\"");
            continue;
        }
        lines.push_back(location + " " + pad_field("TOKEN") + " " + std::string(parser::to_string(scanner.token())) +
                        " \"" + lexeme + "\"");
    }

    // The position after the scanner ran out of input, which is one past the last byte only when it consumed
    // everything. It is the offset every off by one in a line table lands on, so it is checked like a token.
    check_position(scanner);
    lines.push_back(pad_field(std::to_string(scanner.line()) + ":" + std::to_string(scanner.column())) + " EOF");
}

// append_parser_trace parses the whole input and appends the line the parser's trace hook emits for every action. The
// hook reports the error recovery steps too, which the returned tree does not, so the tree and the error are ignored.
void append_parser_trace(std::vector<std::string>& lines, std::string_view source, const std::string& input_path) {
    parser::Scanner scanner(source, input_path);
    parser::TokenSkipper skipper(scanner);

    parser::Parser instance;
    instance.set_trace([&lines](std::string_view line) { lines.emplace_back(line); });

    // The TokenSkipper here, because a skipped rule never reaches the parser.
    static_cast<void>(instance.parse(skipper));
}

// terminal_trace_name names a terminal for a trace line, giving the three tokens the grammar cannot spell a dollar
// name.
std::string terminal_trace_name(parser::Token terminal) {
    switch (terminal) {
    case parser::Token::EndToken:
        return "$end";
    case parser::Token::ErrorToken:
        return "$error";
    case parser::Token::InvalidToken:
        return "$invalid";
    default:
        return std::string(parser::to_string(terminal));
    }
}

// symbol_trace_name is the bare grammar name of a symbol, which for a nonterminal is what to_string returns and for a
// terminal is the name the traces spell it with.
std::string symbol_trace_name(const parser::ParseSymbol& symbol) {
    if (const parser::Token* terminal = std::get_if<parser::Token>(&symbol)) {
        return terminal_trace_name(*terminal);
    }
    return std::string(parser::to_string(std::get<parser::Nonterminal>(symbol)));
}

// reduce_trace_payload renders a node as "lhs => rhs", the way the REDUCE line of a parser trace names the production
// it was reduced from, or as "lhs => ε" for a production with an empty right hand side.
std::string reduce_trace_payload(parser::Nonterminal lhs, const std::vector<parser::ParseNode>& rhs) {
    std::string payload = std::string(parser::to_string(lhs)).append(" =>");
    if (rhs.empty()) {
        return payload.append(" ε");
    }
    for (const parser::ParseNode& child : rhs) {
        payload.append(" ").append(symbol_trace_name(child.symbol));
    }
    return payload;
}

// append_tree_node appends the line of the given node and the lines of everything below it, which is the pre-order the
// tree trace is read in: a node, then what it was built from. The payload carries the indentation and the position and
// span columns do not, so they stay in the same place however deep a node sits.
void append_tree_node(std::vector<std::string>& lines, const parser::Scanner& scanner, const parser::ParseNode& node,
                      std::size_t depth) {
    const parser::Position position = scanner.position(node.byte_offset);
    const std::string location = pad_field(std::to_string(position.line) + ":" + std::to_string(position.column));
    const std::string span = pad_field(std::to_string(node.byte_offset) + "+" + std::to_string(node.byte_length));

    std::string payload(depth * 2, ' ');
    if (const parser::Token* terminal = std::get_if<parser::Token>(&node.symbol)) {
        if (*terminal == parser::Token::ErrorToken) {
            // The error node stands for no token of its own, so its span is all it carries.
            payload += terminal_trace_name(*terminal);
        } else {
            // The text is read off the source through the span and never carried along from the token, which is what
            // makes the trace state that the span is right.
            payload += terminal_trace_name(*terminal) + " \"" +
                       escape_lexeme(scanner.text(node.byte_offset, node.byte_length)) + "\"";
        }
    } else {
        payload += reduce_trace_payload(std::get<parser::Nonterminal>(node.symbol), node.children);
    }

    lines.push_back(location + " " + span + " " + payload);
    for (const parser::ParseNode& child : node.children) {
        append_tree_node(lines, scanner, child, depth + 1);
    }
}

// append_tree_trace parses the whole input and appends one line per node of the tree the parse built, in pre-order. A
// parse which was given up builds no tree and appends nothing, which is the empty trace the harness expects for it.
void append_tree_trace(std::vector<std::string>& lines, std::string_view source, const std::string& input_path) {
    // The scanner stays at hand after the parse, because a node carries the span of the source it covers and not the
    // source itself, so the trace resolves every node through position() and text().
    parser::Scanner scanner(source, input_path);
    parser::TokenSkipper skipper(scanner);

    parser::Parser instance;
    const parser::ParseResult result = instance.parse(skipper);
    if (result.tree.has_value()) {
        append_tree_node(lines, scanner, *result.tree, 0);
    }
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
    write_trace(TREE_TRACE_FILE_NAME, [&](std::vector<std::string>& lines) {
        append_tree_trace(lines, source, input_path);
    });
    return 0;
}
