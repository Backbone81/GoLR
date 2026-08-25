# Scanner Generator Backend: Rust

This backend outputs a scanner as Rust source code. See [scanner generator backends](scannergen-backend.md) for what
every generated scanner does and [parser generator backend: Rust](parsergen-backend-rust.md) for the parser side.

The output is a module which needs nothing beyond the standard library, and puts no constraint on the edition of the
crate it is dropped into.

The scanner is `Scanner` and its tokens are the `Token` enum, with `is_skipped` alongside it. `TokenSkipper` wraps a
scanner to drop the tokens marked for skipping. Both implement the `TokenSource` trait, which is what a parser reads
its tokens through. A scanner borrows its source, and lexemes are a `&[u8]` into it, so the borrow checker enforces
that the source outlives what is taken from it.

## Example

[examples/calculator/rust/](../examples/calculator/rust/) is a calculator built on this backend. Its scanner was
generated with:

```sh
golr scanner \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend rust \
  --backend-file-path src/parser/scanner.rs
```
