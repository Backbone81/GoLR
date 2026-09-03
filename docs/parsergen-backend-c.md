# Parser Generator Backend: C

This backend outputs a parser as C source code. See [parser generator backends](parsergen-backend.md) for what every
generated parser does and [scanner generator backend: C](scannergen-backend-c.md) for the scanner side.

It targets C99 and is a self contained header which carries its own implementation. Include it wherever the parser is
used, and define `<PREFIX>_PARSER_IMPLEMENTATION` before including it in exactly one translation unit, which is where
the tables and the function bodies are emitted.

`--backend-c-prefix` sets the prefix every generated name carries, which defaults to `parser` and has to be the one the
scanner was generated with. `--backend-c-scanner-include` sets the header the token type and the token source are
included from, which defaults to `scanner.h`.

`<prefix>_parser_parse` takes a `<Prefix>TokenSource` and returns a `<Prefix>ParseResult` holding the tree and the
errors. The result owns both, including the nodes of the tree, and is released with `<prefix>_parse_result_free`; a
result already returned is unaffected by later parses. The parser itself is set up with `<prefix>_parser_init` and
released with `<prefix>_parser_free`, and serves one source after another. Releasing it does not touch the results it
produced. An allocation which fails ends the parse and is reported as an error of kind
`<PREFIX>_ERROR_KIND_OUT_OF_MEMORY`.

Every nonterminal node's `production` field names the alternative it was reduced by, as one of the generated
`<Prefix>Production` enumerators (`<PREFIX>_PRODUCTION_EXPRESSION1`, ... - `@name` in the grammar overrides the
auto-generated name). It is `<PREFIX>_NO_PRODUCTION` on a terminal node, which no production reduces to.

## Example

[examples/calculator/c/](../examples/calculator/c/) is a calculator built on this backend. Its parser was generated
with:

```sh
golr parser \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend c \
  --backend-c-prefix calculator \
  --backend-file-path parser/parser.h
```
