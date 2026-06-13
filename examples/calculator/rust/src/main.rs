mod parser;

use parser::scanner::{Scanner, Token, TokenSkipper, TokenSkipperScanner};
use std::process;

fn main() {
    // We expect this binary to be used like "calculator '4 + 5 * 3'"
    let args: Vec<String> = std::env::args().collect();
    if args.len() != 2 {
        eprintln!("usage: calculator <expression>");
        process::exit(1);
    }

    if let Err(e) = tokenize(&args[1]) {
        eprintln!("{}", e);
        process::exit(1);
    }
}

fn tokenize(expression: &str) -> Result<(), String> {
    // The generated TokenSkipper will skip all whitespaces which the parser is not interested in.
    let mut scanner = TokenSkipper::new(
        // The generated Scanner will convert the input into tokens. The file_path argument is used in error messages.
        Scanner::new(expression.as_bytes().to_vec(), "expression"),
    );

    while scanner.next() {
        let token = scanner.token();
        if token == Token::InvalidToken {
            return Err(format!(
                "unexpected input at line {}, column {}",
                scanner.line(),
                scanner.column(),
            ));
        }
        let lexeme = String::from_utf8_lossy(scanner.lexeme());
        println!(
            "line {}, col {:<3}  {:<20}  {}",
            scanner.line(),
            scanner.column(),
            token,
            lexeme,
        );
    }

    Ok(())
}
