# Parser Generator Backend: TypeScript

This backend outputs a parser as TypeScript source code. See [parser generator backends](parsergen-backend.md) for what
every generated parser does and [scanner generator backend: TypeScript](scannergen-backend-typescript.md) for the
scanner side.

The output is an ES module and needs nothing beyond the language itself.

`--backend-typescript-scanner-module` sets the module specifier the token constants are imported from, which defaults
to `./scanner.js`.

`Parser.parse` takes a `TokenSource` and returns a `ParseResult` holding the tree and the errors, and can be called
again with another scanner. The tree is null when the parse could not be finished. Lexemes are a `Uint8Array` viewing
into the source rather than a copy of it.

Every nonterminal node's `production` field names the alternative it was reduced by, as one of the generated
`Production` values (`ProductionExpression1`, ... - `@name` in the grammar overrides the auto-generated name). It is
null on a terminal node.

## Tracing

`Parser` has a `trace` field of type `TraceFunc | null` (`(line: string) => void`). Set it after `new Parser()` for a
trace of every parser action, `null` for none. See [parser generator backends](parsergen-backend.md) for the line format.

## Example

[examples/calculator/typescript/](../examples/calculator/typescript/) is a calculator built on this backend. Its parser
was generated with:

```sh
golr parser \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend typescript \
  --backend-file-path parser/parser.ts
```

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/backend/typescript
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkFromParser/GNU_Bison_3.8.2-32               553           2137336 ns/op          589144 B/op       8454 allocs/op
BenchmarkFromParser/GCC_2.95.3_C-32                  100          11374236 ns/op         2877677 B/op      29693 allocs/op
BenchmarkFromParser/GCC_2.95.3_Objective_C-32         79          13731932 ns/op         4417846 B/op      41288 allocs/op
BenchmarkFromParser/GCC_3.3.6_C++-32                  45          28399160 ns/op        11262482 B/op      81515 allocs/op
BenchmarkFromParser/GCC_4.2.4_Java-32                 76          16142749 ns/op         4440372 B/op      44307 allocs/op
BenchmarkFromParser/Go_1.5.4-32                       81          13501300 ns/op         2795913 B/op      28759 allocs/op
BenchmarkFromParser/PHP_8.6.7-32                      85          18883853 ns/op         8717857 B/op      60500 allocs/op
BenchmarkFromParser/PostgreSQL_18.4-32                 3         460461592 ns/op        139612834 B/op    490375 allocs/op
BenchmarkFromParser/Ruby_3.2.11-32                    54          21939080 ns/op         9515399 B/op      62723 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/backend/typescript        12.108s
```
