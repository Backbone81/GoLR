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
BenchmarkFromParser/GNU_Bison_3.8.2-32          553     2,137,336 ns/op       589,144 B/op     8,454 allocs/op
BenchmarkFromParser/GCC_2.95.3_C-32             100    11,374,236 ns/op     2,877,677 B/op    29,693 allocs/op
BenchmarkFromParser/GCC_2.95.3_Objective_C-32    79    13,731,932 ns/op     4,417,846 B/op    41,288 allocs/op
BenchmarkFromParser/GCC_3.3.6_C++-32             45    28,399,160 ns/op    11,262,482 B/op    81,515 allocs/op
BenchmarkFromParser/GCC_4.2.4_Java-32            76    16,142,749 ns/op     4,440,372 B/op    44,307 allocs/op
BenchmarkFromParser/Go_1.5.4-32                  81    13,501,300 ns/op     2,795,913 B/op    28,759 allocs/op
BenchmarkFromParser/PHP_8.6.7-32                 85    18,883,853 ns/op     8,717,857 B/op    60,500 allocs/op
BenchmarkFromParser/PostgreSQL_18.4-32            3   460,461,592 ns/op   139,612,834 B/op   490,375 allocs/op
BenchmarkFromParser/Ruby_3.2.11-32               54    21,939,080 ns/op     9,515,399 B/op    62,723 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/backend/typescript        12.108s
```
