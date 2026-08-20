# The Python runner for the backend test corpus. It reads an input file, runs the generated scanner and the generated
# parser over it, and writes the canonical scanner trace and parser trace the harness diffs against.
#
# This file has no dependencies, and it must not grow any. It runs in the official python image with no network, which
# is what proves that generated GoLR code needs nothing but the bare language.

import sys

from parser import NonterminalSymbol, Parser, nonterminal_to_string
from scanner import Scanner, Token, TokenSkipper, token_to_string

SCANNER_TRACE_FILE_NAME = "scanner.actual"
PARSER_TRACE_FILE_NAME = "parser.actual"

# The bytes a trace line carries as they are. Everything outside of it is escaped.
PRINTABLE_LOW = 0x20
PRINTABLE_HIGH = 0x7E

# The bytes which have a shorter escape than the hexadecimal one.
NAMED_ESCAPES = {
    0x5C: "\\\\",
    0x22: '\\"',
    0x0A: "\\n",
    0x0D: "\\r",
    0x09: "\\t",
}


def escape_lexeme(lexeme):
    """Escapes the bytes of a lexeme. The caller writes the quotes around the result.

    The lexeme is bytes and is escaped byte by byte, never character by character, so a multi byte UTF-8 sequence
    becomes one \\xHH escape per byte. Decoding it first would report character offsets and disagree with every other
    backend. The repr of a bytes object is no alternative: it escapes a different set and picks its own quotes.
    """
    result = []
    for value in lexeme:
        named = NAMED_ESCAPES.get(value)
        if named is not None:
            result.append(named)
        elif PRINTABLE_LOW <= value <= PRINTABLE_HIGH:
            result.append(chr(value))
        else:
            # The trace asks for a zero padded pair of lower case hexadecimal digits.
            result.append(f"\\x{value:02x}")
    return "".join(result)


def append_scanner_trace(lines, source, input_path):
    """Scans the whole input and appends one line per event."""
    # The plain Scanner and not the TokenSkipper: a skipped rule matched like any other, and the offsets of the tokens
    # around it are only checkable when it is in the trace.
    scanner = Scanner(source, input_path)

    while scanner.next():
        token = scanner.token()

        # For a failed match this is the start of the attempt, not the byte which could not be consumed.
        start = scanner.byte_offset()

        if token == Token.INVALID_TOKEN:
            lines.append(f"ERROR {start}")
            continue

        # The end comes from the length of the lexeme rather than from an offset the scanner reports, so that a lexeme
        # which disagrees with the offsets cannot pass unnoticed.
        lexeme = scanner.lexeme()
        lines.append(f'TOKEN {token_to_string(token)} {start} {start + len(lexeme)} "{escape_lexeme(lexeme)}"')

    # The offset the scanner reports after it ran out of input, not the length of the source. The two agree only when
    # the scanner consumed everything.
    lines.append(f"EOF {scanner.byte_offset()}")


def append_node_trace(lines, node):
    """Appends the post-order walk of the subtree, which is the shift and reduce sequence of the parse."""
    for child in node.children:
        append_node_trace(lines, child)

    if isinstance(node.symbol, NonterminalSymbol):
        lines.append(f"REDUCE {nonterminal_to_string(node.symbol.nonterminal)} {len(node.children)}")
        return

    token = node.symbol.token
    if token == Token.ERROR_TOKEN:
        # The leaf the recovery pushed where it resumed. It stands for the dropped input and names no token.
        lines.append("RESYNC")
        return
    lines.append(f"SHIFT {token_to_string(token)}")


def append_parser_trace(lines, source, input_path):
    """Parses the whole input and appends its errors, the walk of the tree and the accept event.

    The errors come first because a shift carries no offset to interleave them by. A parse which was given up returns
    no tree, so its trace is the errors alone and the missing accept event is what says the two outcomes apart.
    """
    # The TokenSkipper here, because a skipped rule never reaches the parser.
    result = Parser().parse(TokenSkipper(Scanner(source, input_path)))

    for error in result.errors:
        lines.append(f"ERROR {error.byte_offset}")

    if result.tree is None:
        return
    append_node_trace(lines, result.tree)
    lines.append("ACCEPT")


def write_trace(file_name, produce, source, input_path):
    """Produces one trace and writes it to its file.

    Nothing is caught. Whatever goes wrong ends the run with its traceback, and that includes a warning, because the
    interpreter runs with "-W error". This is a test: a run which went wrong has nothing worth salvaging, so no trace
    is written and the harness reports the case by the trace it is missing.
    """
    lines = []
    produce(lines, source, input_path)

    # Every line is terminated, and an empty trace is an empty file rather than a bare newline.
    with open(file_name, "w", encoding="utf-8", newline="\n") as file:
        for line in lines:
            file.write(line + "\n")


def main():
    """Runs both traces over the input file named on the command line."""
    input_path = sys.argv[1]

    # Opened in binary mode, because the generated scanner wants the bytes of the input and not text. Handing it a str
    # would be the classic mistake this harness exists to catch.
    with open(input_path, "rb") as file:
        source = file.read()

    write_trace(SCANNER_TRACE_FILE_NAME, append_scanner_trace, source, input_path)
    write_trace(PARSER_TRACE_FILE_NAME, append_parser_trace, source, input_path)


main()
