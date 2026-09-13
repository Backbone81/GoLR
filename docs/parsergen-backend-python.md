# Parser Generator Backend: Python

This backend outputs a parser as Python source code. See [parser generator backends](parsergen-backend.md) for what
every generated parser does and [scanner generator backend: Python](scannergen-backend-python.md) for the scanner side.

It targets Python 3.10 and imports nothing beyond the standard library.

`--backend-python-scanner-module` sets the module the token constants are imported from, which defaults to `scanner`. A
leading dot makes the import relative, which is what a scanner and a parser sitting in the same package need.

`Parser.parse` takes a `TokenSource` and returns a `ParseResult` holding the tree and the errors, and can be called
again with another scanner. The tree is `None` when the parse could not be finished. Errors are returned rather than
raised, so reporting many of them costs nothing. Lexemes are `bytes` sliced out of the source.

Every nonterminal node's `production` field names the alternative it was reduced by, as one of the generated
`Production` members (`PRODUCTION_EXPRESSION1`, ... - `@name` in the grammar overrides the auto-generated name). It is
`None` on a terminal node.

## Tracing

`Parser` has a `trace` attribute of type `TraceFunc | None` (`Callable[[str], None] | None`). Set it after `Parser()`
for a trace of every parser action, `None` for none. See [parser generator backends](parsergen-backend.md) for the line
format.

## Example

[examples/calculator/python/](../examples/calculator/python/) is a calculator built on this backend. Its parser was
generated with:

```sh
golr parser \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend python \
  --backend-python-scanner-module .scanner \
  --backend-file-path parser/parser.py
```

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/backend/python
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkFromParser/GNU_Bison_3.8.2-32               771           2136063 ns/op          542901 B/op       9218 allocs/op
BenchmarkFromParser/GCC_2.95.3_C-32                  102          10953807 ns/op         2785500 B/op      31730 allocs/op
BenchmarkFromParser/GCC_2.95.3_Objective_C-32         62          16643639 ns/op         4599285 B/op      44064 allocs/op
BenchmarkFromParser/GCC_3.3.6_C++-32                  45          27918389 ns/op        11627543 B/op      86601 allocs/op
BenchmarkFromParser/GCC_4.2.4_Java-32                 82          12407483 ns/op         4649517 B/op      47749 allocs/op
BenchmarkFromParser/Go_1.5.4-32                      123           9915116 ns/op         2701164 B/op      30663 allocs/op
BenchmarkFromParser/PHP_8.6.7-32                      62          19431101 ns/op         8541590 B/op      64467 allocs/op
BenchmarkFromParser/PostgreSQL_18.4-32                 3         486181994 ns/op        138037496 B/op    509326 allocs/op
BenchmarkFromParser/Ruby_3.2.11-32                    66          22212942 ns/op         9312261 B/op      66405 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/backend/python    12.356s
```
