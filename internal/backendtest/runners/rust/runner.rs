// The Rust runner for the backend test corpus. It reads an input file, runs the generated scanner and the generated
// parser over it, and writes the canonical scanner trace, parser trace and tree trace the harness diffs against.
//
// This file has no dependencies, and it must not grow any. It runs in the image with no network, which is what proves
// that generated GoLR code needs nothing but the bare language.
//
// It is the crate root, so rustc finds the two generated modules beside it and the generated parser reaches the
// generated scanner as its sibling.

mod parser;
mod scanner;

use std::cell::RefCell;
use std::panic::{catch_unwind, AssertUnwindSafe};
use std::rc::Rc;

use crate::parser::{Nonterminal, ParseNode, ParseSymbol, Parser};
use crate::scanner::{Scanner, Token, TokenSkipper, TokenSource};

const SCANNER_TRACE_FILE_NAME: &str = "scanner.actual";
const PARSER_TRACE_FILE_NAME: &str = "parser.actual";
const TREE_TRACE_FILE_NAME: &str = "tree.actual";

// The bytes a trace line carries as they are. Everything outside of it is escaped.
const PRINTABLE_LOW: u8 = 0x20;
const PRINTABLE_HIGH: u8 = 0x7e;

fn main() {
    let args: Vec<String> = std::env::args().collect();
    if args.len() != 2 {
        panic!("the runner takes the input file as its only argument");
    }
    let input_path = args[1].clone();

    // The bytes and not a String, because the generated scanner wants the bytes. Handing it a decoded string would be
    // the classic mistake this harness exists to catch.
    let source = std::fs::read(&input_path).expect("reading the input file");

    write_trace(SCANNER_TRACE_FILE_NAME, |lines| {
        append_scanner_trace(lines, &source, &input_path);
    });
    write_trace(PARSER_TRACE_FILE_NAME, |lines| {
        append_parser_trace(lines, &source, &input_path);
    });
    write_trace(TREE_TRACE_FILE_NAME, |lines| {
        append_tree_trace(lines, &source, &input_path);
    });
}

// escape_lexeme escapes the bytes of a lexeme. The caller writes the quotes around the result.
//
// The lexeme is escaped byte by byte, never character by character, so a multi byte UTF-8 sequence becomes one \xHH
// escape per byte. Decoding it first would report character offsets and disagree with every other backend.
fn escape_lexeme(lexeme: &[u8]) -> String {
    let mut result = String::new();
    for &value in lexeme {
        match value {
            b'\\' => result.push_str("\\\\"),
            b'"' => result.push_str("\\\""),
            b'\n' => result.push_str("\\n"),
            b'\r' => result.push_str("\\r"),
            b'\t' => result.push_str("\\t"),
            PRINTABLE_LOW..=PRINTABLE_HIGH => result.push(value as char),
            // The trace asks for a zero padded pair of lower case hex digits.
            _ => result.push_str(&format!("\\x{value:02x}")),
        }
    }
    result
}

// append_scanner_trace scans the whole input and appends one line per event: the position the token or the failed
// match starts at, a keyword, and for a token its rule and lexeme, for a failed match the bytes it could not match.
// check_position holds the offset based position and text against the line, column and lexeme of the token the scanner
// currently sits on. The two ways of asking exist side by side until the release which drops line and column, and the
// corpus is where they have to agree: every case of it is far more input than a hand written test covers.
fn check_position(scanner: &Scanner<'_>) {
    let position = scanner.position(scanner.byte_offset());
    if position.line != scanner.line()
        || position.column != scanner.column()
        || position.file_path != scanner.file_path()
    {
        eprintln!(
            "position({}) is {} {}:{}, but the scanner reports {} {}:{}",
            scanner.byte_offset(),
            position.file_path,
            position.line,
            position.column,
            scanner.file_path(),
            scanner.line(),
            scanner.column()
        );
        std::process::exit(1);
    }

    let lexeme = scanner.lexeme();
    if scanner.text(scanner.byte_offset(), lexeme.len()) != lexeme {
        eprintln!("text({}, {}) differs from the lexeme", scanner.byte_offset(), lexeme.len());
        std::process::exit(1);
    }
}

fn append_scanner_trace(lines: &mut Vec<String>, source: &[u8], input_path: &str) {
    // The plain Scanner and not the TokenSkipper: a skipped rule matched like any other, and the position of the
    // tokens around it is only checkable when it is in the trace.
    let mut scanner = Scanner::new(source, input_path);

    while scanner.next() {
        check_position(&scanner);

        let location = format!("{}:{}", scanner.line(), scanner.column());
        let lexeme = escape_lexeme(scanner.lexeme());

        if scanner.token() == Token::InvalidToken {
            lines.push(format!("{location:<7} {:<7} \"{lexeme}\"", "ERROR"));
            continue;
        }
        lines.push(format!("{location:<7} {:<7} {} \"{lexeme}\"", "TOKEN", scanner.token()));
    }

    // The position after the scanner ran out of input, which is one past the last byte only when it consumed
    // everything. It is the offset every off by one in a line table lands on, so it is checked like a token.
    check_position(&scanner);
    let location = format!("{}:{}", scanner.line(), scanner.column());
    lines.push(format!("{location:<7} {}", "EOF"));
}

// append_parser_trace parses the whole input and appends the line the parser's trace hook emits for every action. The
// hook reports the error recovery steps too, which the returned tree does not, so the tree and the error are ignored.
fn append_parser_trace(lines: &mut Vec<String>, source: &[u8], input_path: &str) {
    // The TokenSkipper here, because a skipped rule never reaches the parser.
    let mut scanner = TokenSkipper::new(Scanner::new(source, input_path));

    // The hook is 'static, so it collects through a shared cell rather than by borrowing the lines directly.
    let collected: Rc<RefCell<Vec<String>>> = Rc::new(RefCell::new(Vec::new()));
    let sink = collected.clone();

    let mut parser = Parser::new();
    parser.trace = Some(Box::new(move |line: &str| sink.borrow_mut().push(line.to_string())));
    let _ = parser.parse(&mut scanner);
    drop(parser);

    // The parser held the only other owner of the cell and is gone, so the lines can be moved straight out.
    let collected = Rc::into_inner(collected).expect("the trace hook was the only other owner");
    lines.extend(collected.into_inner());
}

// terminal_trace_name names a terminal for a trace line, giving the three tokens the grammar cannot spell a dollar
// name.
fn terminal_trace_name(terminal: Token) -> String {
    match terminal {
        Token::EndToken => "$end".to_string(),
        Token::ErrorToken => "$error".to_string(),
        Token::InvalidToken => "$invalid".to_string(),
        _ => terminal.to_string(),
    }
}

// symbol_trace_name is the bare grammar name of a symbol, which for a nonterminal is what its Display writes and for a
// terminal is the name the traces spell it with.
fn symbol_trace_name(symbol: ParseSymbol) -> String {
    match symbol {
        ParseSymbol::Nonterminal(nonterminal) => nonterminal.to_string(),
        ParseSymbol::Terminal(terminal) => terminal_trace_name(terminal),
    }
}

// reduce_trace_payload renders a node as "lhs => rhs", the way the REDUCE line of a parser trace names the production
// it was reduced from, or as "lhs => ε" for a production with an empty right hand side.
fn reduce_trace_payload(lhs: Nonterminal, rhs: &[ParseNode]) -> String {
    let mut payload = format!("{lhs} =>");
    if rhs.is_empty() {
        payload.push_str(" ε");
        return payload;
    }
    for child in rhs {
        payload.push(' ');
        payload.push_str(&symbol_trace_name(child.symbol));
    }
    payload
}

// append_tree_node appends the line of the given node and the lines of everything below it, which is the pre-order the
// tree trace is read in: a node, then what it was built from. The payload carries the indentation and the position and
// span columns do not, so they stay in the same place however deep a node sits.
fn append_tree_node(
    lines: &mut Vec<String>,
    scanner: &TokenSkipper<Scanner<'_>>,
    node: &ParseNode,
    depth: usize,
) {
    let position = scanner.position(node.byte_offset);
    let location = format!("{}:{}", position.line, position.column);
    let span = format!("{}+{}", node.byte_offset, node.byte_length);

    let mut payload = "  ".repeat(depth);
    match node.symbol {
        ParseSymbol::Nonterminal(nonterminal) => {
            payload.push_str(&reduce_trace_payload(nonterminal, &node.children));
        }
        // The error node stands for no token of its own, so its span is all it carries.
        ParseSymbol::Terminal(Token::ErrorToken) => {
            payload.push_str(&terminal_trace_name(Token::ErrorToken));
        }
        ParseSymbol::Terminal(terminal) => {
            // The text is read off the source through the span and never carried along from the token, which is what
            // makes the trace state that the span is right.
            let text = escape_lexeme(scanner.text(node.byte_offset, node.byte_length));
            payload.push_str(&format!("{} \"{text}\"", terminal_trace_name(terminal)));
        }
    }

    lines.push(format!("{location:<7} {span:<7} {payload}"));
    for child in &node.children {
        append_tree_node(lines, scanner, child, depth + 1);
    }
}

// append_tree_trace parses the whole input and appends one line per node of the tree the parse built, in pre-order. A
// parse which was given up builds no tree and appends nothing, which is the empty trace the harness expects for it.
fn append_tree_trace(lines: &mut Vec<String>, source: &[u8], input_path: &str) {
    // The scanner stays at hand after the parse, because a node carries the span of the source it covers and not the
    // source itself, so the trace resolves every node through position and text.
    let mut scanner = TokenSkipper::new(Scanner::new(source, input_path));

    let result = Parser::new().parse(&mut scanner);
    if let Some(tree) = result.tree {
        append_tree_node(lines, &scanner, &tree, 0);
    }
}

// write_trace produces one trace and writes it to its file. Whatever was produced before a panic is written all the
// same, so a runner which breaks half way still says how far it got, and the other trace is still produced.
fn write_trace<F: FnOnce(&mut Vec<String>)>(file_name: &str, produce: F) {
    let mut lines: Vec<String> = Vec::new();

    // The default panic hook has already reported the panic and its location on standard error, which is where the
    // reader of the container output looks for it.
    let outcome = catch_unwind(AssertUnwindSafe(|| produce(&mut lines)));
    if outcome.is_err() {
        eprintln!("producing {file_name} failed");
    }

    // Every line is terminated, and an empty trace is an empty file rather than a bare newline. The line ending is
    // spelled out, because the trace is LF whatever the platform would use.
    let mut text = String::new();
    for line in &lines {
        text.push_str(line);
        text.push('\n');
    }
    std::fs::write(file_name, text).expect("writing the trace file");
}
