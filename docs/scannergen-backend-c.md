# Scanner Generator Backend: C

This backend outputs a scanner as C source code. See [scanner generator backends](scannergen-backend.md) for what every
generated scanner does and [parser generator backend: C](parsergen-backend-c.md) for the parser side.

It targets C99 and is a self contained header which includes nothing beyond the standard library. Include it wherever
the scanner is used, and define `<PREFIX>_SCANNER_IMPLEMENTATION` before including it in exactly one translation unit,
which is where the tables and the function bodies are emitted.

`--backend-c-prefix` sets the prefix every generated name carries, which defaults to `parser`. Types take it in
PascalCase and functions in snake_case, so a prefix of `calculator` yields `CalculatorScanner` and
`calculator_scanner_init`.

The scanner is `<Prefix>Scanner`, set up with `<prefix>_scanner_init` and advanced with `<prefix>_scanner_next`.
`<Prefix>TokenSkipper` wraps one to drop the tokens marked for skipping. Either becomes a `<Prefix>TokenSource` - a
context pointer plus the function pointers to read it with - through `<prefix>_scanner_as_token_source` and
`<prefix>_token_skipper_as_token_source`, which is what a parser is handed. Lexemes are a `<Prefix>StringView` pointing
into the source, not a copy of it and not null terminated.

`<prefix>_scanner_position` returns a `<Prefix>Position` and `<prefix>_scanner_text` a `<Prefix>StringView`. The line
table behind `<prefix>_scanner_position` is allocated on its first call and released by `<prefix>_scanner_free`, so
call that once the scanner is done with, or the table leaks. A reset keeps the allocation for the next source. If the
table cannot be allocated, `<prefix>_scanner_position` still answers correctly by counting line feeds on every call.

## Example

[examples/calculator/c/](../examples/calculator/c/) is a calculator built on this backend. Its scanner was generated
with:

```sh
golr scanner \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend c \
  --backend-c-prefix calculator \
  --backend-file-path parser/scanner.h
```
