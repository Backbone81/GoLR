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

## Tracing

The `Parser` struct has `trace` and `trace_context` fields. Set `trace` to a `<Prefix>TraceFunc`
(`void (*)(void *context, const char *line)`) after `<prefix>_parser_init` for a trace of every parser action; leave it
`null` for none. See [parser generator backends](parsergen-backend.md) for the line format.

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

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/backend/c
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkFromParser/GNU_Bison_3.8.2-32               344           3599985 ns/op          825261 B/op      16574 allocs/op
BenchmarkFromParser/GCC_2.95.3_C-32                   92          14771900 ns/op         3312553 B/op      44254 allocs/op
BenchmarkFromParser/GCC_2.95.3_Objective_C-32         91          17980560 ns/op         4925801 B/op      59409 allocs/op
BenchmarkFromParser/GCC_3.3.6_C++-32                  34          34266507 ns/op        12126923 B/op     109336 allocs/op
BenchmarkFromParser/GCC_4.2.4_Java-32                 62          19235984 ns/op         5007819 B/op      63749 allocs/op
BenchmarkFromParser/Go_1.5.4-32                      100          12390941 ns/op         2950595 B/op      42776 allocs/op
BenchmarkFromParser/PHP_8.6.7-32                      87          23623036 ns/op         9467267 B/op      82761 allocs/op
BenchmarkFromParser/PostgreSQL_18.4-32                 3         469236588 ns/op        141876658 B/op    585185 allocs/op
BenchmarkFromParser/Ruby_3.2.11-32                    60          22309519 ns/op        10237888 B/op      86184 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/backend/c 13.606s
```
