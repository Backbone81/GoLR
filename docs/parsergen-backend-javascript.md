# Parser Generator Backend: JavaScript

This backend outputs a parser as JavaScript source code. See [parser generator backends](parsergen-backend.md) for what
every generated parser does and [scanner generator backend: JavaScript](scannergen-backend-javascript.md) for the
scanner side.

The output is an ES module and needs nothing beyond the language itself.

`--backend-javascript-scanner-module` sets the module specifier the token constants are imported from, which defaults
to `./scanner.js`.

`Parser.parse` takes a scanner and returns a result holding the tree and the errors, and can be called again with
another scanner. The tree is null when the parse could not be finished. Lexemes are a `Uint8Array` viewing into the
source rather than a copy of it.

Every nonterminal node's `production` field names the alternative it was reduced by, as one of the generated
`Production` values (`ProductionExpression1`, ... - `@name` in the grammar overrides the auto-generated name). It is
null on a terminal node.

## Tracing

`Parser` has a `trace` field, a `(line: string) => void` callback or null. Set it after `new Parser()` for a trace of
every parser action, `null` for none. See [parser generator backends](parsergen-backend.md) for the line format.

## Example

[examples/calculator/javascript/](../examples/calculator/javascript/) is a calculator built on this backend. Its parser
was generated with:

```sh
golr parser \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend javascript \
  --backend-file-path parser/parser.js
```

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/backend/javascript
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkFromParser/GNU_Bison_3.8.2-32               577           2256082 ns/op          589185 B/op       8455 allocs/op
BenchmarkFromParser/GCC_2.95.3_C-32                  100          11798187 ns/op         2877714 B/op      29694 allocs/op
BenchmarkFromParser/GCC_2.95.3_Objective_C-32        100          15603573 ns/op         4417709 B/op      41289 allocs/op
BenchmarkFromParser/GCC_3.3.6_C++-32                  42          28147108 ns/op        11264495 B/op      81516 allocs/op
BenchmarkFromParser/GCC_4.2.4_Java-32                100          15583191 ns/op         4440177 B/op      44309 allocs/op
BenchmarkFromParser/Go_1.5.4-32                      134           8789713 ns/op         2795912 B/op      28760 allocs/op
BenchmarkFromParser/PHP_8.6.7-32                      64          18788274 ns/op         8718679 B/op      60501 allocs/op
BenchmarkFromParser/PostgreSQL_18.4-32                 3         462877746 ns/op        139600541 B/op    490375 allocs/op
BenchmarkFromParser/Ruby_3.2.11-32                    55          20867320 ns/op         9515545 B/op      62724 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/backend/javascript        12.652s
```
