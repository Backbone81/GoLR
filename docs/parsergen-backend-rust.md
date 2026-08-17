# Parser Generator Backend: Rust

This backend outputs a parser as Rust source code. The generated Rust code is a table driven parser which holds the
parsing decisions in lookup tables.

The parser uses the token type and the scanner trait of the generated scanner. Use `--backend-rust-scanner-module` to
set the module path, which defaults to `super::scanner`.
