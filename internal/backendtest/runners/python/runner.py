# The Python runner for the backend test corpus. It reads an input file, runs the generated scanner and the generated
# parser over it, and writes the canonical scanner trace and parser trace the harness diffs against.
#
# This file has no dependencies, and it must not grow any. It runs in the official python image with no network, which
# is what proves that generated GoLR code needs nothing but the bare language.

import sys

from parser import Parser
from scanner import Scanner, Token, TokenSkipper

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
    """Scans the whole input and appends one line per event.

    The position the token or the failed match starts at, a keyword, and for a token its rule and lexeme, for a
    failed match the bytes it could not match.
    """
    # The plain Scanner and not the TokenSkipper: a skipped rule matched like any other, and the position of the
    # tokens around it is only checkable when it is in the trace.
    scanner = Scanner(source, input_path)

    while scanner.next():
        location = f"{scanner.line}:{scanner.column}"
        lexeme = escape_lexeme(scanner.lexeme)

        if scanner.token == Token.INVALID_TOKEN:
            lines.append(f'{location:<7} {"ERROR":<7} "{lexeme}"')
            continue

        lines.append(f'{location:<7} {"TOKEN":<7} {scanner.token} "{lexeme}"')

    # The position after the scanner ran out of input, which is one past the last byte only when it consumed
    # everything.
    location = f"{scanner.line}:{scanner.column}"
    lines.append(f"{location:<7} EOF")


def append_parser_trace(lines, source, input_path):
    """Parses the whole input and appends the line the parser's trace hook emits for every action.

    The hook reports the error recovery steps too, which the returned tree does not, so the tree and the error are
    ignored.
    """
    parser = Parser()
    parser.trace = lines.append

    # The TokenSkipper here, because a skipped rule never reaches the parser.
    parser.parse(TokenSkipper(Scanner(source, input_path)))


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


if __name__ == "__main__":
    main()
