# Parser Generator Backend: Rust

This backend outputs a parser as Rust source code. See [parser generator backends](parsergen-backend.md) for what every
generated parser does and [scanner generator backend: Rust](scannergen-backend-rust.md) for the scanner side.

The output is a module which needs nothing beyond the standard library, and puts no constraint on the edition of the
crate it is dropped into.

`--backend-rust-scanner-module` sets the module path the token type and the token source trait are taken from, which
defaults to `super::scanner`.

`Parser::parse` takes anything implementing `TokenSource` and returns a `ParseResult` holding the tree and the errors,
and can be called again with another scanner. The tree is `None` when the parse could not be finished. The result is
`#[must_use]`, since dropping it unlooked at drops every error the parse reported. Lexemes borrow from the source, so
the borrow checker enforces that the source outlives the tree.

## Example

[examples/calculator/rust/](../examples/calculator/rust/) is a calculator built on this backend. Its parser was
generated with:

```sh
golr parser \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend rust \
  --backend-file-path src/parser/parser.rs
```
