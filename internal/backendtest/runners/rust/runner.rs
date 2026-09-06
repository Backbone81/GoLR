// The Rust runner for the backend test corpus. It reads an input file, runs the generated scanner and the generated
// parser over it, and writes the canonical scanner trace and parser trace the harness diffs against.
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

use crate::parser::Parser;
use crate::scanner::{Scanner, Token, TokenSkipper, TokenSource};

const SCANNER_TRACE_FILE_NAME: &str = "scanner.actual";
const PARSER_TRACE_FILE_NAME: &str = "parser.actual";

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

// append_scanner_trace scans the whole input and appends one line per event.
fn append_scanner_trace(lines: &mut Vec<String>, source: &[u8], input_path: &str) {
    // The plain Scanner and not the TokenSkipper: a skipped rule matched like any other, and the offsets of the tokens
    // around it are only checkable when it is in the trace.
    let mut scanner = Scanner::new(source, input_path);

    while scanner.next() {
        // For a failed match this is the start of the attempt, not the byte which could not be consumed.
        let start = scanner.byte_offset();

        if scanner.token() == Token::InvalidToken {
            lines.push(format!("ERROR {start}"));
            continue;
        }

        // The end comes from the length of the lexeme rather than from an offset the scanner reports, so that a lexeme
        // which disagrees with the offsets cannot pass unnoticed.
        let lexeme = scanner.lexeme();
        let end = start + lexeme.len();
        lines.push(format!(
            "TOKEN {} {} {} \"{}\"",
            scanner.token(),
            start,
            end,
            escape_lexeme(lexeme),
        ));
    }

    // The offset the scanner reports after it ran out of input, not the length of the source. The two agree only when
    // the scanner consumed everything.
    lines.push(format!("EOF {}", scanner.byte_offset()));
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
